package rabbitmq

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"email-sender/config"
	"email-sender/internal/handlers/rabbitmq/queues"
	"email-sender/internal/repositories"
	"email-sender/internal/system/logger"

	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type notificationEventHandler struct {
	repos *repositories.Container
	cfg   *config.SMTP
}

func (n *notificationEventHandler) Handle(ctx context.Context, message interface{}) (err error) {
	event, ok := message.(*queues.NotificationEvent)
	if !ok {
		return
	}

	notification := event.Payload
	notification.SentStatus = true

	if sendErr := n.send(ctx, notification.Subject, notification.Message, notification.To); sendErr != nil {
		err = multierr.Append(err, sendErr)
		notification.SentStatus = false
	}

	if _, saveErr := n.repos.Emails.Save(ctx, notification); saveErr != nil {
		err = multierr.Append(err, saveErr)
	}

	if err != nil {
		logger.Fetch(ctx).With(zap.Error(err)).Error("error processing notification")
	}

	return
}

func (n *notificationEventHandler) Message() interface{} {
	return &queues.NotificationEvent{}
}

func newNotificationEventHandler(repos *repositories.Container, cfg *config.SMTP) QueueHandler {
	return &notificationEventHandler{repos: repos, cfg: cfg}
}

func (n *notificationEventHandler) send(ctx context.Context, subject, message string, to []string) error {
	auth := smtp.PlainAuth("", n.cfg.Username, n.cfg.Password, n.cfg.Host)

	// Sending email.
	err := smtp.SendMail(n.cfg.Host+n.cfg.Port,
		auth,
		n.cfg.Username,
		to,
		constructEmailMsg(to, subject, message),
	)
	if err != nil {
		logger.Fetch(ctx).With(zap.Error(err))
		return err
	}

	logger.Fetch(ctx).Info(fmt.Sprintf("mail successfully sent to %v", to))
	return nil
}

func constructEmailMsg(to []string, subject, message string) []byte {
	return []byte(
		fmt.Sprintf("To: %s \r\n", strings.Join(to, ",")) +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			message + "\r\n",
	)
}
