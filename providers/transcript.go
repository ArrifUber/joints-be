package providers

import (
	"github.com/go-playground/validator/v10"
	"github.com/samber/do"
	"gorm.io/gorm"

	sessionRepo "joints-be/modules/session/repository"
	transcriptController "joints-be/modules/transcript/controller"
	transcriptRepo "joints-be/modules/transcript/repository"
	transcriptService "joints-be/modules/transcript/service"
)

// ProvideTranscript registers all transcript-related dependencies into the DI injector.
func ProvideTranscript(i *do.Injector) {
	do.Provide(i, func(i *do.Injector) (transcriptRepo.TranscriptRepository, error) {
		db := do.MustInvoke[*gorm.DB](i)
		return transcriptRepo.NewTranscriptRepository(db), nil
	})

	do.Provide(i, func(i *do.Injector) (transcriptService.TranscriptService, error) {
		tRepo := do.MustInvoke[transcriptRepo.TranscriptRepository](i)
		sRepo := do.MustInvoke[sessionRepo.SessionRepository](i)
		return transcriptService.NewTranscriptService(tRepo, sRepo), nil
	})

	do.Provide(i, func(i *do.Injector) (*transcriptController.TranscriptController, error) {
		svc := do.MustInvoke[transcriptService.TranscriptService](i)
		validate := do.MustInvoke[*validator.Validate](i)
		return transcriptController.NewTranscriptController(svc, validate), nil
	})
}

