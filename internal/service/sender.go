package service

import (
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/smtp"
	"projects/email-sending-service/config"
	"projects/email-sending-service/internal/broker"
	"projects/email-sending-service/internal/models"
	"projects/email-sending-service/internal/repository"
	"strings"
)

type Sender interface {
	Send(subject, message string, to []string) error
	Start() error
}

type sender struct {
	r   repository.EmailRepository
	mq  broker.MessageQueue
	cfg config.MailingServer
}

func (s *sender) Start() error {
	if err := s.mq.Subscribe(s.processIncomingNotification); err != nil {
		return err
	}
	return nil
}

func (s *sender) Send(subject, message string, to []string) error {
	// Authentication.
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	msg := []byte(
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			 message + "\r\n",
	)

	// Sending email.
	err := smtp.SendMail(s.cfg.Host + s.cfg.Port, auth, s.cfg.Username, to, msg)
	if err != nil {
		log.Errorln(err)
		return err
	}
	log.Infof("mail successfully sent to %v\n", to)
	return nil
}

func (s *sender) processIncomingNotification(data []byte) error {
	var allErrors []error

	var notif models.Notification
	if err := json.Unmarshal(data, &notif); err != nil {
		allErrors = append(allErrors, err)
		return err
	}

	notif.SentStatus = true

	// TODO: sent emails here
	err := s.Send(notif.Subject, notif.Message, notif.To)
	if err != nil {
		allErrors = append(allErrors, err)
		notif.SentStatus = false
	}

	// save new notif in DB
	if _, err := s.r.Save(notif); err != nil {
		allErrors = append(allErrors, err)
	}

	errStr := strings.Builder{}
	for _, err := range allErrors {
		if err != nil {
			errStr.WriteString(err.Error())
			errStr.WriteString(" ")
		}
	}

	if errStr.String() != "" {
		return errors.New(errStr.String())
	}

	return nil
}

func NewSender(r repository.EmailRepository,
	mq broker.MessageQueue,
	cfg config.MailingServer) Sender {
	return &sender{
		r:  r,
		mq: mq,
		cfg: cfg,
	}
}
