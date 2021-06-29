package rest

import (
	"email-sender/internal/handlers/rest/acceptor"
	"email-sender/internal/services"
	"email-sender/internal/system/logger"
	"email-sender/internal/system/metrics" //nolint:goimports
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
)

type Handlers interface {
	RegisterRoutes()
}

type handlers struct {
	router           *fiber.App
	logger           *zap.Logger
	metrics          *metrics.Client
	acceptorHandlers acceptor.Handlers
}

func (h *handlers) RegisterRoutes() {
	h.router.Use(
		requestid.New(),
		logger.WithLogger(h.logger),
	)

	h.router.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("hello, world!")
	})

	api := h.router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			notifications := v1.Group("/notifications")
			{
				notifications.Get("", h.acceptorHandlers.ListNotifications)
				notifications.Get("/:id", h.acceptorHandlers.GetNotification)
				notifications.Post("", h.acceptorHandlers.SaveNotification)
			}
		}
	}
}

func New(
	router *fiber.App,
	logger *zap.Logger,
	metrics *metrics.Client,
	acceptorService *services.Acceptor,
) Handlers {
	return &handlers{
		router:           router,
		logger:           logger,
		metrics:          metrics,
		acceptorHandlers: acceptor.New(logger, acceptorService, metrics),
	}
}
