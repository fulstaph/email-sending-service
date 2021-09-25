package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"email-sender/config"
	"email-sender/internal/repositories"
	"email-sender/internal/system/logger"
	"email-sender/internal/system/metrics"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const rmqLabelPrefix = "rmq_"

var HandlerNotRegisterErr = errors.New("handler not registered")

type (
	QueueHandler interface {
		Handle(ctx context.Context, message interface{}) error
		Message() interface{}
	}

	Handler interface {
		Handle(ctx context.Context, queueName string, message amqp.Delivery)
	}

	handler struct {
		handlers map[string]QueueHandler
		metrics  *metrics.Client
	}
)

func NewHandler(config *config.Consumer, cfg *config.SMTP, metrics *metrics.Client, repos *repositories.Container) Handler {
	h := &handler{
		metrics: metrics,
		handlers: map[string]QueueHandler{
			config.NotificationsQueue.Name: newNotificationEventHandler(repos, cfg),
		},
	}

	return h
}

func (h handler) Handle(ctx context.Context, queueName string, message amqp.Delivery) {
	log := logger.Fetch(ctx)
	msgLogger := log.
		With(zap.String("queue_name", queueName)).
		With(zap.String("decoded_value", string(message.Body)))

	startTime := time.Now()
	defer func() {
		h.metrics.RMQMessagesProcessingTime.Add(makeMessageType(queueName), time.Since(startTime).Seconds())
	}()

	h.metrics.RMQMessageCount.AddClaimed(queueName, makeMessageType(queueName))

	handler, err := h.findHandler(queueName)
	if errors.Is(err, HandlerNotRegisterErr) {
		msgLogger.With(zap.Error(err)).Info("skip message")
		h.metrics.RMQMessageCount.AddSkipped(queueName, makeMessageType(queueName))
		return
	} else if err != nil {
		msgLogger.With(zap.Error(err)).Error("find message handler error")
		h.metrics.RMQMessageCount.AddFailed(queueName, makeMessageType(queueName))
		return
	}

	handlerMsg := handler.Message()
	if err := json.Unmarshal(message.Body, &handlerMsg); err != nil {
		msgLogger.With(zap.Error(err)).Error("parse handler message error")
		h.metrics.RMQMessageCount.AddFailed(queueName, makeMessageType(queueName))
		return
	}

	handlerLogger := log.With(zap.String("queue_name", queueName))
	if err := handler.Handle(logger.Enrich(ctx, handlerLogger), handlerMsg); err != nil {
		handlerLogger.With(zap.Error(err)).Error("message handle error")
		h.metrics.RMQMessageCount.AddFailed(queueName, makeMessageType(queueName))
		return
	}

	h.metrics.RMQMessageCount.AddSuccess(queueName, makeMessageType(queueName))
}

func (h handler) findHandler(queueName string) (QueueHandler, error) {
	handler, ok := h.handlers[queueName]
	if ok {
		return handler, nil
	}

	return handler, HandlerNotRegisterErr
}

func makeMessageType(eventName string) string {
	return rmqLabelPrefix + eventName
}
