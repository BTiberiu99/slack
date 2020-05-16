package slack

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/ztrue/tracerr"
)

const (
	Kb = float64(1024)
	Mb = Kb * 1024
	Gb = Mb * 1024
)

var (
	errNoReport = errors.New("Report must not be nil!")
)

type Stats struct {
	appName         string
	report          *Report
	minutes         int
	thresholdMemory float64
	thresholdCPU    float64

	stopGorutine      func()
	onlyOverThreshold bool
	once              sync.Once
	block             chan int
}

type ConfigStats struct {

	//Application Name
	AppName string

	//Report
	Report *Report

	//Report time
	Minutes int

	//Thresholds
	ThresholdMemory float64
	ThresholdCPU    float64

	OnlyOverThreshold bool
}

func NewStats(config *ConfigStats) (*Stats, error) {
	if config.Report == nil {
		return nil, errNoReport
	}

	if config.AppName == "" {
		config.AppName = "Default"
	}

	hostname, err := os.Hostname()

	if err == nil {
		hostname = "@" + hostname
	}

	return &Stats{
		appName:           config.AppName + hostname,
		report:            config.Report,
		thresholdMemory:   config.ThresholdMemory,
		thresholdCPU:      config.ThresholdCPU,
		minutes:           config.Minutes,
		onlyOverThreshold: config.OnlyOverThreshold,
		block:             make(chan int),
	}, nil
}

func (s *Stats) Start() {

	s.once.Do(func() {

		loop := true

		s.stopGorutine = func() {
			loop = false
		}
		go func() {

			for {

				time.Sleep(time.Minute * time.Duration(s.minutes))

				if !loop {
					break
				}
				if err := s.sendStats(); err != nil {
					s.report.Error(err)
				}

			}

			s.block <- 1

		}()
	})
}

func (s *Stats) StopSendingStats() {
	if s.stopGorutine != nil {
		s.stopGorutine()
		s.stopGorutine = nil
	}
}

func (s *Stats) Wait() {
	<-s.block
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

	memoryEmpty := float64(memory.Total-memory.Used) / Mb
	cpuUsage := float64(after.User-before.User) / total * 100

	if memoryEmpty < s.thresholdMemory {
		markRedMem = Red
	}

	if cpuUsage > s.thresholdCPU {
		markRedCpu = Red
	}

	if s.onlyOverThreshold && markRedMem == "" && markRedCpu == "" {
		return nil
	}

	err = s.report.Stats(s.appName,
		fmt.Sprintf("Memory Total: %0.3f GB\n", float64(memory.Total)/Gb),
		fmt.Sprintf("%sMemory Used: %0.3f GB\n", markRedMem, float64(memory.Used)/Gb),
		fmt.Sprintf("Memory Cached:  %0.3f GB\n", float64(memory.Cached)/Gb),
		fmt.Sprintf("Memory Free:  %0.3f GB\n\n", float64(memory.Free)/Gb),
		fmt.Sprintf("%sCPU user: %0.2f %%\n", markRedCpu, float64(after.User-before.User)/total*100),
		fmt.Sprintf("CPU system: %0.2f %%\n", float64(after.System-before.System)/total*100),
		fmt.Sprintf("CPU idle: %0.2f %%\n", float64(after.Idle-before.Idle)/total*100),
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Stats) Copy() *Stats {
	return &Stats{
		appName:           s.appName,
		report:            s.report,
		minutes:           s.minutes,
		thresholdMemory:   s.thresholdMemory,
		thresholdCPU:      s.thresholdCPU,
		onlyOverThreshold: s.onlyOverThreshold,
		block:             make(chan int),
	}
}
