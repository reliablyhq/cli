package gcp

import (
	"encoding/gob"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func TestCalculateLatencyOverThresholdPercentage(t *testing.T) {
	gob.Register(&monitoringpb.TypedValue_DoubleValue{})
	gob.Register(&monitoringpb.TypedValue_Int64Value{})
	latency, err := os.Open("../../../tests/providers/gcp/latency_1min.gob")
	if err != nil {
		t.Errorf("Could not open file: %v", err)
	}

	defer latency.Close()

	var decoded []*monitoringpb.TimeSeries
	if err := gob.NewDecoder(latency).Decode(&decoded); err != nil {
		panic(err)
	}

	threshold := 200

	latencyThresholdPercentage, err := calculateLatencyOverThresholdPercentage(threshold, decoded)
	if err != nil {
		t.Errorf("Could not parse latency: %v", err)
	}

	assert.Equal(t, 63.63636363636363, latencyThresholdPercentage, "Latency percentage incorrect")

}

func TestParseLatency(t *testing.T) {

	gob.Register(&monitoringpb.TypedValue_DoubleValue{})
	latencyFile, err := os.Open("../../../tests/providers/gcp/latency.gob")
	if err != nil {
		t.Errorf("Could not open file: %v", err)
	}

	defer latencyFile.Close()

	var decoded *monitoringpb.TimeSeries
	if err := gob.NewDecoder(latencyFile).Decode(&decoded); err != nil {
		t.Errorf("Could not decode file: %v", err)
	}

	latency, err := parseLatency([]*monitoringpb.TimeSeries{decoded})
	if err != nil {
		t.Errorf("Could not parse latency: %v", err)
	}

	assert.Equal(t, 238.84312225086497, latency, "Latency incorrect")
}

func TestParseStatusErrors(t *testing.T) {

	gob.Register(&monitoringpb.TypedValue_Int64Value{})
	statusFile, err := os.Open("../../../tests/providers/gcp/errorpercentage.gob")
	if err != nil {
		t.Errorf("Could not open file: %v", err)
	}

	defer statusFile.Close()

	var decoded []*monitoringpb.TimeSeries
	if err := gob.NewDecoder(statusFile).Decode(&decoded); err != nil {
		t.Errorf("Could not decode file: %v", err)
	}

	errorPercentage, err := parseStatusErrors(decoded)
	if err != nil {
		t.Errorf("Could not parse errors: %v", err)
	}

	assert.Equal(t, 0.0, errorPercentage, "Error Percentage Incorrect")
}
