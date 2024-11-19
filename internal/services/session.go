package services

import (
	"context"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/cryptoutils"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"go.uber.org/zap"
)

const sessionExpiresIn = 30 * 24 * time.Hour

type SessionService struct {
	store  *store.Store
	cfg    *config.Cfg
	logger *zap.SugaredLogger
}

func (s *SessionService) GenerateSessionToken() (string, error) {
	return cryptoutils.GenerateRandomString(20)
}

func (s *SessionService) CreateSession(token string, userID int64) (*models.Session, error) {
	sessionID := cryptoutils.GetSessionID(token)
	expiresAt := time.Now().Add(sessionExpiresIn)
	session := models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	err := s.store.Session.Create(context.Background(), &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionService) ValidateSessionToken(token string) (*models.Session, error) {
	sessionID := cryptoutils.GetSessionID(token)
	ctx := context.Background()
	session, err := s.store.Session.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if time.Now().After(session.ExpiresAt) {
		err := s.store.Session.Delete(ctx, session.ID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	if time.Now().After(session.ExpiresAt.Add(-sessionExpiresIn / 2)) {
		session.ExpiresAt = time.Now().Add(sessionExpiresIn)
		_ = s.store.Session.UpdateExpires(ctx, session)
	}
	return session, nil
}

func (s *SessionService) InvalidateSession(sessionID string) error {
	err := s.store.Session.Delete(context.Background(), sessionID)
	if err != nil {
		return err
	}
	return nil
}
