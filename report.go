package slack

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/ztrue/tracerr"
)

const (
	Black    = "[30m"
	Red      = "[31m"
	EndColor = "[0m"
	White    = "[1m"
	hexRed   = "#FF0000"
)

var (
	colors = []string{Black, Red, EndColor, White}
)

func replaceColors(str string) string {

	for _, color := range colors {
		str = strings.ReplaceAll(str, color, "")
	}
	return str
}

type Report struct {
	print        bool
	webhook      string
	webhookStats string
}

type ConfigReport struct {
	Print        bool
	Webhook      string
	WebhookStats string
}

func NewReport(config *ConfigReport) *Report {
	return &Report{
		print:        config.Print,
		webhook:      config.WebhookStats,
		webhookStats: config.WebhookStats,
	}
}

func send(webhook, message string, messages ...string) error {

	errs := slack.Send(webhook, "", transfToPayload(message, messages...))

	if len(errs) > 0 {
		errString := ""

		for i := 0; i < len(errs); i++ {
			errString += errs[i].Error() + "\n"
		}

		return errors.New(errString)
	}

	return nil

}

func transfToPayload(message string, messages ...string) slack.Payload {
	payload := slack.Payload{
		Text:     fmt.Sprintf("_*%s*_", message),
		Markdown: true,
	}

	if len(messages) > 0 {
		attachments := make([]slack.Attachment, len(messages))
		for i := 0; i < len(messages); i++ {

			text := replaceColors(messages[i])

			attachments[i] = slack.Attachment{
				Text: &text,
			}

			if strings.Contains(messages[i], Red) {
				red := hexRed
				attachments[i].Color = &red
			}
		}

		payload.Attachments = attachments

	}

	return payload
}

func (r *Report) Stats(message string, messages ...string) error {

	return send(r.webhookStats, message, messages...)
}

func (r *Report) Error(err error) error {
	if _, ok := err.(tracerr.Error); !ok {
		err = tracerr.Wrap(err)
	}

	if r.print {

		stacks := strings.Split(tracerr.SprintSourceColor(err), "\n")

		return send(r.webhook, stacks[0], stacks[1:]...)
	} else {
		tracerr.PrintSourceColor(err)
	}

	return nil

}
