package services

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"

	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/media"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"go.uber.org/zap"
)

var ErrInvalidPayload = errors.New("invalid payload")
var ErrSaveFile = errors.New("not possible to save the file")

type PostService struct {
	store  *store.Store
	cfg    *config.Cfg
	logger *zap.SugaredLogger
}

func (s *PostService) Create(ctx context.Context, user *models.User, payload *models.CreatePostPayload, file []byte) (*models.Post, error) {
	if err := Validate.Struct(payload); err != nil {
		return nil, ErrInvalidPayload
	}
	fileURL := ""
	if file != nil {
		filename, err := media.SaveFileToServer(file, s.cfg.MediaFolder)
		if err != nil {
			return nil, ErrSaveFile
		}
		fileURL = filepath.Join(s.cfg.MediaFolder, filename)
	}
	post := &models.Post{
		Tittle:  payload.Tittle,
		Content: payload.Content,
		Media: sql.NullString{
			String: fileURL,
			Valid:  fileURL != "",
		},
		Tags: payload.Tags,
		User: user,
	}
	if err := s.store.Post.Create(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *PostService) Update(ctx context.Context, post *models.Post, payload *models.UpdatePostPayload) error {
	if err := Validate.Struct(payload); err != nil {
		return ErrInvalidPayload
	}
	if payload.Content != "" {
		post.Content = payload.Content
	}
	if payload.Tittle != "" {
		post.Tittle = payload.Tittle
	}
	return s.store.Post.UpdateByID(ctx, post)
}

func (s *PostService) GetWithUser(ctx context.Context, postID int64) (*models.Post, error) {
	return s.store.Post.GetByIDWithUser(ctx, postID)
}

func (s *PostService) Delete(ctx context.Context, postID int64) error {
	return s.store.Post.DeleteByID(ctx, postID)
}
