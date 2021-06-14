// Package agent contains code for pushing metric indicators
// to reliably entity api
package agent

import (
	"encoding/json"
	"fmt"

	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/metrics"
	log "github.com/sirupsen/logrus"
)

// Error - agent error type used to associate failed objectives
// with its given error
type Error struct {
	Objective *entities.Objective
	err       error
}

// Error - impl error interface
func (e *Error) Error() string {
	return e.err.Error()
}

// IndicatorHandler - type used to handle indicators after
// they are generated
type IndicatorHandler func(*entities.Indicator) error

func defaultIndicatorHandler(s *entities.Indicator) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("indicator generated: %s", string(b))
	return nil
}

// ErrorHandler - type used to handle errors in Job
type ErrorHandler func(*Error)

func defaultErrorHandler(e *Error) {
	b, _ := json.MarshalIndent(e.Objective.Metadata.Labels, "", "  ")
	fmt.Printf("objective metadata: \n%s\nerror: %s", string(b), e.err)
}

// ExitSignal - empty struct type used to
// trigger agent Job close
type ExitSignal struct{}

type result struct {
	Objective *entities.Objective
	Indicator *entities.Indicator
}

// Job - an Agent job defines the objectives and handlers
// for generating indicators.
type Job struct {
	// The number of seconds between indicator calculations and pushing
	Interval int64

	// Objectives - The objectives to create indicators from
	Objectives []*entities.Objective

	// indicatorHandler - executes a given function againsts all
	// indicators generated.
	indicatorHandler IndicatorHandler

	log Logger

	// errorHandler - error handle, executed against all errors
	// detected in the workflow
	errorHandler ErrorHandler

	done    chan ExitSignal
	results chan *result
}

// Do - run job
func (j *Job) Do() {
	j.log.Info("starting agent workflow")
	// Ctrl+C listener
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		j.log.Infof("\nCTRL+C pressed... exiting\n")
		j.done <- ExitSignal{}
	}()

	go func() {
		for ch := time.Tick(time.Second * time.Duration(j.Interval)); ; <-ch {
			jobs := make(chan *entities.Objective, 50)
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
						indicator, err := getIndicatorFromObjective(obj)
						if err != nil {
							j.log.Debugf("error detected while getting indicator for objective: %s - %s", obj, err)
							j.errorHandler(&Error{
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
			if err := j.indicatorHandler(r.Indicator); err != nil {
				j.log.Debugf("indicator handler failed: %s", err)
				j.errorHandler(&Error{
					Objective: r.Objective,
					err:       err,
				})
			}
		}
	}

}

// ErrorFunc - set the job ErrorHandler value
func (j *Job) ErrorFunc(f ErrorHandler) *Job {
	j.errorHandler = f
	return j
}

// IndicatorFunc - set the job IndicatorHandler value
func (j *Job) IndicatorFunc(f IndicatorHandler) *Job {
	j.indicatorHandler = f
	return j
}

// Logger - set Logger for this Job
func (j *Job) Logger(l Logger) {
	j.log = l
}

// NewJob - creates a new agent job
func NewJob(interval int64, objectives []*entities.Objective) *Job {
	return &Job{
		Interval:         interval,
		Objectives:       objectives,
		log:              log.StandardLogger(),
		results:          make(chan *result, 50),
		done:             make(chan ExitSignal),
		indicatorHandler: defaultIndicatorHandler,
		errorHandler:     defaultErrorHandler,
	}
}

func getIndicatorFromObjective(obj *entities.Objective) (*entities.Indicator, error) {
	var (
		provider   metrics.Provider
		resourceID string
		err        error
	)

	// identify provider
	for _, f := range metrics.ProviderFactories {
		provider, err = f()
		if err != nil {
			return nil, err
		}

		// check for resource ID
		if resourceID = provider.ResourceFromSelector(obj.Spec.IndicatorSelector); resourceID != "" {
			break
		}
	}

	// if resourceID is still undefined, error
	if resourceID == "" {
		return nil, fmt.Errorf("unable to identify provider and resource id for objective: %v",
			obj.Spec.IndicatorSelector)
	}

	defer provider.Close()

	var indicator entities.Indicator
	// get from/to window from (now - given_duration), to now
	now := time.Now()
	indicator.Spec.From = now.Add(time.Duration(-(int64(obj.Spec.Window.Duration))))
	indicator.Spec.To = now

	var f func(resourceID string, from time.Time, to time.Time) (float64, error)
	switch obj.Spec.IndicatorSelector["category"] {
	case "latency":
		f = provider.Get99PercentLatencyMetricForResource

	case "availability":
		f = provider.GetAvailabilityPercentage

	default:
		return nil, fmt.Errorf("unsupported indicator category: %s",
			obj.Spec.IndicatorSelector["category"])
	}

	indicator.Spec.Percent, err = f(resourceID, indicator.Spec.From, indicator.Spec.To)
	if err != nil {
		return nil, err
	}

	indicator.Metadata.Labels = obj.Spec.IndicatorSelector
	indicator.TypeMeta.Kind = "indicator"
	indicator.TypeMeta.APIVersion = obj.Version()
	return &indicator, nil
}
