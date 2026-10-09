package service_test

import (
	"context"
	"errors"
	"testing"

	"joints-be/modules/session/dto"
	"joints-be/modules/session/entity"
	"joints-be/modules/session/service"
)

type mockSessionRepo struct {
	sessions map[string]*entity.Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{sessions: make(map[string]*entity.Session)}
}

func (m *mockSessionRepo) Create(ctx context.Context, s *entity.Session) error {
	if s.ID == "" {
		s.ID = "test-session-uuid"
	}
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return s, nil
}

func TestSessionService_CreateSession(t *testing.T) {
	repo := newMockSessionRepo()
	svc := service.NewSessionService(repo)

	req := dto.CreateSessionRequest{
		Subject:  "Fisika",
		Topic:    "Dinamika",
		Subtopic: "Hukum Newton",
	}

	res, err := svc.CreateSession(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.ID == "" {
		t.Errorf("expected session ID to be generated")
	}
	if res.Subject != req.Subject || res.Topic != req.Topic || res.Subtopic != req.Subtopic {
		t.Errorf("expected subject/topic/subtopic to match request")
	}
	if res.Status != string(entity.SessionStatusActive) {
		t.Errorf("expected status 'active', got %s", res.Status)
	}
}

func TestSessionService_GetSessionByID(t *testing.T) {
	repo := newMockSessionRepo()
	svc := service.NewSessionService(repo)

	// Test Not Found
	_, err := svc.GetSessionByID(context.Background(), "unknown-id")
	if err == nil {
		t.Errorf("expected error for unknown ID, got nil")
	}

	// Create and Get
	created, _ := svc.CreateSession(context.Background(), dto.CreateSessionRequest{
		Subject:  "Kimia",
		Topic:    "Ikatan Kimia",
		Subtopic: "Ikatan Kovalen",
	})

	found, err := svc.GetSessionByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected session to be found, got %v", err)
	}
	if found.ID != created.ID || found.Subject != "Kimia" {
		t.Errorf("retrieved session does not match created session")
	}
}

