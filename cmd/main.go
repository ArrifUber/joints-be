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
	contextController "joints-be/modules/context/controller"
	contextRoutes "joints-be/modules/context/routes"
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

	// 2. Set Gin mode
	if cfg.AppEnv != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. Build DI container — order matters: shared deps first
	injector := do.New()
	providers.ProvideValidator(injector)
	providers.ProvideDB(injector)
	providers.ProvideSession(injector)
	providers.ProvideTranscript(injector)
	providers.ProvideContextEngine(injector)
	providers.ProvideContext(injector)

	// 4. Fail-fast dependency resolution
	for _, resolve := range []func() error{
		func() error { _, err := do.Invoke[*sessionController.SessionController](injector); return err },
		func() error { _, err := do.Invoke[*transcriptController.TranscriptController](injector); return err },
		func() error { _, err := do.Invoke[*contextController.ContextController](injector); return err },
	} {
		if err := resolve(); err != nil {
			log.Fatalf("[main] failed to initialize dependencies: %v", err)
		}
	}

	// 5. Setup Gin router
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		sessCtrl := do.MustInvoke[*sessionController.SessionController](injector)
		sessionRoutes.RegisterSessionRoutes(v1, sessCtrl)

		transCtrl := do.MustInvoke[*transcriptController.TranscriptController](injector)
		transcriptRoutes.RegisterTranscriptRoutes(v1, transCtrl)

		ctxCtrl := do.MustInvoke[*contextController.ContextController](injector)
		contextRoutes.RegisterContextRoutes(v1, ctxCtrl)
	}

	// 6. Start HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		log.Printf("[main] server starting on port %s (env: %s)", cfg.AppPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] server error: %v", err)
		}
	}()

	// 7. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[main] shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[main] forced shutdown: %v", err)
	}
	if err := injector.Shutdown(); err != nil {
		log.Printf("[main] DI shutdown error: %v", err)
	}
	log.Println("[main] server exited cleanly")
}
