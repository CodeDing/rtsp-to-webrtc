package webrtc

import (
	"encoding/json"
	"fmt"
)

// LogLevel is the logLevel parameter.
type LogLevel Level

// MarshalJSON implements json.Marshaler.
func (d LogLevel) MarshalJSON() ([]byte, error) {
	var out string

	switch d {
	case LogLevel(Error):
		out = "error"

	case LogLevel(Warn):
		out = "warn"

	case LogLevel(Info):
		out = "info"

	default:
		out = "debug"
	}

	return json.Marshal(out)
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *LogLevel) UnmarshalJSON(b []byte) error {
	var in string
	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}

	switch in {
	case "error":
		*d = LogLevel(Error)

	case "warn":
		*d = LogLevel(Warn)

	case "info":
		*d = LogLevel(Info)

	case "debug":
		*d = LogLevel(Debug)

	default:
		return fmt.Errorf("invalid log level: %s", in)
	}

	return nil
}

// unmarshalEnv implements envUnmarshaler.
func (d *LogLevel) unmarshalEnv(s string) error {
	return d.UnmarshalJSON([]byte(`"` + s + `"`))
}
