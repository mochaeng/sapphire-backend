package app

import (
	"net/http"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/models"
	service "github.com/mochaeng/sapphire-backend/internal/services"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

// RegisterUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		models.RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	models.RegisterUserResponse	"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/register/user [post]
func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload models.RegisterUserPayload
	if err := httpio.ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	userInviation, err := app.Service.Auth.RegisterUser(r.Context(), &payload)
	if err != nil {
		switch err {
		case service.ErrInvalidPayload:
			app.BadRequestResponse(w, r, err)
		case store.ErrDuplicateEmail, store.ErrDuplicateUsername:
			app.ConflictResponse(w, r, err)
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}

	response := &models.RegisterUserResponse{
		Username:  userInviation.User.Username,
		CreatedAt: userInviation.User.CreatedAt,
		IsActive:  userInviation.User.IsActive,
		Token:     userInviation.Token,
	}
	if err := httpio.JsonResponse(w, http.StatusCreated, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
	}
}

// CreateUserToken godoc
//
//	@Summary		Creates a token for a activated user
//	@Description	This token is used for a user to access protected routes
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		models.CreateUserTokenPayload	true	"User credentials"
//	@Success		201		{object}	models.CreateTokenResponse	"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/token [post]
func (app *Application) createUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload models.CreateUserTokenPayload
	if err := httpio.ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	token, err := app.Service.Auth.CreateUserToken(r.Context(), &payload)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.UnauthorizedErrorResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}

	response := models.CreateTokenResponse{
		Token: token,
	}
	if err := httpio.JsonResponse(w, http.StatusCreated, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}
