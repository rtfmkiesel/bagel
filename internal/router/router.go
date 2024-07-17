package router

import (
	"bagel/internal/logger"
	"context"
	"embed"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	static "github.com/soulteary/gin-static"
	"gorm.io/gorm"
)

var (
	//go:embed static
	EmbedFSStatic embed.FS

	//go:embed templates
	EmbedFSTemplates embed.FS

	srv *http.Server
	db  *gorm.DB
)

const (
	// Default address to listen on
	defaultAddr = ":8080"
)

// Start starts the router
func Start(database *gorm.DB) {
	db = database

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), Logger(), ErrorHandler())

	// Add custom functions to the template
	funcMaps := template.FuncMap{}

	// Load the templates from the embedded filesystem
	templ := template.New("").Funcs(funcMaps)
	templ, err := templ.ParseFS(EmbedFSTemplates, "templates/*.tmpl")
	if err != nil {
		logger.Fatal(err)
	}
	r.SetHTMLTemplate(templ)

	// Register the routes
	r.GET("/", listScans)
	r.POST("/scan/new", newScan)
	r.GET("/scan/:id", getScan)
	r.GET("/scan/:id/json", getScanJSON)
	r.DELETE("/scan/:id", deleteScan)

	r.Use(static.ServeEmbed("", EmbedFSStatic))

	addr := defaultAddr
	if os.Getenv("INSIDETHEMATRIX") != "true" {
		// Only listen on localhost if we're not in a container
		addr = "127.0.0.1:8080"
	}

	srv = &http.Server{
		Addr:              addr,
		Handler:           r.Handler(),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	go func() {
		logger.Info("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()
}

// Stop stop the router via a graceful shutdown
func Stop() error {
	logger.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	logger.Info("Server stopped")
	return nil
}
