package models

import (
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// validation errors
var (
	MessageEmptyValidationErr = errors.New("empty message string")
	WrongEmailFormatErr       = errors.New("wrong email format")
	NoEmailsProvidedErr       = errors.New("no email provided")
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\\-]+@[a-z0-9.\\-]+\\.[a-z]{2,4}$`)

type Notification struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	PostNotification `bson:",inline"`
	SentStatus       bool      `json:"sent_status" bson:"sent_status"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
}

type PostNotification struct {
	Sender  string   `json:"sender,omitempty" bson:"sender"`
	To      []string `json:"to" bson:"to"`
	Subject string   `json:"subject,omitempty" bson:"subject"`
	Message string   `json:"message" bson:"message"`
}

func (p *PostNotification) Validate() error {
	var allErrors []error

	if p.Message == "" {
		allErrors = append(allErrors, MessageEmptyValidationErr)
	}

	if len(p.To) == 0 {
		allErrors = append(allErrors, NoEmailsProvidedErr)
	}

	for _, t := range p.To {
		if !isEmailValid(t) {
			allErrors = append(allErrors, errors.Errorf("%s is a %s", t, WrongEmailFormatErr.Error()))
		}
	}

	var errStr strings.Builder
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

func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
