// Package agent contains code for pushing metric indicators
// to reliably entity api
package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/metrics"
)

type Labels entities.Labels

func (l Labels) String() string {
	var s []string
	for k, v := range l {
		s = append(s, fmt.Sprintf(`%s='%s'`, k, v))
	}

	return strings.Join(s, ", ")
}

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

	logger.Infof("indicator generated: %s", string(b))
	return nil
}

// ErrorHandler - type used to handle errors in Job
type ErrorHandler func(*Error)

func defaultErrorHandler(e *Error) {
	b, _ := json.MarshalIndent(e.Objective.Metadata.Labels, "", "  ")
	logger.Errorf("objective metadata: \n%s\nerror: %s", string(b), e.err)
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

	// errorHandler - error handle, executed against all errors
	// detected in the workflow
	errorHandler ErrorHandler

	done    chan ExitSignal
	results chan *result
}

// Do - run job
func (j *Job) Do() {
	logger.Info("--- starting slo indicator agent ---")
	// Ctrl+C listener
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Infof("\nCTRL+C pressed... exiting\n")
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
						logger.Infof("getting indicators for objective: [%s]", Labels(obj.Labels))
						indicator, err := getIndicatorFromObjective(obj)
						if err != nil {
							err = fmt.Errorf("error detected while getting indicator for objective: %v - %s", obj, err)
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

			logger.Infof("indicator percent: [%.2f] for objective: [%s]", r.Indicator.Spec.Percent, Labels(r.Objective.Labels))
			if err := j.indicatorHandler(r.Indicator); err != nil {
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

// NewJob - creates a new agent job
func NewJob(interval int64, objectives []*entities.Objective) *Job {
	return &Job{
		Interval:         interval,
		Objectives:       objectives,
		results:          make(chan *result, 50),
		done:             make(chan ExitSignal),
		indicatorHandler: defaultIndicatorHandler,
		errorHandler:     defaultErrorHandler,
	}
}

func getIndicatorFromObjective(obj *entities.Objective) (*entities.Indicator, error) {
	var (
		provider metrics.Provider
		err      error
	)

	now := time.Now()
	from := now.Add(time.Duration(-(int64(obj.Spec.Window.Duration))))
	to := now

	// loop over all providers to compute the objective indicate
	// we defer handling capability check to provider
	for _, f := range metrics.ProviderFactories {
		provider, err = f()
		if err != nil {
			logger.Debug(err)
			continue
		}

		if !provider.CanHandleSelector(obj.Spec.IndicatorSelector) {
			continue
		}

		// we stop/return on the first provider being able to compute the objective indicator
		i, err := provider.ComputeObjective(obj, from, to)
		if err != nil {
			logger.Debug(err)
		} else {
			// closing provider only when no errors
			provider.Close()
		}
		return i, err

	}

	return nil, errors.New("No provider was able to compute the objective indicator")

}
