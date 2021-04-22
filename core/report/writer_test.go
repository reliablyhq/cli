package report

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// This test ensure that any objective (success/failed/unknow) is
// well rendered for any writer (ie any outptu format)
func TestWritersOutputsContainSLOs(t *testing.T) {
	formats := []Format{SimpleText, TABBED, MARKDOWN}

	sloName := "This is a dummy SLO"
	now := time.Now()

	tests := []struct {
		name string
		slr  *ServiceLevelResult
	}{
		{
			name: "slo-unknown",
			slr:  nil,
		},
		{
			name: "slo-success",
			slr: &ServiceLevelResult{
				Actual:   100.0,
				Delta:    1.0,
				sloIsMet: true,
			},
		},
		{
			name: "slo-failed",
			slr: &ServiceLevelResult{
				Actual:   90.0,
				Delta:    -9.0,
				sloIsMet: false,
			},
		},
	}
	for _, tt := range tests {
		for _, format := range formats {
			testName := fmt.Sprintf("%s-%s", tt.name, format)
			t.Run(testName, func(t *testing.T) {

				r := &Report{
					APIVersion: "test/v1",
					Timestamp:  now,
					Services: []*Service{
						{
							Name: "test",
							ServiceLevels: []*ServiceLevel{
								{
									Name:      sloName,
									Type:      "dummy",
									Objective: 99,
									Result:    tt.slr,
									ObservationWindow: Window{
										From: now,
										To:   now,
									},
								},
							},
						},
					},
				}

				var buf bytes.Buffer
				Write(format, r, &buf, log.StandardLogger())
				//t.Log(buf.String())

				// assert the SLO line is renderded - by checking its name existence
				assert.Contains(t, buf.String(), sloName, "SLO has not been rendered in report format")
			})
		}

	}
}
