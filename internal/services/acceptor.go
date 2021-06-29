package services

import (
	"context"
	"time"

	"email-sender/config"
	"email-sender/internal/entities"
	"email-sender/internal/repositories"
	"email-sender/internal/system/broker/events"
	"email-sender/internal/system/broker/producer"
	"email-sender/internal/system/logger"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var (
	ErrIDNotValid         = errors.New("id is not valid")
	ErrLimitNumberTooHigh = errors.New("limit number is too high")
)

type Acceptor struct {
	repos    *repositories.Container
	producer producer.Producer
	cfg      *config.Producer
}

func (a *Acceptor) Get(ctx context.Context, notificationID string) (*entities.Notification, error) {
	id, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return nil, ErrIDNotValid
	}

	notification, err := a.repos.Emails.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

func (a *Acceptor) List(ctx context.Context, limit, skip int64) ([]entities.Notification, int64, int64, error) {
	if limit == 0 {
		return nil, 0, 0, nil
	}

	if limit > 1000 {
		return nil, 0, 0, ErrLimitNumberTooHigh
	}

	notifs, totalDocsCount, err := a.repos.Emails.List(ctx, limit, skip)
	if err != nil {
		return nil, 0, 0, err
	}

	var totalPagesCount = getPagesCount(totalDocsCount, limit)

	return notifs, totalDocsCount, totalPagesCount, nil
}

func (a *Acceptor) Save(ctx context.Context, notification *entities.PostNotification) (string, error) {
	log := logger.Fetch(ctx)

	fullNotification := &entities.Notification{
		ID:               primitive.NewObjectID(),
		PostNotification: *notification,
		SentStatus:       false,
		CreatedAt:        time.Now(),
	}

	event, err := events.NewNotificationCreatedEvent(fullNotification, a.cfg.Exchange)
	if err != nil {
		log.With(zap.Error(err)).Error("error creating notification event")
		return "", err
	}

	if err := a.producer.Produce(event); err != nil {
		log.With(zap.Error(err)).Error("error producing notification event")
		return "", err
	}

	return fullNotification.ID.Hex(), nil
}

func NewAcceptor(repos *repositories.Container, producer producer.Producer, cfg *config.Producer) *Acceptor {
	return &Acceptor{
		repos:    repos,
		producer: producer,
		cfg:      cfg,
	}
}

func getPagesCount(totalCount, perPageCount int64) int64 {
	if perPageCount > 0 {
		result := totalCount / perPageCount
		if result > 0 && (totalCount > (perPageCount * result)) {
			return result + 1
		}
		return result
	}
	return 1
}
