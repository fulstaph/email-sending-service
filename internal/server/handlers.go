package server

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"projects/email-sending-service/internal/models"
	"projects/email-sending-service/internal/service"
	"strconv"
)

type APIErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (s *Server) getNotifications(c *gin.Context) {
	var params models.PaginationParams
	if err := c.Bind(&params); err != nil {
		s.l.Error(err)
		c.JSON(http.StatusBadRequest, APIErrorResponse{
			Code: http.StatusBadRequest,
			Msg:  err.Error(),
		})
		return
	}
	notifs, totalDocsCount, totalPagesCount, err := s.a.Get(params.PerPage, params.Page)
	if err != nil {
		switch err {
		case service.LimitNumberTooHighErr:
			s.l.Error(err)
			c.JSON(http.StatusBadRequest, APIErrorResponse{
				Code: http.StatusBadRequest,
				Msg:  err.Error(),
			})
			return
		default:
			s.l.Error(err)
			c.JSON(http.StatusInternalServerError, APIErrorResponse{
				Code: http.StatusInternalServerError,
				Msg:  err.Error(),
			})
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
			s.l.Warn(err)
			c.JSON(http.StatusNotFound, APIErrorResponse{
				Code: http.StatusNotFound,
				Msg:  err.Error(),
			})
			return
		case service.IdNotValidErr:
			s.l.Warn(err)
			c.JSON(http.StatusBadRequest, APIErrorResponse{
				Code: http.StatusBadRequest,
				Msg:  err.Error(),
			})
			return
		default:
			s.l.Error(err)
			c.JSON(http.StatusInternalServerError, APIErrorResponse{
				Code: http.StatusInternalServerError,
				Msg:  err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, notif)
}

func (s *Server) saveNotification(c *gin.Context) {
	var notif models.PostNotification
	if err := c.BindJSON(&notif); err != nil {
		s.l.Warn(err)
		c.JSON(http.StatusBadRequest, APIErrorResponse{
			Code: http.StatusBadRequest,
			Msg:  err.Error(),
		})
		return
	}

	if err := notif.Validate(); err != nil {
		s.l.Warnf("validation err: %v", err)
		c.JSON(http.StatusBadRequest, APIErrorResponse{
			Code: http.StatusBadRequest,
			Msg:  err.Error(),
		})
		return
	}

	id, err := s.a.Add(notif)
	if err != nil {
		s.l.Errorf("err in acceptor.Add: %v", err)
		c.JSON(http.StatusInternalServerError, APIErrorResponse{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"id": id})
}
