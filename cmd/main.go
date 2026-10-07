package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"joints-be/config"
	sessionController "joints-be/modules/session/controller"
	sessionRoutes "joints-be/modules/session/routes"
	transcriptController "joints-be/modules/transcript/controller"
	transcriptRoutes "joints-be/modules/transcript/routes"
	"joints-be/providers"
)

func main() {
	// 1. Load configuration from .env
	config.Load()
	cfg := config.AppConfig

	// 2. Set Gin mode based on environment
	if cfg.AppEnv != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. Build DI container
	injector := do.New()

	// Register providers (order matters: shared deps first)
	providers.ProvideValidator(injector)
	providers.ProvideDB(injector)
	providers.ProvideSession(injector)
	providers.ProvideTranscript(injector)

	// 4. Resolve dependencies — fail fast if any provider errors
	if _, err := do.Invoke[*sessionController.SessionController](injector); err != nil {
		log.Fatalf("[main] failed to initialize session dependencies: %v", err)
	}
	if _, err := do.Invoke[*transcriptController.TranscriptController](injector); err != nil {
		log.Fatalf("[main] failed to initialize transcript dependencies: %v", err)
	}

	// 5. Setup Gin router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		sessCtrl := do.MustInvoke[*sessionController.SessionController](injector)
		sessionRoutes.RegisterSessionRoutes(v1, sessCtrl)

		transCtrl := do.MustInvoke[*transcriptController.TranscriptController](injector)
		transcriptRoutes.RegisterTranscriptRoutes(v1, transCtrl)
	}

	// 6. Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	// 7. Start server in goroutine
	go func() {
		log.Printf("[main] server starting on port %s (env: %s)", cfg.AppPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] server error: %v", err)
		}
	}()

	// 8. Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[main] shutdown signal received, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[main] server forced to shutdown: %v", err)
	}

	if err := injector.Shutdown(); err != nil {
		log.Printf("[main] error during DI container shutdown: %v", err)
	}

	log.Println("[main] server exited cleanly")
}
