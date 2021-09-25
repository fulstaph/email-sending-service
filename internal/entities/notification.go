package entities

import (
	"regexp"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/multierr"
)

// validation errors
var (
	ErrMessageEmptyValidation = errors.New("empty message string")
	ErrWrongEmailFormat       = errors.New("wrong email format")
	ErrNoEmailsProvided       = errors.New("no email provided")
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Notification struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	PostNotification `bson:",inline"`
	SentStatus       bool      `json:"sent_status" bson:"sent_status"` //nolint:tagliatelle
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`   //nolint:tagliatelle
}

type PostNotification struct {
	Sender  string   `json:"sender,omitempty" bson:"sender"`
	To      []string `json:"to" bson:"to"`
	Subject string   `json:"subject,omitempty" bson:"subject"`
	Message string   `json:"message" bson:"message"`
}

func (p *PostNotification) Validate() (err error) {
	if p.Message == "" {
		err = multierr.Append(err, ErrMessageEmptyValidation)
	}

	if len(p.To) == 0 {
		err = multierr.Append(err, ErrNoEmailsProvided)
	}

	for _, t := range p.To {
		if !isEmailValid(t) {
			err = multierr.Append(err, errors.Errorf("%s is a %s", t, ErrWrongEmailFormat.Error()))
		}
	}
	return
}

func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
