package applications

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"email-sender/config"
	"email-sender/internal/handlers/rest"
	"email-sender/internal/repositories"
	"email-sender/internal/services"
	"email-sender/internal/system/broker/producer"
	"email-sender/internal/system/database/mongodb"
	"email-sender/internal/system/logger"
	"email-sender/internal/system/metrics" //nolint:goimports

	"github.com/gofiber/fiber/v2"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Acceptor struct {
	logger         *zap.Logger
	config         *config.ConfigAcceptor
	metricsClient  *metrics.Client
	metricsServer  *http.Server
	server         *fiber.App
	handlers       rest.Handlers
	mongoClient    mongodb.Client
	producer       producer.Producer
	producerClient producer.Client
}

func NewAcceptor() (*Acceptor, error) {
	cfg, err := config.LoadAcceptor()
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

	client, err := producer.NewClient(cfg.Producer, appLogger)
	if err != nil {
		return nil, err
	}

	metricsClient := metrics.New()
	metricsServer := &http.Server{Addr: cfg.MetricsPort}

	producer := producer.New(client)

	acceptor := services.NewAcceptor(repos, producer, cfg.Producer)

	server := fiber.New()
	handlers := rest.New(server, appLogger, metricsClient, acceptor)

	return &Acceptor{
		config:         cfg,
		logger:         appLogger,
		mongoClient:    mongoClient,
		metricsClient:  metricsClient,
		metricsServer:  metricsServer,
		producerClient: client,
		producer:       producer,
		handlers:       handlers,
		server:         server,
	}, nil
}

func (a *Acceptor) Run() error {
	a.handlers.RegisterRoutes()

	go func() {
		a.logger.Info("run application", zap.String("port", a.config.Port))
		if err := a.server.Listen(a.config.Port); err != nil {
			a.logger.Error("failed to listening", zap.Error(err))
		}
	}()

	go func() {
		http.Handle("/metrics", a.metricsClient.Handler())
		a.logger.Sugar().Infof("start metrics http serve on port: %v!", a.config.MetricsPort)
		if err := a.metricsServer.ListenAndServe(); err != nil {
			a.logger.With(zap.Error(err)).Error("metrics server serve error")
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	return a.shutdown()
}

func (a *Acceptor) shutdown() (err error) {
	if mongoCloseErr := a.mongoClient.Close(); mongoCloseErr != nil {
		err = multierr.Append(err, mongoCloseErr)
	}
	if metricsCloseErr := a.metricsServer.Close(); metricsCloseErr != nil {
		err = multierr.Append(err, metricsCloseErr)
	}
	if producerCloseErr := a.producerClient.Close(); producerCloseErr != nil {
		err = multierr.Append(err, producerCloseErr)
	}
	if serverShutdownError := a.server.Shutdown(); serverShutdownError != nil {
		err = multierr.Append(err, serverShutdownError)
	}
	return
}
