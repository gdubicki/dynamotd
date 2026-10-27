package dynamotd

import (
	"bytes"
	"fmt"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// NoSettings is the settings type of the row types that don't have any.
type NoSettings struct{}

// Validator is implemented by the settings that need to be validated.
type Validator interface {
	Validate() error
}

// decodeSettings returns the defaults overridden with the given settings from the config file.
func decodeSettings[S any](defaults S, settingsNode *yaml.Node) (S, error) {
	settings := defaults
	if settingsNode != nil {
		// re-encode to be able to use the decoder that reports unknown keys (typos)
		data, err := yaml.Marshal(settingsNode)
		if err != nil {
			return settings, err
		}
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		if err := decoder.Decode(&settings); err != nil {
			return settings, err
		}
	}
	if validator, ok := any(settings).(Validator); ok {
		if err := validator.Validate(); err != nil {
			return settings, err
		}
	}
	return settings, nil
}

// Thresholds define when a value is shown as a warning or as critical.
type Thresholds struct {
	Warning  float64 `yaml:"warning"`
	Critical float64 `yaml:"critical"`
}

func (thresholds Thresholds) Validate() error {
	if thresholds.Warning > thresholds.Critical {
		return fmt.Errorf("warning threshold (%v) is higher than the critical one (%v)", thresholds.Warning, thresholds.Critical)
	}
	return nil
}

// Color returns the ok, warning or critical color depending on which threshold the value reached.
func (thresholds Thresholds) Color(value float64) color.Attribute {
	switch {
	case value >= thresholds.Critical:
		return ValueCriticalColor
	case value >= thresholds.Warning:
		return ValueWarningColor
	default:
		return ValueOkColor
	}
}
