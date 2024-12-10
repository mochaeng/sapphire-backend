package app

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/models/responses"
	"github.com/mochaeng/sapphire-backend/internal/services"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

// GetUser godoc
//
//	@Summary		Fetches a user
//	@Description	Fetches by ID users that are already activated in the system.
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
	response := &responses.GetUserResponse{
		UserResponse: &responses.UserResponse{
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
	user, err := app.Service.User.GetByUsername(r.Context(), username)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}
	response := &responses.GetUserResponse{
		UserResponse: &responses.UserResponse{
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
	followerdUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	if err := app.Service.User.Follow(r.Context(), followerdUser.ID, followedID); err != nil {
		switch err {
		case store.ErrConflict:
			app.ConflictResponse(w, r, err)
		case store.ErrForeignKeyViolation, store.ErrNotFound:
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
	unfollowerUser := getUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	if err := app.Service.User.Unfollow(r.Context(), unfollowerUser.ID, unfollowedID); err != nil {
		switch err {
		case store.ErrConflict:
			app.ConflictResponse(w, r, err)
		case store.ErrForeignKeyViolation, store.ErrNotFound:
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
//	@Router			/verify-email/{token} [put]
func (app *Application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		app.BadRequestResponse(w, r, httpio.ErrEmptyParam)
		return
	}
	err := app.Service.User.Activate(r.Context(), token)
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

// GetUserProfileByUsername godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by their username
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string	true	"User username"
//	@Success		200			{object}	models.GetUserResponse
//	@Failure		400			{object}	error
//	@Failure		404			{object}	error
//	@Failure		500			{object}	error
//	@Router			/user/profile/{username} [get]
func (app *Application) getUserProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		app.BadRequestResponse(w, r, httpio.ErrEmptyParam)
		return
	}
	userProfile, err := app.Service.User.GetProfile(r.Context(), username)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
		case services.ErrInvalidPayload:
			app.BadRequestResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}
	response := &responses.GetUserProfileResponse{
		Username:      userProfile.User.Username,
		FirstName:     userProfile.User.FirstName,
		LastName:      userProfile.User.LastName,
		Description:   userProfile.Description,
		AvatarURL:     userProfile.AvatarURL,
		BannerURL:     userProfile.BannerURL,
		Location:      userProfile.Location,
		UserLink:      userProfile.UserLink,
		NumFollowing:  userProfile.NumFollowing,
		NumFollowers:  userProfile.NumFollowers,
		NumPosts:      userProfile.NumPosts,
		NumMediaPosts: userProfile.NumMediaPosts,
		CreatedAt:     userProfile.CreatedAt,
		UpdatedAt:     userProfile.UpdatedAt,
	}
	if err := httpio.JsonResponse(w, http.StatusCreated, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}

func (app *Application) getUserPosts(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		app.BadRequestResponse(w, r, httpio.ErrEmptyParam)
		return
	}
	query := r.URL.Query()
	limitParam := query.Get("limit")
	limit := 10
	if limitParam != "" {
		numParsed, err := httpio.ParseAsInt(limitParam)
		if err != nil {
			app.BadRequestResponse(w, r, err)
			return
		}
		limit = numParsed
	}
	cursorParam := query.Get("cursor")
	cursor := time.Now()
	if cursorParam != "" {
		parsedTime, err := time.Parse(time.DateTime, cursorParam)
		if err != nil {
			app.BadRequestResponse(w, r, err)
			return
		}
		cursor = parsedTime
	}
	posts, err := app.Service.User.GetPosts(r.Context(), username, cursor, limit)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	var response responses.GetUserPostsResponse
	response.Posts = make([]responses.PostResponse, len(posts))
	for idx, post := range posts {
		response.Posts[idx] = responses.PostResponse{
			ID:        post.ID,
			Tittle:    post.Tittle,
			Content:   post.Content,
			MediaURL:  post.Media.String,
			Tags:      post.Tags,
			CreatedAt: post.CreatedAt,
			UpdatedAt: post.UpdatedAt,
			User: responses.UserResponse{
				ID:        post.User.ID,
				Username:  post.User.Username,
				FirstName: post.User.FirstName,
				LastName:  post.User.LastName,
			},
		}
	}
	if err := httpio.JsonResponse(w, http.StatusOK, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}
