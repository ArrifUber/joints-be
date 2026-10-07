package service

import (
	"context"

	"joints-be/modules/session"
	"joints-be/modules/session/dto"
	"joints-be/modules/session/repository"
)

type SessionService interface {
	CreateSession(ctx context.Context, req dto.CreateSessionRequest) (*dto.SessionResponse, error)
	GetSessionByID(ctx context.Context, id string) (*dto.SessionResponse, error)
}

type sessionServiceImpl struct {
	repo repository.SessionRepository
}

func NewSessionService(repo repository.SessionRepository) SessionService {
	return &sessionServiceImpl{repo: repo}
}

func (s *sessionServiceImpl) CreateSession(ctx context.Context, req dto.CreateSessionRequest) (*dto.SessionResponse, error) {
	newSession := &session.Session{
		Subject:  req.Subject,
		Topic:    req.Topic,
		Subtopic: req.Subtopic,
		Status:   "active",
	}

	if err := s.repo.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return &dto.SessionResponse{
		ID:        newSession.ID,
		Subject:   newSession.Subject,
		Topic:     newSession.Topic,
		Subtopic:  newSession.Subtopic,
		Status:    newSession.Status,
		CreatedAt: newSession.CreatedAt,
		UpdatedAt: newSession.UpdatedAt,
	}, nil
}

func (s *sessionServiceImpl) GetSessionByID(ctx context.Context, id string) (*dto.SessionResponse, error) {
	sess, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.SessionResponse{
		ID:        sess.ID,
		Subject:   sess.Subject,
		Topic:     sess.Topic,
		Subtopic:  sess.Subtopic,
		Status:    sess.Status,
		CreatedAt: sess.CreatedAt,
		UpdatedAt: sess.UpdatedAt,
	}, nil
}

