package core

import (
	"testing"
	"time"

	"encoding/json"
	iso8601 "github.com/ChannelMeter/iso8601duration"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestDurationMarshaling(t *testing.T) {

	type teststruct struct {
		Duration Duration `yaml:"duration" json:"duration"`
	}

	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "milliseconds",
			duration: 300 * time.Millisecond,
			want:     "300ms",
		},
		{
			name:     "seconds",
			duration: 30 * time.Second,
			want:     "30s",
		},
		{
			name:     "minutes",
			duration: 45 * time.Minute,
			want:     "45m",
		},
		{
			name:     "hours",
			duration: 72 * time.Hour,
			want:     "72h",
		},
		{
			name:     "full",
			duration: 72*time.Hour + 45*time.Minute + 30*time.Second + 123*time.Millisecond,
			want:     "72h45m30.123s",
		},
	}

	for _, tt := range tests {
		for _, format := range []string{"json"} {
			t.Run(tt.name+"-"+format, func(t *testing.T) {

				var marshaled string
				var unmarshaled teststruct

				var original teststruct = teststruct{
					Duration: Duration{Duration: tt.duration},
				}

				switch format {
				// marshal then unmarshal for both formats
				case "json":
					b, err := json.Marshal(original)
					if err != nil {
						t.Error("Unexpected marshaling original to json:", err)
					}
					marshaled = string(b)
					err = json.Unmarshal([]byte(marshaled), &unmarshaled)
					if err != nil {
						t.Error("Unexpected unmarshaling from json:", err)
					}
				case "yaml":
					b, err := yaml.Marshal(original)
					if err != nil {
						t.Error("Unexpected marshaling original to yaml:", err)
					}
					marshaled = string(b)
					err = yaml.Unmarshal([]byte(marshaled), &unmarshaled)
					if err != nil {
						t.Error("Unexpected unmarshaling from yaml:", err)
					}
				}

				t.Log(marshaled)
				t.Log(unmarshaled)

				// check the iso8601 duration has been converted to string
				// on marshaling & is well parsed on unmarshaling
				assert.Contains(t, marshaled, tt.want, "Marshaled duration is not found")
				assert.Equal(t, original, unmarshaled, "Unmarshaled content does not match original")
			})
		}

	}

}

func TestIso8601DurationMarshaling(t *testing.T) {

	type teststruct struct {
		Duration Iso8601Duration `yaml:"duration" json:"duration"`
	}

	tests := []struct {
		name     string
		duration iso8601.Duration
		want     string
	}{
		{
			name:     "weeks",
			duration: iso8601.Duration{Weeks: 4},
			want:     "P4W",
		},
		{
			name:     "full days & time",
			duration: iso8601.Duration{Days: 6, Hours: 23, Minutes: 59, Seconds: 59},
			want:     "P6DT23H59M59S",
		},
		{
			name:     "days only",
			duration: iso8601.Duration{Days: 30},
			want:     "P30D",
		},
		{
			name:     "time only",
			duration: iso8601.Duration{Hours: 23, Minutes: 59, Seconds: 59},
			want:     "PT23H59M59S",
		},
	}

	for _, tt := range tests {
		for _, format := range []string{"json"} {
			t.Run(tt.name+"-"+format, func(t *testing.T) {

				var marshaled string
				var unmarshaled teststruct

				var original teststruct = teststruct{
					Duration: Iso8601Duration{Duration: tt.duration},
				}

				switch format {
				// marshal then unmarshal for both formats
				case "json":
					b, err := json.Marshal(original)
					if err != nil {
						t.Error("Unexpected marshaling original to json:", err)
					}
					marshaled = string(b)
					err = json.Unmarshal([]byte(marshaled), &unmarshaled)
					if err != nil {
						t.Error("Unexpected unmarshaling from json:", err)
					}
				case "yaml":
					b, err := yaml.Marshal(original)
					if err != nil {
						t.Error("Unexpected marshaling original to yaml:", err)
					}
					marshaled = string(b)
					err = yaml.Unmarshal([]byte(marshaled), &unmarshaled)
					if err != nil {
						t.Error("Unexpected unmarshaling from yaml:", err)
					}
				}

				// check the iso8601 duration has been converted to string
				// on marshaling & is well parsed on unmarshaling
				assert.Contains(t, marshaled, tt.want, "Marshaled duration is not found")
				assert.Equal(t, original, unmarshaled, "Unmarshaled content does not match original")
			})
		}

	}
}

func TestHumanizeDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "full time",
			duration: time.Second*30 + time.Minute*5 + time.Hour*12,
			want:     "12 hours 5 minutes 30 seconds",
		},
		{
			name:     "time without minutes",
			duration: time.Second*30 + time.Hour*12,
			want:     "12 hours 30 seconds",
		},
		{
			name:     "time without seconds",
			duration: time.Second*30 + time.Hour*12,
			want:     "12 hours 30 seconds",
		},
		{
			name:     "full time with days",
			duration: time.Second*30 + time.Minute*5 + time.Hour*78,
			want:     "3 days 6 hours 5 minutes 30 seconds",
		},
		{
			name:     "days only",
			duration: time.Hour * 72,
			want:     "3 days",
		},
		{
			name:     "single units",
			duration: time.Second + time.Minute + time.Hour*25,
			want:     "1 day 1 hour 1 minute 1 second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := HumanizeDuration(tt.duration)
			assert.Equal(t, tt.want, str, "Unexpected duration humanized string")
		})
	}
}

func TestHumanizeDurationShort(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "full time",
			duration: time.Second*30 + time.Minute*5 + time.Hour*12,
			want:     "12h5m30s",
		},
		{
			name:     "hours with seconds no minutes",
			duration: time.Second*30 + time.Hour*12,
			want:     "12h0m30s",
		},
		{
			name:     "full time with days",
			duration: time.Second*30 + time.Minute*5 + time.Hour*78,
			want:     "3d6h5m30s",
		},
		{
			name:     "days only",
			duration: time.Hour * 72,
			want:     "3d",
		},
		{
			name:     "years",
			duration: time.Hour * 24 * 365 * 10,
			want:     "10y",
		},

		{
			name:     "years with time but no days",
			duration: time.Hour*24*365*10 + time.Hour*2 + time.Minute*5,
			want:     "10y0d2h5m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := HumanizeDurationShort(tt.duration)
			assert.Equal(t, tt.want, str, "Unexpected duration humanized string")
		})
	}
}
