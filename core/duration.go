package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	iso8601 "github.com/ChannelMeter/iso8601duration"
)

// This has been built to mitigate the poor support in JSON and YAML unmarshalling for the time.Duration type
type Duration struct {
	time.Duration
}

func (d Duration) MarhalJSON() ([]byte, error) {
	s := fmt.Sprintf("%s", d.String())
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
	return fmt.Sprintf("%s", d.String()), nil
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

// Another JSON & YAML support for iso8601duration strings
type Iso8601Duration struct {
	iso8601.Duration
}

func (d *Iso8601Duration) String() string {
	return d.Duration.String()
}

func (d *Iso8601Duration) ToDuration() time.Duration {
	return d.Duration.ToDuration()
}

func (d Iso8601Duration) MarhalJSON() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Iso8601Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		{
			if t, err := iso8601.FromString(strings.ToUpper(value)); err == nil {
				d.Duration = *t
			} else {
				return err
			}
		}
	default:
		return errors.New("invalid ISO8601 duration")
	}
	return nil
}

func (d Iso8601Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

func (d *Iso8601Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmp interface{}
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	switch value := tmp.(type) {
	case string:
		{
			if t, err := iso8601.FromString(strings.ToUpper(value)); err == nil {
				d.Duration = *t
			} else {
				return err
			}
		}
	default:
		return errors.New("invalid ISO8601 duration")
	}

	return nil
}
