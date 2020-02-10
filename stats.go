package slack

import (
	"errors"
	"fmt"
	"os"

	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/ztrue/tracerr"
)

const (
	GB = float64(1024 * 1024 * 1024)
)

type Stats struct {
	thresholdMemory float64
	thresholdCPU    float64
	minutes         int
	report          *Report
	appName         string
	started         bool
}

type ConfigStats struct {
	Report          *Report
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

	hostname, err := os.Hostname()

	if err == nil {
		hostname = "@" + hostname
	}

	return &Stats{
		appName:         config.AppName + hostname,
		report:          config.Report,
		thresholdMemory: config.ThresholdMemory,
		thresholdCPU:    config.ThresholdCPU,
		minutes:         config.Minutes,
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

	markRedMem, markRedCpu := "", ""

	if float64(memory.Total-memory.Used)/(1024*1024) < s.thresholdMemory {
		markRedMem = Red
	}

	if float64(after.User-before.User)/total*100 > s.thresholdCPU {
		markRedCpu = Red
	}

	err = s.report.Stats(s.appName,
		fmt.Sprintf("Memory Total: %0.3f GB\n", float64(memory.Total)/GB),
		fmt.Sprintf("%sMemory Used: %0.3f GB\n", markRedMem, float64(memory.Used)/GB),
		fmt.Sprintf("Memory Cached:  %0.3f GB\n", float64(memory.Cached)/GB),
		fmt.Sprintf("Memory Free:  %0.3f GB\n", float64(memory.Free)/GB),
		fmt.Sprintf("%sCPU user: %0.2f %% \n", markRedCpu, float64(after.User-before.User)/total*100),
		fmt.Sprintf("CPU system: %0.2f %%\n", float64(after.System-before.System)/total*100),
		fmt.Sprintf("CPU idle: %0.2f %%\n", float64(after.Idle-before.Idle)/total*100),
	)
	if err != nil {
		return err
	}

	return nil
}
