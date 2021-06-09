// Package agent contains code for pushing metric indicators
// to reliably entity api
package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/reliablyhq/cli/core/metrics"
	"github.com/reliablyhq/entity-server/server/types"
	"github.com/reliablyhq/entity-server/server/types/v1/service_level"
)

type JobObjective struct {
	service_level.Objective
	ResourceID string
}

// Error - agent error type used to associate failed objectives
// with its given error
type Error struct {
	Objective *JobObjective
	err       error
}

// Error - impl error interface
func (e *Error) Error() string {
	return e.err.Error()
}

// IndicatorHandler - type used to handle indicators after
// they are generated
type IndicatorHandler func(*service_level.Indicator) error

func defaultIndicatorHandler(s *service_level.Indicator) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("indicator generated: %s", string(b))
	return nil
}

// ErrorHandler - type used to handle errors in Job
type ErrorHandler func(*Error)

func defaultErrorHandler(e *Error) {
	b, _ := json.MarshalIndent(e.Objective.MetadataSpec.Labels, "", "  ")
	log.Printf("objective metadata: \n%s\nerror: %s", string(b), e.err)
}

// ExitSignal - empty struct type used to
// trigger agent Job close
type ExitSignal struct{}

type result struct {
	Objective *JobObjective
	Indicator *service_level.Indicator
}

// Job - an Agent job defines the objectives and handlers
// for generating indicators.
type Job struct {
	// The number of seconds between indicator calculations and pushing
	Interval int64

	// Objectives - The objectives to create indicators from
	Objectives []*JobObjective

	// Cloud provider to query metrics for.
	// The current assumption is that only one cloud provider will
	// be supported per agent job
	Provider metrics.ProviderType

	// Handler - executes a given function againsts all
	// indicators generated.
	IndicatorHandler IndicatorHandler

	// Error - error handle, executed against all errors
	// detected in the workflow
	ErrorHandler ErrorHandler

	// ExitChan - the given channel is used to exit the workflow.
	// That is, any ExitSignal received to this chan will kill/stop the agent job
	done    chan ExitSignal
	results chan *result
}

// Do - run job
func (j *Job) Do() {
	// Ctrl+C listener
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("\nCTRL+C pressed... exiting\n")
		j.done <- ExitSignal{}
	}()

	go func() {
		for ch := time.Tick(time.Second * time.Duration(j.Interval)); ; <-ch {
			jobs := make(chan *JobObjective, 50)
			var w sync.WaitGroup

			// 1. populate jobs channel
			w.Add(1)
			go func() {
				defer w.Done()
				for _, obj := range j.Objectives {
					jobs <- obj
				}
				close(jobs)
			}()

			// 2. start works (go-routines) to process objectives in jobs channel
			// note: limiting the number for concurrent jobs to 5
			for i := 0; i < 5; i++ {
				w.Add(1)
				go func() {
					defer w.Done()
					for obj := range jobs {
						indicator, err := GetIndicatorFromObjective(j.Provider, obj)
						if err != nil {
							j.ErrorHandler(&Error{
								Objective: obj,
								err:       err,
							})
							continue
						}

						j.results <- &result{
							Objective: obj,
							Indicator: indicator,
						}
					}
				}()
			}

			w.Wait()
		}
	}()

	for {
		select {
		case <-j.done:
			return

		case r := <-j.results:
			if err := j.IndicatorHandler(r.Indicator); err != nil {
				j.ErrorHandler(&Error{
					Objective: r.Objective,
					err:       err,
				})
			}
		}
	}

}

// ErrorFunc - set the job ErrorHandler value
func (j *Job) ErrorFunc(f ErrorHandler) *Job {
	j.ErrorHandler = f
	return j
}

// IndicatorFunc - set the job IndicatorHandler value
func (j *Job) IndicatorFunc(f IndicatorHandler) *Job {
	j.IndicatorHandler = f
	return j
}

// NewJob - creates a new agent job
func NewJob(interval int64, objectives []*JobObjective, provider metrics.ProviderType) *Job {
	return &Job{
		Interval:         interval,
		Objectives:       objectives,
		Provider:         provider,
		results:          make(chan *result, 50),
		done:             make(chan ExitSignal),
		IndicatorHandler: defaultIndicatorHandler,
		ErrorHandler:     defaultErrorHandler,
	}
}

func GetIndicatorFromObjective(providerType metrics.ProviderType, obj *JobObjective) (*service_level.Indicator, error) {
	provider, err := metrics.ProviderFactories[providerType]()
	if err != nil {
		return nil, err
	}

	defer provider.Close()

	if err := obj.IsValid(); err != nil {
		return nil, err
	}

	var indicator service_level.Indicator
	// get from/to window from (now - given_duration), to now
	now := time.Now()
	indicator.Spec.From = now.Add(time.Duration(-(int64(obj.Spec.Window.Duration))))
	indicator.Spec.To = now

	// target := obj.Spec.IndicatorSelector["latency_target"]
	switch obj.Spec.IndicatorSelector["category"] {
	case "latency":
		indicator.Spec.Percent, err = provider.Get99PercentLatencyMetricForResource(obj.ResourceID,
			indicator.Spec.From, indicator.Spec.To)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported indicator category: %s",
			obj.Spec.IndicatorSelector["category"])
	}

	indicator.MetadataSpec.Labels = types.Labels(obj.Spec.IndicatorSelector)
	return &indicator, nil
}
