package slack

import (
	"fmt"
	"os"
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
	LIVE   = false
)

func replaceColors(str string) string {

	for _, color := range colors {
		str = strings.ReplaceAll(str, color, "")
	}
	return str
}

func report(webhook, message string, messages ...string) {

	err := slack.Send(webhook, "", transfToPayload(message, messages...))
	if len(err) > 0 {
		fmt.Printf("error: %s\n", err)
	}

}

func transfToPayload(message string, messages ...string) slack.Payload {
	payload := slack.Payload{
		Text:     fmt.Sprintf("_*%s*_", message),
		Username: "Scrapper",
		Channel:  "#scrapper",
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

func Report(message string, messages ...string) {

	report(os.Getenv("SLACK_WEBHOOK"), message, messages...)

}
func ReportStats(message string, messages ...string) {

	report(os.Getenv("SLACK_WEBHOOK_STATS"), message, messages...)
}

func ReportError(err error) {
	if _, ok := err.(tracerr.Error); !ok {
		err = tracerr.Wrap(err)
	}

	if LIVE {

		stacks := strings.Split(tracerr.SprintSourceColor(err), "\n")

		report(os.Getenv("SLACK_WEBHOOK"), stacks[0], stacks[1:]...)
	} else {
		tracerr.PrintSourceColor(err)
	}

}

func EnvLoaded() {
	LIVE = os.Getenv("LIVE") == "true"
}
