package service

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"projects/email-sending-service/internal/models"
	"projects/email-sending-service/internal/repository"
	"time"
)

var (
	IdNotValidErr = errors.New("id is not valid")
	LimitNumberTooHighErr = errors.New("limit number is too high")
)

type Acceptor interface {
	GetByID(id string) (models.Notification, error)
	Get(limit, skip int) ([]models.Notification, int64, int64, error)
	Add(notif models.PostNotification) (string, error)
}

type acceptor struct {
	r repository.EmailRepository
}

func (a *acceptor) GetByID(id string) (models.Notification, error) {
	oID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return models.Notification{}, IdNotValidErr
	}

	notif, err := a.r.GetByID(oID)
	if err != nil {
		return models.Notification{}, err
	}

	return notif, nil
}

func (a *acceptor) Get(limit, skip int) ([]models.Notification, int64, int64, error) {
	if limit > 1000 {
		return nil, 0, 0, LimitNumberTooHighErr
	}

	notifs, totalDocsCount, err :=  a.r.Get(limit, skip)
	if err != nil {
		return nil, 0, 0, err
	}

	var totalPagesCount = getPagesCount(totalDocsCount, int64(limit))

	return notifs, totalDocsCount, totalPagesCount, nil
}

func (a *acceptor) Add(notif models.PostNotification) (string, error) {
	fullNotification := models.Notification{
		ID:               primitive.NewObjectID(),
		PostNotification: notif,
		SentStatus:       false,
		CreatedAt:        time.Now(),
	}

	// TODO: send to MQ here

	return fullNotification.ID.Hex(), nil
}

func NewAcceptor(r repository.EmailRepository) Acceptor {
	return &acceptor{r: r}
}

func getPagesCount(totalCount, perPageCount int64) int64 {
	if perPageCount > 0 {
		result := totalCount / perPageCount
		if result > 0 && (totalCount > (perPageCount * result)) {
			result++
		}
	}
	return 1
}