package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Duration that can be JSON and YAML marshaled
// (JSON marshaller copied from https://stackoverflow.com/a/48051946/2361497)
type Duration struct {
	time.Duration
}

// MarshalYAML ...
func (d *Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

// UnmarshalYAML ...
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return fmt.Errorf("failed unmarshal to string")
	}

	var err error
	d.Duration, err = time.ParseDuration(s)
	return err
}

// MarshalJSON ...
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON ...
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
