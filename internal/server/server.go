package server

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"projects/email-sending-service/config"
	"projects/email-sending-service/internal/service"
)

type Server struct {
	a   service.Acceptor
	r   *gin.Engine
	cfg config.Server
}

func (s *Server) Start() {
	s.r = gin.Default()
	s.r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := s.r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			notifs := v1.Group("/notifs")
			{
				notifs.GET("", s.getNotifications)
				notifs.GET("/:id", s.getNotification)
				notifs.POST("", s.saveNotification)
			}
		}
	}

	if err := s.r.Run(s.cfg.Port); err != nil {
		log.Fatal(err)
	}
}

func New(
	a service.Acceptor,
	cfg config.Server,
) *Server {
	return &Server{
		a: a,
		cfg:  cfg,
	}
}
