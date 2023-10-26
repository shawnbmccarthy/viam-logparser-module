package sensors

import (
	"context"
	"errors"
	"github.com/edaniels/golog"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/resource"
	"os"
	"strings"
	"sync"
)

var LogParserModel = resource.NewModel("viam-soleng", "android", "logparser")

type LogParser struct {
	resource.Named
	mu              sync.Mutex
	logFiles        []string
	outputDirectory string
	logger          golog.Logger
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

	lp.logFiles = conf.Attributes.StringSlice("log_files")
	lp.outputDirectory = conf.Attributes.String("output_directory")

	if len(lp.logFiles) == 0 {
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
	return map[string]interface{}{"dummy_val": 0.0}, nil
}

func (lp *LogParser) Close(_ context.Context) error {
	return nil
}
