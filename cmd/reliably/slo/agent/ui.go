package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/config"
	"github.com/reliablyhq/cli/core/agent"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/report"
	log "github.com/sirupsen/logrus"
)

const (
	agentView  = "agent"
	reportView = "report"
)

// NewUI create new agent ui
func runUI(client *api.Client, opts *AgentOptions, m *entities.Manifest, org string) error {
	done := make(chan struct{})
	var w sync.WaitGroup

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()
	g.SetManagerFunc(layout(opts))

	// set keybinding for ctrl_C
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitHandler(done)); err != nil {
		return err
	}

	logger := NewAgentLogger()
	agent.SetLogger(logger)

	// run agent
	go func() {
		job := agent.NewJob(opts.Interval, *m)
		job.ErrorFunc(func(e *agent.Error) {
			logger.Errorf(
				"error processing objective: %v\nerror: %s",
				e.Objective, e.Error())
		}).IndicatorFunc(func(i *entities.Indicator) error {
			return api.CreateEntity(client, config.Hostname, org, i)
		}).Do()
	}()

	// run agent output writer
	w.Add(1)
	go agentViewWriter(g, &w, logger, done)

	// run report writer
	w.Add(1)
	go reportViewWriter(g, &w, done, opts)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}

	w.Wait()
	return nil
}

// ui layout
func layout(opts *AgentOptions) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView(reportView, -1, 1, maxX, maxY-15); err != nil &&
			err != gocui.ErrUnknownView {
			return err
		} else {
			v.Wrap = true
			v.Title = " Report (3s) "
		}

		if v, err := g.SetView(agentView, -1, maxY/2, maxX, maxY); err != nil &&
			err != gocui.ErrUnknownView {
			return err
		} else {
			v.Wrap = true
			v.Title = fmt.Sprintf(" Agent (%ds) ", opts.Interval)
			v.Autoscroll = true
		}
		return nil
	}
}

func quitHandler(done chan struct{}) func(*gocui.Gui, *gocui.View) error {
	return func(_ *gocui.Gui, _ *gocui.View) error {
		close(done)
		return gocui.ErrQuit
	}
}

// write to agent view
func agentViewWriter(g *gocui.Gui, w *sync.WaitGroup, logger *AgentUILogger, done chan struct{}) {
	defer w.Done()

	updateFn := func(msg interface{}) func(g *gocui.Gui) error {
		return func(g *gocui.Gui) error {
			v, err := g.View(agentView)
			if err != nil {
				return err
			}
			// v.Clear()
			fmt.Fprintln(v, " ", msg)
			return nil
		}
	}

	for {
		select {
		case m := <-logger.InfoChan:
			g.Update(updateFn(m))

		case m := <-logger.WarnChan:
			g.Update(updateFn(color.YellowString("%s\n", m)))

		case m := <-logger.ErrorChan:
			g.Update(updateFn(color.RedString("%s\n", m)))

		case m := <-logger.DebugChan:
			g.Update(updateFn(color.MagentaString("%s\n", m)))

		case <-done:
			return
		}
	}
}

// write to report view
func reportViewWriter(g *gocui.Gui, w *sync.WaitGroup, done chan struct{}, opts *AgentOptions) {
	defer w.Done()
	rChan := make(chan []*report.Report)
	errChan := make(chan error, 1)

	updateFn := func(reports []*report.Report) func(g *gocui.Gui) error {
		return func(g *gocui.Gui) error {
			v, err := g.View(reportView)
			if err != nil {
				return err
			}
			v.Clear()
			report.Write(
				report.TABBED, reports[0], v, "",
				log.StandardLogger(), reports[1], report.EditReportSlice(reports))
			return nil
		}
	}

	go func() {
		for ch := time.Tick(time.Second * 3); ; <-ch {
			reports, err := report.GetReports(&report.ReportOptions{
				Selector:     opts.Selector,
				ManifestPath: opts.ManifestPath,
			})
			if err != nil {
				errChan <- err
			}
			rChan <- reports
		}
	}()

	for {
		select {
		case reports := <-rChan:
			g.Update(updateFn(reports))

		case e := <-errChan:
			g.Update(func(g *gocui.Gui) error {
				v, err := g.View(reportView)
				if err != nil {
					return err
				}
				v.Clear()
				fmt.Fprintln(v, color.RedString(e.Error()))
				return nil
			})

		case <-done:
			return

		}
	}
}
