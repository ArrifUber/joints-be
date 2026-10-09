package service_test

import (
	"context"
	"errors"
	"testing"

	sessionEntity "joints-be/modules/session/entity"
	"joints-be/modules/transcript/dto"
	transcriptEntity "joints-be/modules/transcript/entity"
	"joints-be/modules/transcript/repository"
	"joints-be/modules/transcript/service"
)

type mockSessionRepo struct {
	exists map[string]bool
}

func (m *mockSessionRepo) Create(ctx context.Context, s *sessionEntity.Session) error {
	m.exists[s.ID] = true
	return nil
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id string) (*sessionEntity.Session, error) {
	if !m.exists[id] {
		return nil, errors.New("not found")
	}
	return &sessionEntity.Session{ID: id, Status: sessionEntity.SessionStatusActive}, nil
}

type mockTranscriptRepo struct {
	chunks []transcriptEntity.TranscriptChunk
}

func newMockTranscriptRepo() *mockTranscriptRepo {
	return &mockTranscriptRepo{chunks: make([]transcriptEntity.TranscriptChunk, 0)}
}

func (m *mockTranscriptRepo) Create(ctx context.Context, t *transcriptEntity.TranscriptChunk) error {
	for _, c := range m.chunks {
		if c.SessionID == t.SessionID && c.Sequence == t.Sequence {
			return repository.ErrDuplicateSequence
		}
	}
	t.ID = "chunk-uuid"
	m.chunks = append(m.chunks, *t)
	return nil
}

func (m *mockTranscriptRepo) FindPending(ctx context.Context, limit int) ([]transcriptEntity.TranscriptChunk, error) {
	var result []transcriptEntity.TranscriptChunk
	for _, c := range m.chunks {
		if c.Status == transcriptEntity.TranscriptStatusPending {
			result = append(result, c)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockTranscriptRepo) FindBySessionID(ctx context.Context, sessionID string) ([]transcriptEntity.TranscriptChunk, error) {
	var result []transcriptEntity.TranscriptChunk
	for _, c := range m.chunks {
		if c.SessionID == sessionID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockTranscriptRepo) FindRecentBySession(ctx context.Context, sessionID string, upToSequence int64, limit int) ([]transcriptEntity.TranscriptChunk, error) {
	var result []transcriptEntity.TranscriptChunk
	for _, c := range m.chunks {
		if c.SessionID == sessionID && c.Sequence <= upToSequence {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockTranscriptRepo) UpdateStatus(ctx context.Context, id string, status transcriptEntity.TranscriptStatus) error {
	for i, c := range m.chunks {
		if c.ID == id {
			m.chunks[i].Status = status
			return nil
		}
	}
	return nil
}

func TestTranscriptService_AcceptTranscript(t *testing.T) {
	sRepo := &mockSessionRepo{exists: map[string]bool{"sess-1": true}}
	tRepo := newMockTranscriptRepo()
	svc := service.NewTranscriptService(tRepo, sRepo)

	// 1. Success Acceptance (202 response data)
	res, err := svc.AcceptTranscript(context.Background(), "sess-1", dto.CreateTranscriptRequest{
		Sequence: 1,
		Text:     "Hari ini kita akan belajar dinamika dan hukum Newton.",
	})
	if err != nil {
		t.Fatalf("expected transcript to be accepted, got error: %v", err)
	}
	if !res.Accepted || res.Sequence != 1 {
		t.Errorf("expected accepted=true and sequence=1, got %+v", res)
	}

	// 2. Scenario 4 (§38): Duplicate Sequence returns ErrDuplicateSequence
	_, err = svc.AcceptTranscript(context.Background(), "sess-1", dto.CreateTranscriptRequest{
		Sequence: 1,
		Text:     "Teks duplikat",
	})
	if !errors.Is(err, service.ErrDuplicateSequence) {
		t.Fatalf("expected ErrDuplicateSequence on duplicate sequence, got: %v", err)
	}

	// 3. Unknown Session returns ErrSessionNotFound
	_, err = svc.AcceptTranscript(context.Background(), "non-existent-session", dto.CreateTranscriptRequest{
		Sequence: 2,
		Text:     "Teks sesi tidak dikenal",
	})
	if !errors.Is(err, service.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound for unknown session, got: %v", err)
	}
}

