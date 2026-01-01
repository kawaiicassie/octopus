package handlers

import (
	"net/http"

	"github.com/bestruirui/octopus/internal/server/resp"
	"github.com/bestruirui/octopus/internal/server/router"
	"github.com/gin-gonic/gin"
)

func init() {
	r := router.NewGroupRouter("")

	r.AddRoute(
		router.NewRoute("/health", http.MethodGet).
			Handle(health),
	)

	r.AddRoute(
		router.NewRoute("/ping", http.MethodGet).
			Handle(ping),
	)
}

func health(c *gin.Context) {
	resp.Success(c, gin.H{
		"status": "healthy",
		"service": "octopus",
	})
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"pong": true})
}