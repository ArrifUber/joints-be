package service

import (
	"context"
	"errors"
	"log"

	"joints-be/modules/transcript/dto"
	"joints-be/modules/transcript/entity"
	"joints-be/modules/transcript/repository"
	sessionRepo "joints-be/modules/session/repository"
)

// ErrSessionNotFound is returned when the session does not exist.
var ErrSessionNotFound = errors.New("session not found")

// ErrDuplicateSequence is re-exported for the controller layer.
var ErrDuplicateSequence = repository.ErrDuplicateSequence

type TranscriptService interface {
	// AcceptTranscript validates the session, saves the transcript as pending,
	// and returns immediately — never waits for AI processing.
	AcceptTranscript(ctx context.Context, sessionID string, req dto.CreateTranscriptRequest) (*dto.CreateTranscriptResponse, error)

	// ListBySession returns all transcript chunks for a session ordered by sequence.
	ListBySession(ctx context.Context, sessionID string) ([]entity.TranscriptChunk, error)
}

type transcriptServiceImpl struct {
	transcriptRepo repository.TranscriptRepository
	sessionRepo    sessionRepo.SessionRepository
}

func NewTranscriptService(
	transcriptRepo repository.TranscriptRepository,
	sessionRepo sessionRepo.SessionRepository,
) TranscriptService {
	return &transcriptServiceImpl{
		transcriptRepo: transcriptRepo,
		sessionRepo:    sessionRepo,
	}
}

func (s *transcriptServiceImpl) AcceptTranscript(ctx context.Context, sessionID string, req dto.CreateTranscriptRequest) (*dto.CreateTranscriptResponse, error) {
	// Validate that the session exists before accepting transcript
	_, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	chunk := &entity.TranscriptChunk{
		SessionID: sessionID,
		Sequence:  req.Sequence,
		Text:      req.Text,
		Status:    entity.TranscriptStatusPending,
	}

	if err := s.transcriptRepo.Create(ctx, chunk); err != nil {
		if errors.Is(err, repository.ErrDuplicateSequence) {
			return nil, ErrDuplicateSequence
		}
		log.Printf("[transcript] failed to save chunk session_id=%s sequence=%d error=%v", sessionID, req.Sequence, err)
		return nil, err
	}

	log.Printf("[transcript] accepted session_id=%s sequence=%d id=%s", sessionID, req.Sequence, chunk.ID)

	return &dto.CreateTranscriptResponse{
		Accepted: true,
		Sequence: req.Sequence,
	}, nil
}

func (s *transcriptServiceImpl) ListBySession(ctx context.Context, sessionID string) ([]entity.TranscriptChunk, error) {
	return s.transcriptRepo.FindBySessionID(ctx, sessionID)
}

