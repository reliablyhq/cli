package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// This has been built to mitigate the poor support in JSON and YAML unmarshalling for the time.Duration type
type Duration struct {
	time.Duration
}

func (d Duration) MarhalJSON() ([]byte, error) {
	s := fmt.Sprintf("%dms", d.Milliseconds())
	return []byte(s), nil
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		{
			d.Duration = time.Duration(value)
		}
	case string:
		{
			if x, err := time.ParseDuration(value); err == nil {
				d.Duration = x
			} else {
				return err
			}
		}
	default:
		return errors.New("invalid duration")
	}
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return fmt.Sprintf("%dms", d.Milliseconds()), nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmp interface{}
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	switch value := tmp.(type) {
	case float64:
		{
			d.Duration = time.Duration(value)
		}
	case string:
		{
			t, err := time.ParseDuration(value)
			if err != nil {
				return err
			}

			d.Duration = t
		}
	default:
		return errors.New("invalid type")
	}

	return nil
}
