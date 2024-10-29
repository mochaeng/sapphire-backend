package app

import (
	"net/http"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/models"
	service "github.com/mochaeng/sapphire-backend/internal/services"
)

// GetUserFeed godoc
//
//	@Summary		Gets the user feed
//	@Description	A feed contains the user own's posts and the ones their follow
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//	@Param			limit	query		string	false	"Limit"
//	@Param			offset	query		string	false	"Offset"
//	@Param			sort	query		string	false	"Offset"
//	@Param			tags	query		string	false	"Offset"
//	@Param			search	query		string	false	"Offset"
//	@Success		200		{object}	[]models.PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/user/feed [get]
func (app *Application) GetUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	feedQuery := models.PaginateFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	if err := feedQuery.Parse(r); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	feed, err := app.Service.Feed.Get(r.Context(), user.ID, &feedQuery)
	if err != nil {
		switch err {
		case service.ErrInvalidPayload:
			app.BadRequestResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}
	if err := httpio.JsonResponse(w, http.StatusOK, feed); err != nil {
		app.InternalServerErrorResponse(w, r, err)
	}
}
