package applications

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"email-sender/config"
	"email-sender/internal/handlers/rabbitmq"
	"email-sender/internal/repositories"
	"email-sender/internal/system/broker/consumer"
	"email-sender/internal/system/database/mongodb"
	"email-sender/internal/system/logger"
	"email-sender/internal/system/metrics"

	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Sender struct {
	logger        *zap.Logger
	config        *config.ConfigSender
	metricsClient *metrics.Client
	metricsServer *http.Server
	mongoClient   mongodb.Client
	consumer      consumer.Consumer
}

func NewSender() (*Sender, error) {
	cfg, err := config.LoadSender()
	if err != nil {
		return nil, err
	}

	appLogger, err := logger.New(cfg.LogLevel, cfg.AppName)
	if err != nil {
		return nil, err
	}

	mongoClient, err := mongodb.NewClient(cfg.Database)
	if err != nil {
		return nil, err
	}

	repos := repositories.New(mongoClient.GetConnection())

	metricsClient := metrics.New()
	metricsServer := &http.Server{Addr: cfg.MetricsPort}

	rmqHandler := rabbitmq.NewHandler(cfg.Consumer, cfg.SMTP, metricsClient, repos)
	rmqConsumer, err := consumer.NewConsumer(cfg.Consumer, rmqHandler, appLogger, metricsClient)
	if err != nil {
		return nil, err
	}

	return &Sender{
		config:        cfg,
		logger:        appLogger,
		mongoClient:   mongoClient,
		metricsClient: metricsClient,
		metricsServer: metricsServer,
		consumer:      rmqConsumer,
	}, nil
}

func (s *Sender) Run() error {
	s.consumer.Consume()

	go func() {
		http.Handle("/metrics", s.metricsClient.Handler())
		s.logger.Sugar().Infof("start metrics http serve on port: %v!", s.config.MetricsPort)
		if err := s.metricsServer.ListenAndServe(); err != nil {
			s.logger.With(zap.Error(err)).Error("metrics server serve error")
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	return s.shutdown()
}

func (s *Sender) shutdown() (err error) {
	if mongoCloseErr := s.mongoClient.Close(); mongoCloseErr != nil {
		err = multierr.Append(err, mongoCloseErr)
	}
	if metricsCloseErr := s.metricsServer.Close(); metricsCloseErr != nil {
		err = multierr.Append(err, metricsCloseErr)
	}
	if consumerCloseErr := s.consumer.Close(); consumerCloseErr != nil {
		err = multierr.Append(err, consumerCloseErr)
	}
	return
}
