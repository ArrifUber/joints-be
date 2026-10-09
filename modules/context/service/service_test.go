package service_test

import (
	"context"
	"errors"
	"testing"

	contextEntity "joints-be/modules/context/entity"
	"joints-be/modules/context/service"
	sessionEntity "joints-be/modules/session/entity"
	transcriptEntity "joints-be/modules/transcript/entity"
	"joints-be/pkg/contextengine"
	"joints-be/pkg/event"
)

type mockContextRepo struct {
	chunks []contextEntity.ContextChunk
}

func (m *mockContextRepo) Create(ctx context.Context, c *contextEntity.ContextChunk) error {
	c.ID = "ctx-chunk-uuid"
	m.chunks = append(m.chunks, *c)
	return nil
}

func (m *mockContextRepo) FindBySessionID(ctx context.Context, sessionID string) ([]contextEntity.ContextChunk, error) {
	return m.chunks, nil
}

func (m *mockContextRepo) FindLatestBySessionID(ctx context.Context, sessionID string) (*contextEntity.ContextChunk, error) {
	if len(m.chunks) == 0 {
		return nil, errors.New("record not found")
	}
	return &m.chunks[len(m.chunks)-1], nil
}

type mockSessionRepoContext struct{}

func (m *mockSessionRepoContext) Create(ctx context.Context, s *sessionEntity.Session) error {
	return nil
}

func (m *mockSessionRepoContext) GetByID(ctx context.Context, id string) (*sessionEntity.Session, error) {
	return &sessionEntity.Session{
		ID:       id,
		Subject:  "Fisika",
		Topic:    "Dinamika",
		Subtopic: "Hukum Newton",
		Status:   sessionEntity.SessionStatusActive,
	}, nil
}

type mockTranscriptRepoContext struct{}

func (m *mockTranscriptRepoContext) Create(ctx context.Context, t *transcriptEntity.TranscriptChunk) error {
	return nil
}

func (m *mockTranscriptRepoContext) FindPending(ctx context.Context, limit int) ([]transcriptEntity.TranscriptChunk, error) {
	return nil, nil
}

func (m *mockTranscriptRepoContext) FindBySessionID(ctx context.Context, sessionID string) ([]transcriptEntity.TranscriptChunk, error) {
	return nil, nil
}

func (m *mockTranscriptRepoContext) FindRecentBySession(ctx context.Context, sessionID string, upToSequence int64, limit int) ([]transcriptEntity.TranscriptChunk, error) {
	return nil, nil
}

func (m *mockTranscriptRepoContext) UpdateStatus(ctx context.Context, id string, status transcriptEntity.TranscriptStatus) error {
	return nil
}

type mockEventPublisher struct {
	published []event.Event
}

func (m *mockEventPublisher) Publish(ctx context.Context, sessionID string, ev event.Event) error {
	m.published = append(m.published, ev)
	return nil
}

type failingContextEngine struct{}

func (f *failingContextEngine) Analyze(ctx context.Context, input contextengine.ContextInput) (*contextengine.ContextResult, error) {
	return nil, errors.New("ai provider timeout")
}

func TestContextService_ProcessTranscript(t *testing.T) {
	cRepo := &mockContextRepo{}
	sRepo := &mockSessionRepoContext{}
	tRepo := &mockTranscriptRepoContext{}
	engine := contextengine.NewMockContextEngine()
	publisher := &mockEventPublisher{}

	svc := service.NewContextService(cRepo, sRepo, tRepo, engine, publisher)

	// 1. Process valid transcript chunk
	res, err := svc.ProcessTranscript(
		context.Background(),
		"sess-1",
		"trans-1",
		1,
		[]string{"Hari ini kita membahas rumus Hukum Newton F = m x a."},
	)
	if err != nil {
		t.Fatalf("expected successful context processing, got: %v", err)
	}
	if res == nil {
		t.Fatalf("expected context chunk to be created")
	}
	if len(cRepo.chunks) != 1 {
		t.Errorf("expected 1 context chunk in repository, got %d", len(cRepo.chunks))
	}
	if len(publisher.published) != 1 {
		t.Errorf("expected 1 event published, got %d", len(publisher.published))
	}
	if publisher.published[0].Event != "context_update" {
		t.Errorf("expected event type 'context_update', got '%s'", publisher.published[0].Event)
	}

	// 2. Process with empty input -> "no_change" (§23: Semantic Context Update)
	resNoChange, err := svc.ProcessTranscript(
		context.Background(),
		"sess-1",
		"trans-2",
		2,
		[]string{}, // Empty input results in no_change
	)
	if err != nil {
		t.Fatalf("expected no error for no_change, got: %v", err)
	}
	if resNoChange != nil {
		t.Errorf("expected nil result on no_change, got %+v", resNoChange)
	}
	// Repository and published event count should NOT increase
	if len(cRepo.chunks) != 1 {
		t.Errorf("expected context chunk count to stay 1, got %d", len(cRepo.chunks))
	}
	if len(publisher.published) != 1 {
		t.Errorf("expected published event count to stay 1, got %d", len(publisher.published))
	}

	// 3. AI Processing failure scenario (§38 Scenario 5)
	svcWithFailEngine := service.NewContextService(cRepo, sRepo, tRepo, &failingContextEngine{}, publisher)
	_, err = svcWithFailEngine.ProcessTranscript(
		context.Background(),
		"sess-1",
		"trans-3",
		3,
		[]string{"Teks apapun"},
	)
	if err == nil {
		t.Fatalf("expected error from failing engine, got nil")
	}
}

