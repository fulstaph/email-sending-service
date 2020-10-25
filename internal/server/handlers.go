package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"projects/email-sending-service/internal/models"
	"projects/email-sending-service/internal/service"
	"strconv"
)

func (s *Server) getNotifications(c *gin.Context) {
	var params models.PaginationParams
	if err := c.Bind(&params); err != nil {
		log.Error(err)
		c.Status(http.StatusBadRequest)
		return
	}
	notifs, totalDocsCount, totalPagesCount, err := s.a.Get(params.PerPage, params.Page)
	if err != nil {
		switch err {
		case service.LimitNumberTooHighErr:
			log.Error(err)
			c.Status(http.StatusBadRequest)
			return
		default:
			log.Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Header("X-Total", strconv.FormatInt(totalDocsCount, 10))
	c.Header("X-Total-Pages", strconv.FormatInt(totalPagesCount, 10))
	c.Header("X-Per-Page", strconv.FormatInt(int64(params.PerPage), 10))
	c.Header("X-Page", strconv.FormatInt(int64(params.Page), 10))
	c.JSON(http.StatusOK, notifs)
}

func (s *Server) getNotification(c *gin.Context) {
	id := c.Param("id")
	notif, err := s.a.GetByID(id)
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			log.Error(err)
			c.Status(http.StatusNotFound)
			return
		case service.IdNotValidErr:
			log.Error(err)
			c.Status(http.StatusBadRequest)
			return
		default:
			log.Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, notif)
}

func (s *Server) saveNotification(c *gin.Context) {
	var notif models.PostNotification
	if err := c.BindJSON(&notif); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	id, err := s.a.Add(notif)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{ "id": id })
}