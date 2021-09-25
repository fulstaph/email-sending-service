package acceptor

import (
	"net/http"
	"strconv"

	"email-sender/internal/entities"
	"email-sender/internal/services"
	"email-sender/internal/system/metrics" //nolint:goimports

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handlers interface {
	ListNotifications(c *fiber.Ctx) error
	GetNotification(c *fiber.Ctx) error
	SaveNotification(c *fiber.Ctx) error
}

type acceptorHandlers struct {
	logger   *zap.Logger
	acceptor *services.Acceptor
	metrics  *metrics.Client
}

func New(
	logger *zap.Logger,
	acceptor *services.Acceptor,
	metrics *metrics.Client,
) Handlers {
	return &acceptorHandlers{
		logger:   logger,
		acceptor: acceptor,
		metrics:  metrics,
	}
}

func (h *acceptorHandlers) ListNotifications(c *fiber.Ctx) error {
	var params PaginationParams
	if err := bindRequestParams(c, &params); err != nil {
		h.logger.With(zap.Error(err))
		return fiber.NewError(http.StatusBadRequest, "error binding request parameters")
	}

	notification, totalDocsCount, totalPagesCount, err := h.acceptor.List(c.Context(), params.PerPage, params.Page)
	if err != nil {
		switch err {
		case services.ErrLimitNumberTooHigh:
			h.logger.With(zap.Error(err)).Warn("limit is too big")
			return fiber.NewError(http.StatusBadRequest, "limit is greater than 1000")
		default:
			h.logger.With(zap.Error(err)).Error("error acceptor.List")
			return fiber.NewError(http.StatusInternalServerError, "error fetching notifications")
		}
	}

	c.Set("X-Total", strconv.FormatInt(totalDocsCount, 10))
	c.Set("X-Total-Pages", strconv.FormatInt(totalPagesCount, 10))
	c.Set("X-Per-Page", strconv.FormatInt(params.PerPage, 10))
	c.Set("X-Page", strconv.FormatInt(params.Page, 10))
	return c.Status(http.StatusOK).JSON(notification)
}

func (h *acceptorHandlers) GetNotification(c *fiber.Ctx) error {
	id := c.Params("id")
	notification, err := h.acceptor.Get(c.Context(), id)
	if err != nil {
		switch err {
		case services.ErrIDNotValid:
			h.logger.With(zap.Error(err)).Warn("id not valid")
			return fiber.NewError(http.StatusBadRequest, "invalid id param")
		default:
			h.logger.With(zap.Error(err)).Error("error in GetNotification")
			return fiber.NewError(http.StatusInternalServerError, "error fetching notification")
		}
	}

	return c.Status(http.StatusOK).JSON(notification)
}

func (h *acceptorHandlers) SaveNotification(c *fiber.Ctx) error {
	var notification entities.PostNotification
	if err := c.BodyParser(&notification); err != nil {
		h.logger.With(zap.Error(err)).Warn("error binding notification")
		return fiber.NewError(http.StatusBadRequest, "error binding notification")
	}

	if err := notification.Validate(); err != nil {
		h.logger.With(zap.Error(err)).Warn("error validation notification")
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	id, err := h.acceptor.Save(c.Context(), &notification)
	if err != nil {
		h.logger.With(zap.Error(err)).Error("error in acceptor.Save")
		return fiber.NewError(http.StatusInternalServerError, "error saving notification")
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"id": id})
}
