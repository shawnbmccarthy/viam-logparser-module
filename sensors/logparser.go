package sensors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/edaniels/golog"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/resource"
	"os"
	"strings"
	"sync"
	"time"
)

var LogParserModel = resource.NewModel("viam-soleng", "android", "logparser")

type LastMessage struct {
	lastRun    time.Time
	services   []string
	logs       []string
	searchFrom time.Time
	searchTo   time.Time
}

type LogParser struct {
	resource.Named
	mu              sync.Mutex
	logFileDirs     []string
	outputDirectory string
	logger          golog.Logger
	lastMessage     LastMessage
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
	logger golog.Logger,
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
	lp.lastMessage = LastMessage{}

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

	return nil
}

func (lp *LogParser) Readings(
	_ context.Context,
	_ map[string]interface{},
) (map[string]interface{}, error) {
	if lp.lastMessage.lastRun.IsZero() {
		return map[string]interface{}{"msg": "no searches have been run"}, nil
	}
	return toMap(lp)
}

func (lp *LogParser) Close(_ context.Context) error {
	return nil
}

func (lp *LogParser) DoCommand(_ context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	from, err := time.Parse("2006-01-02 13:01", strings.TrimSpace(cmd["from"].(string)))
	if err != nil {
		return nil,  fmt.Errorf("error parsing from time (format required: YYYY-MM-DD HH:MM) -> %w", err)
	}

	to, err := time.Parse("2006-01-02 13:01", strings.TrimSpace(cmd["to"].(string)))
	if err != nil {
		return nil, fmt.Errorf("error parsing to time (format required: YYY-MM-DD HH:MM) -> %w", err)
	}


	services := strings.Split(strings.TrimSpace(cmd["services"].(string)), ",")
	if len(services) == 0 {
		services = append(services, "*")
	}

	return doSearch(lp.logFileDirs, from, to, services)
}

func toMap(lp *LogParser) (map[string]interface{}, error) {
	var tmpMap map[string]interface{}
	d, err := json.Marshal(lp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(d, tmpMap)
	if err != nil {
		return nil, err
	}
	return tmpMap, nil

}
func doSearch(logDirs []string, from time.Time, to time.Time, services []string) (map[string]interface{}, error) {
	logsMoved := []string

	for _, dir := range logDirs {
		for _, service := range services {

		}
	}
}
