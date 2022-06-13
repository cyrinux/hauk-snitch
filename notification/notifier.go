package notification

import (
	"fmt"
	"log"
	"net/smtp"

	"net/http"
	"net/url"

	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
)

// Notifier can send email notifications about events in the mapper
type Notifier interface {
	NotifyNewSession(topic string, URL string)
}

type notifier struct {
	config Config
}

// New returns a new Notifier instance
func New(config Config) Notifier {
	return &notifier{config: config}
}

func (t *notifier) NotifyNewSession(topic string, URL string) {
	if t.config.Gotify.Enabled {
		myURL, _ := url.Parse(t.config.Gotify.URL)
		client := gotify.NewClient(myURL, &http.Client{})

		params := message.NewCreateMessageParams()
		params.Body = &models.MessageExternal{
			Title:    "Hauk-Snitch",
			Message:  fmt.Sprintf("Forwarding %s to Hauk\r\n\r\nNew session: %s", topic, URL),
			Priority: t.config.Gotify.Priority,
		}
		_, err := client.Message.CreateMessage(params, auth.TokenAuth(t.config.Gotify.AppToken))

		if err != nil {
			log.Fatalf("Gotify: could not send message %v", err)
			return
		}
		log.Println("Gotify: message Sent!")
	}

	if t.config.Smtp.Enabled {
		host := fmt.Sprintf("%s:%d", t.config.Smtp.Host, t.config.Smtp.Port)
		err := smtp.SendMail(host, nil, t.config.Smtp.From, []string{t.config.Smtp.To}, []byte(fmt.Sprintf("Subject: Forwarding %s to Hauk\r\n\r\nNew session: %s", topic, URL)))
		if err != nil {
			log.Printf("Smtp: could not send email notification: %v", err)
		}
	}
}
