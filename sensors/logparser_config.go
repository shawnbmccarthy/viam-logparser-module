package sensors

import "fmt"

type LogParserConfig struct {
	LogFiles        []string `json:"log_files"`
	OutputDirectory string   `json:"output_directory"`
}

func (cfg *LogParserConfig) Validate(path string) ([]string, error) {
	if len(cfg.LogFiles) == 0 {
		return nil, fmt.Errorf(`"log_files" attribute requires at least one log for parser %q`, path)
	}

	if cfg.OutputDirectory == "" {
		return nil, fmt.Errorf(`"output_directory" required for parser %q`, path)
	}

	return nil, nil
}
