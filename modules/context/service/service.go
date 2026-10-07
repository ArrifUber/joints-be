package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"gorm.io/gorm"
	"joints-be/modules/context/dto"
	contextEntity "joints-be/modules/context/entity"
	"joints-be/modules/context/repository"
	sessionRepo "joints-be/modules/session/repository"
	transcriptRepo "joints-be/modules/transcript/repository"
	"joints-be/pkg/contextengine"
)

type ContextService interface {
	// ProcessTranscript membangun context dari transcript chunk yang diberikan.
	// Dipanggil oleh background worker — bukan HTTP handler.
	ProcessTranscript(ctx context.Context, sessionID string, transcriptID string, sequence int64, texts []string) (*contextEntity.ContextChunk, error)

	// ListBySession mengembalikan semua context untuk satu session.
	ListBySession(ctx context.Context, sessionID string) ([]dto.ContextResponse, error)
}

type contextServiceImpl struct {
	contextRepo    repository.ContextRepository
	sessionRepo    sessionRepo.SessionRepository
	transcriptRepo transcriptRepo.TranscriptRepository
	engine         contextengine.ContextEngine
}

func NewContextService(
	contextRepo repository.ContextRepository,
	sessionRepo sessionRepo.SessionRepository,
	transcriptRepo transcriptRepo.TranscriptRepository,
	engine contextengine.ContextEngine,
) ContextService {
	return &contextServiceImpl{
		contextRepo:    contextRepo,
		sessionRepo:    sessionRepo,
		transcriptRepo: transcriptRepo,
		engine:         engine,
	}
}

func (s *contextServiceImpl) ProcessTranscript(
	ctx context.Context,
	sessionID string,
	transcriptID string,
	sequence int64,
	texts []string,
) (*contextEntity.ContextChunk, error) {
	// 1. Ambil metadata session (subject, topic, subtopic)
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 2. Bangun "current context" dari context chunk terakhir
	var currentResult *contextengine.ContextResult
	latest, err := s.contextRepo.FindLatestBySessionID(ctx, sessionID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[context] failed to get latest context session_id=%s error=%v", sessionID, err)
		// Non-fatal: lanjutkan tanpa current context
	}
	if latest != nil {
		var parsed contextengine.ContextResult
		if err := json.Unmarshal(latest.Data, &parsed); err == nil {
			currentResult = &parsed
		}
	}

	// 3. Kirim ke AI engine untuk dianalisis
	input := contextengine.ContextInput{
		Subject:           sess.Subject,
		Topic:             sess.Topic,
		Subtopic:          sess.Subtopic,
		CurrentContext:    currentResult,
		RecentTranscripts: texts,
	}

	result, err := s.engine.Analyze(ctx, input)
	if err != nil {
		return nil, err
	}

	// 4. Jika no_change, tidak perlu simpan ke DB
	if result.Type == "no_change" {
		log.Printf("[context] no_change session_id=%s sequence=%d", sessionID, sequence)
		return nil, nil //nolint:nilnil
	}

	// 5. Encode result ke JSON untuk disimpan
	dataJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	chunk := &contextEntity.ContextChunk{
		SessionID: sessionID,
		Sequence:  sequence,
		Type:      result.Type,
		Data:      dataJSON,
	}

	if err := s.contextRepo.Create(ctx, chunk); err != nil {
		return nil, err
	}

	log.Printf("[context] saved session_id=%s sequence=%d type=%s", sessionID, sequence, result.Type)
	return chunk, nil
}

func (s *contextServiceImpl) ListBySession(ctx context.Context, sessionID string) ([]dto.ContextResponse, error) {
	chunks, err := s.contextRepo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ContextResponse, len(chunks))
	for i, c := range chunks {
		responses[i] = dto.ContextResponse{
			ID:        c.ID,
			SessionID: c.SessionID,
			Sequence:  c.Sequence,
			Type:      c.Type,
			Data:      c.Data,
			CreatedAt: c.CreatedAt,
		}
	}
	return responses, nil
}

