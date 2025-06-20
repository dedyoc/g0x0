package server

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dedyoc/g0x0/internal/config"
	"github.com/dedyoc/g0x0/internal/features/files"
)

type Server struct {
	echo   *echo.Echo
	config *config.Config
	db     *sql.DB
}

func New(cfg *config.Config, db *sql.DB) *Server {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(
		middleware.CORSConfig{
			AllowOrigins: []string{"*", "http://localhost:8080", "http://localhost:3000"},
		},
	))

	// Static files
	e.Static("/static", "web/static")

	s := &Server{
		echo:   e,
		config: cfg,
		db:     db,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// File handling routes
	fileHandler := files.NewHandler(s.db, s.config)
	s.echo.GET("/", fileHandler.Index)
	s.echo.POST("/", fileHandler.Upload)
	s.echo.GET("/:id", fileHandler.Get)
	s.echo.POST("/:id", fileHandler.Manage)
	s.echo.GET("/s/:secret/:id", fileHandler.GetSecret)

	// URL shortening routes
	// urlHandler := urls.NewHandler(s.db, s.config)
	// s.echo.POST("/shorten", urlHandler.Shorten)

	// Health check
	s.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
}

func (s *Server) Start(address string) error {
	return s.echo.Start(address)
}
