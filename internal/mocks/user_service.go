package mocks

import (
	"context"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Follow(ctx context.Context, followerID int64, followedID int64) error {
	args := m.Called(ctx, followerID, followedID)
	return args.Error(0)
}

func (m *MockUserService) Unfollow(ctx context.Context, unfollowerID int64, unfollowedID int64) error {
	args := m.Called(ctx, unfollowerID, unfollowedID)
	return args.Error(0)
}

func (m *MockUserService) Activate(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetCached(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetProfile(ctx context.Context, username string) (*models.UserProfile, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.UserProfile), args.Error(1)
}

func (m *MockUserService) GetPosts(ctx context.Context, username string, cursor time.Time, limit int) ([]*models.Post, error) {
	args := m.Called(ctx, username, cursor, limit)
	return args.Get(0).([]*models.Post), args.Error(1)
}
