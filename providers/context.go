package providers

import (
	"github.com/samber/do"
	"gorm.io/gorm"

	contextController "joints-be/modules/context/controller"
	contextRepo "joints-be/modules/context/repository"
	contextService "joints-be/modules/context/service"
	sessionRepo "joints-be/modules/session/repository"
	transcriptRepo "joints-be/modules/transcript/repository"
	"joints-be/pkg/contextengine"
)

// ProvideContextEngine registers the ContextEngine implementation.
// Switch AI_PROVIDER in .env to change the provider (e.g. "mock", "gemini").
func ProvideContextEngine(i *do.Injector) {
	do.Provide(i, func(i *do.Injector) (contextengine.ContextEngine, error) {
		// cfg := config.AppConfig
		// switch cfg.AIProvider {
		// case "gemini": return gemini.New(cfg.AIAPIKey), nil
		// default:
		return contextengine.NewMockContextEngine(), nil
		// }
	})
}

// ProvideContext registers all context-related dependencies into the DI injector.
func ProvideContext(i *do.Injector) {
	do.Provide(i, func(i *do.Injector) (contextRepo.ContextRepository, error) {
		db := do.MustInvoke[*gorm.DB](i)
		return contextRepo.NewContextRepository(db), nil
	})

	do.Provide(i, func(i *do.Injector) (contextService.ContextService, error) {
		cRepo := do.MustInvoke[contextRepo.ContextRepository](i)
		sRepo := do.MustInvoke[sessionRepo.SessionRepository](i)
		tRepo := do.MustInvoke[transcriptRepo.TranscriptRepository](i)
		engine := do.MustInvoke[contextengine.ContextEngine](i)
		return contextService.NewContextService(cRepo, sRepo, tRepo, engine), nil
	})

	do.Provide(i, func(i *do.Injector) (*contextController.ContextController, error) {
		svc := do.MustInvoke[contextService.ContextService](i)
		return contextController.NewContextController(svc), nil
	})
}

