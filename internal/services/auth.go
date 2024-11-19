package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mochaeng/sapphire-backend/internal/auth"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/mailer"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"go.uber.org/zap"
)

var (
	ErrSetPasswordHash  = errors.New("could not set password hash")
	ErrEmailSending     = errors.New("could not send e-mail")
	ErrUserTokeCreation = errors.New("could not create user token")
)

type AuthService struct {
	store         *store.Store
	cfg           *config.Cfg
	mailer        mailer.Client
	authenticator auth.Authenticator
	logger        *zap.SugaredLogger
}

func (s *AuthService) RegisterUser(ctx context.Context, payload *models.RegisterUserPayload) (*models.UserInvitation, error) {
	if err := Validate.Struct(payload); err != nil {
		return nil, ErrInvalidPayload
	}
	user := &models.User{
		Username:  payload.Username,
		Email:     payload.Email,
		FirstName: payload.FirstName,
		Role: models.Role{
			ID: config.Roles["user"].ID,
		},
	}
	if err := user.Password.Set(payload.Password); err != nil {
		return nil, ErrSetPasswordHash
	}

	plainToken := uuid.NewString()
	sha256Token := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(sha256Token[:])
	userInvitation := &models.UserInvitation{
		User:    user,
		Token:   hashToken,
		Expired: s.cfg.Mail.Expired,
	}

	err := s.store.User.CreateAndInvite(ctx, userInvitation)
	if err != nil {
		return nil, err
	}
	userInvitation.Token = plainToken

	isSandBox := s.cfg.Env == "dev"
	activationURL := fmt.Sprintf("%s/confirm/%s", s.cfg.FrontedURL, plainToken)
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}
	status, err := s.mailer.Send(
		mailer.UserWelcomeTemplate,
		user.Username,
		user.Email,
		vars,
		isSandBox,
	)
	if err != nil {
		s.logger.Errorw("error sending welcome email", "error", err)
		if err := s.store.User.Delete(ctx, user.ID); err != nil {
			s.logger.Errorw("error deleting user", "error", err)
		}
		return nil, ErrEmailSending
	}
	s.logger.Infow("Email sent", "status code", status)
	return userInvitation, nil
}

func (s *AuthService) Authenticate(ctx context.Context, payload *models.SigninPayload) (*models.User, error) {
	if err := models.Validate.Struct(payload); err != nil {
		return nil, ErrInvalidPayload
	}

	user, err := s.store.User.GetByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		return nil, store.ErrNotFound
	}

	return user, nil
}

func (s *AuthService) ValidateToken(token string) (*jwt.Token, error) {
	return s.authenticator.ValidateToken(token)
}
