package handlers

import (
	"net/http"

	"github.com/bestruirui/octopus/internal/conf"
	"github.com/bestruirui/octopus/internal/server/middleware"
	"github.com/bestruirui/octopus/internal/server/resp"
	"github.com/bestruirui/octopus/internal/server/router"
	"github.com/gin-gonic/gin"
)

func init() {
	// Replace original update handlers with disabled versions
	r := router.NewGroupRouter("/api/v1/update").
		Use(middleware.Auth()).
		Use(middleware.RequireJSON())

	// Return current version as both current and latest
	r.AddRoute(
		router.NewRoute("", http.MethodGet).
			Handle(func(c *gin.Context) {
				// Return current version as latest to prevent update prompt
				resp.Success(c, gin.H{
					"tag_name": conf.Version,
					"body": "Auto-update is disabled on Render",
				})
			}),
	)

	r.AddRoute(
		router.NewRoute("/now-version", http.MethodGet).
			Handle(func(c *gin.Context) {
				resp.Success(c, conf.Version)
			}),
	)

	// Disable actual update
	r.AddRoute(
		router.NewRoute("", http.MethodPost).
			Handle(func(c *gin.Context) {
				resp.Error(c, http.StatusNotImplemented, "Auto-update is not supported on Render. Please redeploy manually.")
			}),
	)
}