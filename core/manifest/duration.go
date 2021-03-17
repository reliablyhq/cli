package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const durationSuffix = "ms"

type Duration struct {
	time.Duration
}

func (d Duration) MarhalJSON() ([]byte, error) {
	return marshal(d)
}

func (d Duration) MarshalYAML() ([]byte, error) {
	return marshal(d)
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

func marshal(d Duration) ([]byte, error) {
	s := fmt.Sprintf("%v%s", d.Milliseconds(), durationSuffix)
	return []byte(s), nil
}
