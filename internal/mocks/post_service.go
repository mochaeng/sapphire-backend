package mocks

import (
	"context"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/models/payloads"
	"github.com/stretchr/testify/mock"
)

type MockPostService struct {
	mock.Mock
}

func (m *MockPostService) Create(ctx context.Context, user *models.User, payload *payloads.CreatePostDataValuesPayload, file []byte) (*models.Post, error) {
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

func (m *MockPostService) Update(ctx context.Context, post *models.Post, payload *payloads.UpdatePostPayload) error {
	args := m.Called(ctx, post, payload)
	return args.Error(0)
}
