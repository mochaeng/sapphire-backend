package app

import (
	"net/http"
	"time"

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
//	@Failure		409		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/signup [post]
func (app *Application) signupHandler(w http.ResponseWriter, r *http.Request) {
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

// SiginHandler godoc
//
//	@Summary		Signs user in the application
//	@Description
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		models.CreateUserTokenPayload	true	"User credentials"
//	@Success		204		"user has signin"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/signin [post]
func (app *Application) signinHandler(w http.ResponseWriter, r *http.Request) {
	var payload models.SigninPayload
	if err := httpio.ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	user, err := app.Service.Auth.Authenticate(r.Context(), &payload)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.UnauthorizedErrorResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.Service.Auth.GenerateSessionToken()
	if err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}

	session, err := app.Service.Auth.CreateSession(token, user.ID)
	if err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}

	cookie := http.Cookie{
		Name:     AuthTokenKey,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	httpio.NoContentResponse(w)
}

// SignoutHandler godoc
//
//	@Summary		Signouts a user from the application
//	@Description	Invalidate the user's session and delete the HTTP-only auth token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		204		"user session deleted"
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/signout [post]
func (app *Application) signoutHandler(w http.ResponseWriter, r *http.Request) {
	session := getSessionFromContext(r)
	if session == nil {
		app.InternalServerErrorResponse(w, r, ErrSessionContextNotFound)
		return
	}

	err := app.Service.Auth.InvalidateSession(session.ID)
	if err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}

	app.deleteUserSessionCookie(w)

	httpio.NoContentResponse(w)
}

// AuthStatusHandler godoc
//
//	@Summary		Check the auth status of a user
//	@Description	Check if the session token set with HTTPOnly by the backend is valid
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		204		"user is authenticated"
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/status [get]
func (app *Application) authStatusHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		app.InternalServerErrorResponse(w, r, ErrUserContextNotFound)
		return
	}

	httpio.NoContentResponse(w)
}
