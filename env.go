package slack

import (
	"strconv"
	"strings"

	"github.com/gobuffalo/envy"
)

func FromEnvReport() (*Report, error) {
	slackWebhook, err := envy.MustGet("SLACK_WEBHOOK")

	if err != nil {
		return nil, err
	}

	slackWebhookStats, err := envy.MustGet("SLACK_WEBHOOK_STATS")

	if err != nil {
		return nil, err
	}

	report, err := NewReport(&ConfigReport{
		Print:        strings.ToLower(envy.Get("LIVE", "true")) != "true",
		Webhook:      slackWebhook,
		WebhookStats: slackWebhookStats,
	})

	if err != nil {
		return nil, err
	}

	return report, nil
}

func FromEnvStats(report *Report) (*Stats, error) {

	if report == nil {
		return nil, errNoReport
	}

	minutes, err := strconv.ParseInt(envy.Get("REPORT_STATS_MINUTES", "30"), 10, 64)

	if err != nil {

		return nil, err
	}

	thresholdMemory, err := strconv.ParseFloat(envy.Get("THRESHOLD_MEMORY", "1024"), 64)

	if err != nil {

		return nil, err
	}

	thresholdCPU, err := strconv.ParseFloat(envy.Get("THRESHOLD_CPU", "80"), 64)

	if err != nil {

		return nil, err
	}

	stats, err := NewStats(&ConfigStats{
		Report:          report,
		Minutes:         int(minutes),
		ThresholdCPU:    thresholdCPU,
		ThresholdMemory: thresholdMemory,
		AppName:         envy.Get("APP_NAME", "Default App Name"),
	})

	return stats, nil

}

func FromEnv() (*Report, *Stats, error) {

	report, err := FromEnvReport()

	if err != nil {
		return nil, nil, err
	}

	stats, err := FromEnvStats(report)

	if err != nil {
		return nil, nil, err
	}

	return report, stats, nil
}
