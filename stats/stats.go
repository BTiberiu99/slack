package stats

import (
	"fmt"
	"os"
	"util/slack"

	"strconv"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/ztrue/tracerr"
)

var (
	start = false
)

func Init() {
	if start {
		return
	}

	start = true

	threshold_memory, err := strconv.ParseFloat(os.Getenv("THRESHOLD_MEMORY"), 64)

	if err != nil {
		slack.ReportError(err)
		return
	}

	threshold_cpu, err := strconv.ParseFloat(os.Getenv("THRESHOLD_CPU"), 64)

	if err != nil {
		slack.ReportError(err)
		return
	}

	go func() {

		for {
			time.Sleep(time.Hour * 1)
			if err := sendStats(threshold_memory, threshold_cpu); err != nil {
				slack.ReportError(err)
			}

		}

	}()

}

func sendStats(threshold_memory, threshold_cpu float64) error {

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
	if float64(memory.Total-memory.Used)/(1024*1024) < threshold_memory {
		markRedMem = slack.Red
	}

	markRedCpu := ""
	if float64(after.User-before.User)/total*100 > threshold_cpu {
		markRedCpu = slack.Red
	}

	slack.ReportStats(os.Getenv("APP_NAME"),
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
