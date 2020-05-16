package slack

import (
	"strconv"
	"strings"

	"github.com/gobuffalo/envy"
)

func FromEnv() (*Report, *Stats) {
	slackWebhook, err := envy.MustGet("SLACK_WEBHOOK")

	if err != nil {
		panic(err)
	}

	slackWebhookStats, err := envy.MustGet("SLACK_WEBHOOK_STATS")

	if err != nil {
		panic(err)
	}

	report, err := NewReport(&ConfigReport{
		Print:        strings.ToLower(envy.Get("LIVE", "true")) != "true",
		Webhook:      slackWebhook,
		WebhookStats: slackWebhookStats,
	})

	if err != nil {
		panic(err)
	}

	minutes, err := strconv.ParseInt(envy.Get("REPORT_STATS_MINUTES", "30"), 10, 64)

	if err != nil {
		report.Error(err)
		return report, nil
	}

	thresholdMemory, err := strconv.ParseFloat(envy.Get("THRESHOLD_MEMORY", "1024"), 64)

	if err != nil {
		report.Error(err)
		return report, nil
	}

	thresholdCPU, err := strconv.ParseFloat(envy.Get("THRESHOLD_CPU", "80"), 64)

	if err != nil {
		report.Error(err)
		return report, nil
	}

	stats, err := NewStats(&ConfigStats{
		Report:          report,
		Minutes:         int(minutes),
		ThresholdCPU:    thresholdCPU,
		ThresholdMemory: thresholdMemory,
		AppName:         envy.Get("APP_NAME", "Default App Name"),
	})

	return report, stats
}
