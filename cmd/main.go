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
	contextWorker "joints-be/modules/context/worker"
	sessionController "joints-be/modules/session/controller"
	sessionRoutes "joints-be/modules/session/routes"
	transcriptController "joints-be/modules/transcript/controller"
	transcriptRoutes "joints-be/modules/transcript/routes"
	wsPkg "joints-be/pkg/websocket"
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
	providers.ProvideWebSocket(injector) // Provides Hub, EventPublisher, and wsHandler
	providers.ProvideSession(injector)
	providers.ProvideTranscript(injector)
	providers.ProvideContextEngine(injector)
	providers.ProvideContext(injector)

	// 4. Fail-fast dependency resolution
	for _, resolve := range []func() error{
		func() error { _, err := do.Invoke[*sessionController.SessionController](injector); return err },
		func() error { _, err := do.Invoke[*transcriptController.TranscriptController](injector); return err },
		func() error { _, err := do.Invoke[*contextController.ContextController](injector); return err },
		func() error { _, err := do.Invoke[*contextWorker.ContextWorker](injector); return err },
		func() error { _, err := do.Invoke[*wsPkg.Handler](injector); return err },
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

	// WebSocket endpoint for realtime context streaming (§16, §24)
	wsHandler := do.MustInvoke[*wsPkg.Handler](injector)
	router.GET("/ws/sessions/:id", wsHandler.HandleConnection)

	// API v1 group
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

	// 7. Start Context Background Worker
	worker := do.MustInvoke[*contextWorker.ContextWorker](injector)
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	go worker.Start(workerCtx)

	// 8. Graceful shutdown sequence (§43):
	// SIGINT/SIGTERM -> Stop HTTP -> Stop Worker -> Close WebSocket -> Close DB -> Exit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[main] shutting down gracefully...")

	// 8.1 Stop accepting new HTTP requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[main] forced HTTP shutdown: %v", err)
	}

	// 8.2 Stop background worker
	log.Println("[main] stopping background worker...")
	cancelWorker()

	// 8.3 Close WebSocket connections
	log.Println("[main] closing websocket connections...")
	wsHub := do.MustInvoke[wsPkg.Hub](injector)
	wsHub.Close()

	// 8.4 Close DB and other DI resources
	if err := injector.Shutdown(); err != nil {
		log.Printf("[main] DI shutdown error: %v", err)
	}

	log.Println("[main] server exited cleanly")
}
