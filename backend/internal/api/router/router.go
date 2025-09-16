package router

import (
	"github.com/wb-go/wbf/ginext"

	"github.com/aliskhannn/url-shortener/internal/api/handlers/analytics"
	"github.com/aliskhannn/url-shortener/internal/api/handlers/link"
	"github.com/aliskhannn/url-shortener/internal/middleware"
)

// New creates a new Gin engine with routes and middlewares for the notification API.
//
// It applies standard middlewares (CORS, logging, recovery) and sets up the
// /api group with the following routes:
//   - POST	/api/shorten				-> linkHandler.ShortenLink
//   - GET	/api/s/:short_url" 			-> linkHandler.RedirectLink
//   - GET	/api/analytics/:short_url 	-> analyticsHandler.GetAnalytics
func New(linkHandler *link.Handler, analyticsHandler *analytics.Handler) *ginext.Engine {
	// Create a new Gin engine using the extended gin wrapper.
	e := ginext.New()

	// Apply middlewares: CORS, logger, and recovery.
	e.Use(middleware.CORSMiddleware())
	e.Use(ginext.Logger())
	e.Use(ginext.Recovery())

	// Create an API group for notifications.
	api := e.Group("/api")
	{
		api.POST("/shorten", linkHandler.ShortenLink)
		api.GET("/s/:short_url", linkHandler.RedirectLink)
		api.GET("/analytics/:short_url", analyticsHandler.GetAnalytics)
	}

	return e
}
