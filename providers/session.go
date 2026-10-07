package providers

import (
	"github.com/go-playground/validator/v10"
	"github.com/samber/do"
	"gorm.io/gorm"

	sessionController "joints-be/modules/session/controller"
	sessionRepo "joints-be/modules/session/repository"
	sessionService "joints-be/modules/session/service"
)

// ProvideSession registers all session-related dependencies into the DI injector.
func ProvideSession(i *do.Injector) {
	do.Provide(i, func(i *do.Injector) (sessionRepo.SessionRepository, error) {
		db := do.MustInvoke[*gorm.DB](i)
		return sessionRepo.NewSessionRepository(db), nil
	})

	do.Provide(i, func(i *do.Injector) (sessionService.SessionService, error) {
		repo := do.MustInvoke[sessionRepo.SessionRepository](i)
		return sessionService.NewSessionService(repo), nil
	})

	do.Provide(i, func(i *do.Injector) (*sessionController.SessionController, error) {
		svc := do.MustInvoke[sessionService.SessionService](i)
		validate := do.MustInvoke[*validator.Validate](i)
		return sessionController.NewSessionController(svc, validate), nil
	})
}

// ProvideValidator registers *validator.Validate into the DI injector.
func ProvideValidator(i *do.Injector) {
	do.Provide(i, func(i *do.Injector) (*validator.Validate, error) {
		return validator.New(), nil
	})
}

