package app

import (
	"net/http"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/models/pagination"
	"github.com/mochaeng/sapphire-backend/internal/models/responses"
	service "github.com/mochaeng/sapphire-backend/internal/services"
)

// GetUserFeed godoc
//
//	@Summary		Gets the user feed
//	@Description	A feed contains the user own's posts and the ones their follow
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		string	false	"Limit"
//	@Param			cursor	query		string	false	"Cursor"
//	@Success		200		{object}	[]models.PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/user/feed [get]
func (app *Application) GetUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// feedQuery := pagination.PaginateFeedQuery{
	// 	Limit:  20,
	// 	Cursor: 0,
	// 	Sort:   "desc",
	// }
	// if err := feedQuery.Parse(r); err != nil {
	// 	app.BadRequestResponse(w, r, err)
	// 	return
	// }
	query := r.URL.Query()
	limitParam := query.Get("limit")
	cursorParam := query.Get("cursor")

	feedQuery := pagination.PaginateFeedQuery{}
	feedQuery.Parse(limitParam, cursorParam)

	user := getUserFromContext(r)
	if user == nil {
		app.UnauthorizedErrorResponse(w, r, nil)
		return
	}

	posts, err := app.Service.Feed.Get(r.Context(), user.ID, &feedQuery)
	if err != nil {
		switch err {
		case service.ErrInvalidPayload:
			app.BadRequestResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}

	var response responses.FeedResponse
	response.Posts = make([]responses.PostResponse, len(posts))
	response.NextCursor = feedQuery.NextCursor

	for idx, post := range posts {
		response.Posts[idx] = responses.PostResponse{
			ID:        post.ID,
			Content:   post.Content,
			MediaURL:  post.Media.String,
			CreatedAt: post.CreatedAt,
			User: &responses.UserResponse{
				ID:        post.User.ID,
				Username:  post.User.Username,
				FirstName: post.User.FirstName,
				LastName:  post.User.LastName,
			},
		}
	}

	if err := httpio.JsonResponse(w, http.StatusOK, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
	}
}
