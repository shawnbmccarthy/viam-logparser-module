package sensors

import (
	"context"
	"errors"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var LogParserModel = resource.NewModel("viam-soleng", "android", "logparser")

type LastMessage struct {
	lastRun    time.Time // when did the last log run occur
	services   []string  // what services did we search for
	logs       []string  // what logs were found
	searchFrom time.Time // time of search (from)
	searchTo   time.Time // time of search (to)
}

type LogParser struct {
	resource.Named
	mu              sync.Mutex
	logFileDirs     []string // directories to search
	outputDirectory string   // where to copy files to
	logger          logging.Logger
	lastMessage     *LastMessage // last message (used by readings)
	timeZone        string       // system timezone
	offset          int          // system offset
}

func init() {
	resource.RegisterComponent(
		sensor.API,
		LogParserModel,
		resource.Registration[sensor.Sensor, *LogParserConfig]{
			Constructor: NewLogParser,
		},
	)
}

func NewLogParser(
	ctx context.Context,
	deps resource.Dependencies,
	conf resource.Config,
	logger logging.Logger,
) (sensor.Sensor, error) {
	lp := &LogParser{
		Named:  conf.ResourceName().AsNamed(),
		logger: logger,
	}

	if err := lp.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}

	return lp, nil
}

func (lp *LogParser) Reconfigure(
	_ context.Context,
	_ resource.Dependencies,
	conf resource.Config,
) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.logFileDirs = conf.Attributes.StringSlice("log_file_dirs")
	lp.outputDirectory = conf.Attributes.String("output_directory")
	lp.lastMessage = &LastMessage{}

	if len(lp.logFileDirs) == 0 {
		return errors.New("array of logfiles must be provided")
	}

	if len(strings.TrimSpace(lp.outputDirectory)) == 0 {
		return errors.New("an output directory must be specified")
	} else {
		if _, err := os.Stat(lp.outputDirectory); err != nil {
			if os.IsNotExist(err) {
				lp.logger.Errorf(`output directory does not exist: %q`, lp.outputDirectory)
				return err
			} else {
				lp.logger.Errorf(`check output directory permissions: %q`, lp.outputDirectory)
				return err
			}
		}
	}
	lp.timeZone, lp.offset = time.Now().Zone()
	return nil
}

func (lp *LogParser) Readings(_ context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
	if lp.lastMessage.lastRun.IsZero() {
		return map[string]interface{}{"msg": "no searches have been run"}, nil
	}

	return map[string]interface{}{
		"lastRun":  lp.lastMessage.lastRun.Local().Format(time.RFC3339),
		"from":     lp.lastMessage.searchFrom.Local().Format(time.RFC3339),
		"to":       lp.lastMessage.searchTo.Local().Format(time.RFC3339),
		"services": lp.lastMessage.services,
	}, nil
}

func (lp *LogParser) Close(_ context.Context) error {
	return nil
}

func (lp *LogParser) DoCommand(_ context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	start := time.Now()
	// build time
	from, err := buildRFC3339TimeString(strings.TrimSpace(cmd["from"].(string)), lp.offset)
	if err != nil {
		lp.logger.Errorf("error parsing from time, format required: YYYY-MM-DDTHH:MM, %v", err)
		return nil, err
	}

	to, err := buildRFC3339TimeString(strings.TrimSpace(cmd["to"].(string)), lp.offset)
	if err != nil {
		lp.logger.Errorf("error parsing to time, format required: YYYY-MM-DDTHH:MM, %v", err)
		return nil, err
	}

	var services []string
	val, ok := cmd["services"]
	if ok {
		services = strings.Split(strings.TrimSpace(val.(string)), ",")
	} else {
		services = append(services, "*")
	}

	files, err := doSearch(lp.logFileDirs, from, to, services, lp.logger)
	if err != nil {
		return nil, err
	}

	err = doCopy(files, lp.outputDirectory, lp.logger)
	if err != nil {
		return nil, err
	}

	lp.lastMessage = &LastMessage{
		lastRun:    start,
		services:   services,
		logs:       files,
		searchFrom: from,
		searchTo:   to,
	}

	return map[string]interface{}{
		"filesCopied": files,
		"dateFrom":    from.String(),
		"dateTo":      to.String(),
		"services":    services,
		"runtime":     time.Since(start).String(),
	}, nil
}

func buildRFC3339TimeString(date string, offset int) (time.Time, error) {
	// only accounts for whole numbers
	tz := offset / 60 / 60
	utz := absolute(tz)
	tzString := ""

	// negative add negative sign
	if tz < 0 {
		tzString = tzString + "-"
	} else {
		tzString = tzString + "+"
	}

	// less than 10?
	if utz < 10 {
		tzString = tzString + "0" + strconv.Itoa(utz) + ":00"
	} else {
		tzString = tzString + strconv.Itoa(utz) + ":00"
	}

	return time.Parse(time.RFC3339, date+tzString)
}

func absolute(x int) int {
	if x < 0 {
		return -1 * x
	}
	return x
}

func doSearch(logDirs []string, from, to time.Time, services []string, logger logging.Logger) ([]string, error) {
	var files []string
	logger.Debugf("searching -> from: %s, to: %s", from.Format(time.RFC3339), to.Format(time.RFC3339))

	// for each search directory
	for _, dir := range logDirs {
		// get full path
		path, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		if err := filepath.Walk(dir, func(pathItem string, pathInfo os.FileInfo, err error) error {
			if err != nil {
				logger.Errorf("error processing directory: %v", err)
				return err
			}
			// is this a file between given times and starts with a service name
			if !pathInfo.IsDir() && pathInfo.ModTime().Local().After(from.Local()) && pathInfo.ModTime().Local().Before(to.Local()) {
				fileToAppend := filepath.Join(path, filepath.Base(pathItem))
				// is the file in the data range part of a service we want to collect?
				for _, service := range services {
					if service == "*" {
						files = append(files, fileToAppend)
					} else {
						if strings.HasPrefix(pathItem, service) {
							files = append(files, fileToAppend)
						}
					}
				}
			}
			return nil
		}); err != nil {
			logger.Errorf("error walking directory: %v", err)
			return nil, err
		}
	}
	return files, nil
}

func doCopy(files []string, uploadDirectory string, logger logging.Logger) error {
	for _, file := range files {
		logger.Debugf("attempting copy of file: %s", file)
		dir, fn := filepath.Split(file)
		uploadTo := filepath.Join(uploadDirectory, dir)
		if err := os.MkdirAll(uploadTo, 0775); err != nil {
			return err
		}
		source, err := os.Open(file)
		if err != nil {
			logger.Errorf("error opening file(%s), %v", file, err)
			return err
		}

		dest, err := os.Create(filepath.Join(uploadTo, fn))
		if err != nil {
			logger.Errorf("failed to create: %s", filepath.Join(uploadTo, fn))
			return err
		}

		bytes, err := io.Copy(dest, source)
		if err != nil {
			logger.Errorf("failed to copy: from->%v to->%v", dest, source)
			return err
		}

		err = source.Close()
		if err != nil {
			logger.Warnf("failed to close source: %v", err)
		}
		err = dest.Close()
		if err != nil {
			logger.Warnf("failed to close dest: %v", err)
		}

		logger.Debugf("copied %d bytes to %s", bytes, dest.Name())
	}
	return nil
}
