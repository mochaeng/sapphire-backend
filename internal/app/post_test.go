package app

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mochaeng/sapphire-backend/internal/mocks"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetPostHandler(t *testing.T) {
	app := newTestApplication(t)
	mux := app.Mount()

	expectedPost := &models.Post{
		ID:      101,
		Tittle:  "Sample post",
		Content: "This is a sample post content",
		User:    &models.User{ID: 1, Username: "testuser"},
	}
	nonExistentID := 1
	app.Service.Post.(*mocks.MockPostService).On(
		"GetWithUser",
		mock.Anything,
		int64(101),
	).Return(expectedPost, nil)
	app.Service.Post.(*mocks.MockPostService).On(
		"GetWithUser",
		mock.Anything,
		int64(nonExistentID),
	).Return(nil, store.ErrNotFound)

	t.Run("returns status 200 for a existent post", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/post/101", nil)
		require.NoError(t, err)

		rr := executeRequest(req, mux)
		assert.Equal(t, http.StatusOK, rr.Code)

		var response struct {
			Data models.GetPostResponse `json:"data"`
		}
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err, "failed to decode response: %v", err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, expectedPost.Tittle, response.Data.Tittle)
		assert.Equal(t, expectedPost.Content, response.Data.Content)
		assert.Equal(t, expectedPost.User.Username, response.Data.User.Username)
	})

	t.Run("returns status 404 for a non-existent post", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/post/1", nil)
		require.NoError(t, err)

		rr := executeRequest(req, mux)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.JSONEq(t, `{"error":"not found"}`, rr.Body.String())
	})
}
