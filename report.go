package slack

import (
	"errors"
	"fmt"
	"regexp"
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
	colors             = []string{Black, Red, EndColor, White}
	regexReplaceColors = regexp.MustCompile(strings.ReplaceAll(strings.Join(colors, "|"), "[", `\[`))
)

var (
	errWebhookStats = errors.New("WebhookStats is not defined!")
	errWebhook      = errors.New("Webhook is not defined!")
)

type Report struct {
	print bool

	webhook string

	webhookStats string

	withStack bool
}

type ConfigReport struct {

	//Print or send to slack errors
	Print bool

	//Webhook for errors
	Webhook string

	//Webhook for sending stats
	WebhookStats string

	WithStack bool
}

func NewReport(config *ConfigReport) (*Report, error) {

	if config.Webhook == "" {
		return nil, errWebhook
	}

	return &Report{
		print:        config.Print,
		webhook:      config.Webhook,
		webhookStats: config.WebhookStats,
		withStack:    config.WithStack,
	}, nil
}

//Sends messages to slack
func send(webhook, message string, messages []string, withStack bool) error {

	errs := slack.Send(webhook, "", transfToPayload(message, messages, withStack))

	//Create only one error from multiple errors
	if len(errs) > 0 {
		errString := ""

		for i := 0; i < len(errs); i++ {
			errString += errs[i].Error() + "\n"
		}

		return errors.New(errString)
	}

	return nil

}

func transfToPayload(message string, messages []string, withStack bool) slack.Payload {
	payload := slack.Payload{
		Text:     fmt.Sprintf("_*%s*_", message),
		Markdown: true,
	}

	lenM := len(messages)

	if lenM == 0 || !withStack {
		return payload
	}

	attachments := make([]slack.Attachment, lenM)
	red := hexRed

	max := 80

	//Create message
	for i := 0; i < lenM && i < max; i++ {

		text := regexReplaceColors.ReplaceAllLiteralString(messages[i], "")

		attachment := slack.Attachment{
			Text: &text,
		}

		if strings.Contains(messages[i], Red) {
			attachment.Color = &red
		}

		attachments[i] = attachment
	}

	payload.Attachments = attachments

	return payload
}

//Stats ... Send stats to slack
func (r *Report) Stats(message string, messages ...string) error {
	if r.webhookStats == "" {
		return errWebhookStats
	}
	return send(r.webhookStats, message, messages, r.withStack)
}

//Error ... prints or sends error to slack
func (r *Report) Error(err error) error {

	//Add wrapper
	if _, ok := err.(tracerr.Error); !ok {
		err = tracerr.Wrap(err)
	}

	if r.print {

		//Print to console
		tracerr.PrintSourceColor(err)

	} else {

		//Send to slack
		stacks := strings.Split(tracerr.SprintSourceColor(err), "\n")

		return send(r.webhook, stacks[0], stacks[1:], r.withStack)
	}

	return nil

}

//ErrorAndPanic ... prints or sends error to slack and panics
func (r *Report) ErrorAndPanic(err error) {
	//send error
	r.Error(err)

	//panic
	panic(err)
}
