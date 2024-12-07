package mocks

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/models"
	service "github.com/mochaeng/sapphire-backend/internal/services"
	"github.com/stretchr/testify/mock"
)

func NewMockService() service.Service {
	return service.Service{
		User: &MockUserService{},
		Post: &MockPostService{},
	}
}

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

type MockPostService struct {
	mock.Mock
}

func (m *MockPostService) Create(ctx context.Context, user *models.User, payload *models.CreatePostPayload, file []byte) (*models.Post, error) {
	args := m.Called(ctx, user, payload, file)
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostService) GetWithUser(ctx context.Context, postID int64) (*models.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostService) Delete(ctx context.Context, postID int64) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostService) Update(ctx context.Context, post *models.Post, payload *models.UpdatePostPayload) error {
	args := m.Called(ctx, post, payload)
	return args.Error(0)
}
