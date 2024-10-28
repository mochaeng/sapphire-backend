package app

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

func getUserFromContext(r *http.Request) *models.User {
	user, _ := r.Context().Value(userCtx).(*models.User)
	return user
}

func getAuthenticatedUserFromContext(r *http.Request) *models.User {
	user, _ := r.Context().Value(userCtx).(*models.User)
	return user
}

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		200		{object}	models.GetUserResponse
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/user/{userID} [get]
func (app *Application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	response := &models.GetUserResponse{
		UserResponse: &models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}
	if err := httpio.JsonResponse(w, http.StatusCreated, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
	}
}

// GetUserByUsername godoc
//
//	@Summary		Fetches a user
//	@Description	Fetches a user by their username
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string	true	"User username"
//	@Success		200			{object}	models.GetUserResponse
//	@Failure		400			{object}	error
//	@Failure		404			{object}	error
//	@Failure		500			{object}	error
//	@Router			/user/{username} [get]
func (app *Application) getUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		app.BadRequestResponse(w, r, httpio.ErrEmptyParam)
		return
	}
	user, err := app.Store.User.GetByUsername(r.Context(), username)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}
	response := &models.GetUserResponse{
		UserResponse: &models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}
	if err := httpio.JsonResponse(w, http.StatusCreated, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Allows a user to follow another one
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			userID	path	int	true	"User ID"
//	@Success		204		"User followed"
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		409		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/user/{userID}/follow [put]
func (app *Application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerdUser := getAuthenticatedUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	if err := app.Store.User.Follow(r.Context(), followerdUser.ID, followedID); err != nil {
		switch err {
		case store.ErrConflict:
			app.ConflictResponse(w, r, err)
		case store.ForeignKeyViolation:
			app.NotFoundResponse(w, r, err)
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}
	httpio.NoContentResponse(w)
}

// UnfollowUser godoc
//
//	@Summary		Unfollows a user
//	@Description	Allows a user to unfollow another one
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			userID	path	int	true	"User ID"
//	@Success		204		"User unfollowed"
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		409		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/user/{userID}/unfollow [put]
func (app *Application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowerUser := getAuthenticatedUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	if err := app.Store.User.Unfollow(r.Context(), unfollowerUser.ID, unfollowedID); err != nil {
		switch err {
		case store.ErrConflict:
			app.ConflictResponse(w, r, err)
		case store.ForeignKeyViolation:
			app.NotFoundResponse(w, r, err)
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}
	httpio.NoContentResponse(w)
}

// ActiveUser godoc
//
//	@Summary		Activates a user in the application
//	@Description	Activates a user by using a invitation token
//	@Tags			user
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/user/activate/{token} [put]
func (app *Application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		app.BadRequestResponse(w, r, httpio.ErrEmptyParam)
		return
	}
	err := app.Store.User.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}
	httpio.NoContentResponse(w)
}

func (app *Application) getUserWithCache(ctx context.Context, userID int64) (*models.User, error) {
	if !app.Config.Cacher.IsEnable {
		return app.Store.User.GetByID(ctx, userID)
	}
	app.Logger.Infow("cache hit", "userID", userID)
	user, err := app.CacheStore.User.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		app.Logger.Infow("fetching from the database", "id", userID)
		user, err = app.Store.User.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if err := app.CacheStore.User.Set(ctx, user); err != nil {
			return nil, err
		}
	}
	return user, nil
}
