package slack

import (
	"errors"
	"fmt"

	"github.com/BTiberiu99/slack"

	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/ztrue/tracerr"
)

type Stats struct {
	thresholdMemory float64
	thresholdCPU    float64
	minutes         int
	report          *report
	appName         string
	started         bool
}

type ConfigStats struct {
	Report          *report
	Minutes         int
	ThresholdMemory float64
	ThresholdCPU    float64
	AppName         string
}

func NewStats(config *ConfigStats) (*Stats, error) {
	if config.Report == nil {
		return nil, errors.New("Report must not be nil!")
	}
	if config.AppName == "" {
		config.AppName = "Default"
	}
	return &Stats{
		appName:         config.AppName,
		report:          config.Report,
		thresholdMemory: config.ThresholdMemory,
		thresholdCPU:    config.ThresholdCPU,
	}, nil
}

func (s *Stats) Start() {

	if s.started {
		return
	}
	s.started = true
	go func() {

		for {

			time.Sleep(time.Minute * time.Duration(s.minutes))

			if err := s.sendStats(); err != nil {
				s.report.Error(err)
			}

		}

	}()
}

func (s *Stats) sendStats() error {

	memory, err := memory.Get()

	if err != nil {
		return tracerr.Wrap(err)
	}

	before, err := cpu.Get()
	if err != nil {
		return tracerr.Wrap(err)
	}

	time.Sleep(time.Duration(1) * time.Second)
	after, err := cpu.Get()

	if err != nil {
		return tracerr.Wrap(err)
	}

	total := float64(after.Total - before.Total)

	div := float64(1024 * 1024 * 1024)

	markRedMem := ""

	if float64(memory.Total-memory.Used)/(1024*1024) < s.thresholdMemory {
		markRedMem = slack.Red
	}

	markRedCpu := ""

	if float64(after.User-before.User)/total*100 > s.thresholdCPU {
		markRedCpu = slack.Red
	}

	s.report.Stats(s.appName,
		fmt.Sprintf("Memory Total: %0.3f GB\n", float64(memory.Total)/div),
		fmt.Sprintf("%sMemory Used: %0.3f GB\n", markRedMem, float64(memory.Used)/div),
		fmt.Sprintf("Memory Cached:  %0.3f GB\n", float64(memory.Cached)/div),
		fmt.Sprintf("Memory Free:  %0.3f GB\n", float64(memory.Free)/div),
		fmt.Sprintf("%sCPU user: %0.2f %% \n", markRedCpu, float64(after.User-before.User)/total*100),
		fmt.Sprintf("CPU system: %0.2f %%\n", float64(after.System-before.System)/total*100),
		fmt.Sprintf("CPU idle: %0.2f %%\n", float64(after.Idle-before.Idle)/total*100),
	)

	return nil
}
