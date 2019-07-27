// Package slackrus provides a Slack hook for the logrus loggin package.
package slackrus

import (
	"fmt"

	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

// Project version
const (
	VERISON = "0.1.0"
)

// SlackrusHook is a logrus Hook for dispatching messages to the specified
// channel on Slack.
type SlackrusHook struct {
	// Messages with a log level not contained in this array
	// will not be dispatched. If nil, all messages will be dispatched.
	AcceptedLevels []logrus.Level
	LegacyToken    string
	IconURL        string
	Channel        string
	IconEmoji      string
	Username       string
	Asynchronous   bool
	Extra          map[string]interface{}
	Disabled       bool
}

// Levels sets which levels to sent to slack
func (sh *SlackrusHook) Levels() []logrus.Level {
	if sh.AcceptedLevels == nil {
		return AllLevels
	}
	return sh.AcceptedLevels
}

// Fire -  Sent event to slack
func (sh *SlackrusHook) Fire(e *logrus.Entry) error {
	api := slack.New(sh.LegacyToken)

	if sh.Disabled {
		return nil
	}

	color := ""
	switch e.Level {
	case logrus.DebugLevel:
		color = "#9B30FF"
	case logrus.InfoLevel:
		color = "good"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		color = "danger"
	default:
		color = "warning"
	}

	param := slack.PostMessageParameters{
		Username:  sh.Username,
		Channel:   sh.Channel,
		IconEmoji: sh.IconEmoji,
		IconURL:   sh.IconURL,
	}

	attach := slack.Attachment{}

	newEntry := sh.newEntry(e)
	// If there are fields we need to render them at attachments
	if len(newEntry.Data) > 0 {

		// Add a header above field data
		attach.Text = "Message fields"

		for k, v := range newEntry.Data {
			slackField := &slack.AttachmentField{}

			slackField.Title = k
			slackField.Value = fmt.Sprint(v)
			// If the field is <= 20 then we'll set it to short
			if len(slackField.Value) <= 20 {
				slackField.Short = true
			}

			attach.Fields = append(attach.Fields, *slackField)
		}
		attach.Pretext = newEntry.Message
	} else {
		attach.Text = newEntry.Message
	}
	attach.Fallback = newEntry.Message
	attach.Color = color

	if sh.Asynchronous {
		go api.PostMessage(sh.Channel, slack.MsgOptionPostMessageParameters(param), slack.MsgOptionAttachments(attach))
		return nil
	}

	_, _, err := api.PostMessage(sh.Channel, slack.MsgOptionPostMessageParameters(param), slack.MsgOptionAttachments(attach))
	return err
}

func (sh *SlackrusHook) newEntry(entry *logrus.Entry) *logrus.Entry {
	data := map[string]interface{}{}

	for k, v := range sh.Extra {
		data[k] = v
	}
	for k, v := range entry.Data {
		data[k] = v
	}

	newEntry := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    data,
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}

	return newEntry
}
