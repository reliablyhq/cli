package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	iso8601 "github.com/ChannelMeter/iso8601duration"
)

// This has been built to mitigate the poor support in JSON and YAML unmarshalling for the time.Duration type
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	val := strconv.Quote(d.String())
	return []byte(val), nil
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
	return d.String(), nil
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

func (d Iso8601Duration) String() string {
	return d.Duration.String()
}

func (d Iso8601Duration) ToDuration() time.Duration {
	return d.Duration.ToDuration()
}

func (d Iso8601Duration) MarshalJSON() ([]byte, error) {
	val := strconv.Quote(d.String())
	return []byte(val), nil
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

// Help

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

// HumanizeDuration returns the time.Duration with better output format
// including the number of years, days (rather than very long hours)
// https://gist.github.com/harshavardhana/327e0577c4fed9211f65#gistcomment-2557682
// NB: Small adjusmtents were made to now show optional trailing 0d0s
// but not in the middle ie 1y0d3h will be kept as is
func HumanizeDurationShort(d time.Duration) string {
	if d < day {
		return d.String()
	}

	var b strings.Builder

	if d >= year {
		years := d / year
		fmt.Fprintf(&b, "%dy", years)
		d -= years * year
	}

	if d > 0 {
		days := d / day
		d -= days * day
		fmt.Fprintf(&b, "%dd", days)
	}

	if d > 0 {
		fmt.Fprintf(&b, "%s", d)
	}

	return b.String()
}

// HumanizeDuration returns the duration with more human friendly format
// https://gist.github.com/harshavardhana/327e0577c4fed9211f65#gistcomment-2366908
func HumanizeDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"day", days},
		{"hour", hours},
		{"minute", minutes},
		{"second", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		case 1:
			parts = append(parts, fmt.Sprintf("%d %s", chunk.amount, chunk.singularName))
		default:
			parts = append(parts, fmt.Sprintf("%d %ss", chunk.amount, chunk.singularName))
		}
	}

	return strings.Join(parts, " ")
}
