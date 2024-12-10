package app

import (
	"net/http"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/models/payloads"
	"github.com/mochaeng/sapphire-backend/internal/models/responses"
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
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	RegisterUserResponse	"User registered"
//	@Failure		400		{object}	error
//	@Failure		409		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/signup [post]
func (app *Application) signupHandler(w http.ResponseWriter, r *http.Request) {
	var payload payloads.RegisterUserPayload
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

	response := &responses.RegisterUserResponse{
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
//	@Param			payload	body		responses.CreateUserTokenPayload	true	"User credentials"
//	@Success		204		"user has signin"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/signin [post]
func (app *Application) signinHandler(w http.ResponseWriter, r *http.Request) {
	var payload payloads.SigninPayload
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
//	@Router			/auth/status [post]
func (app *Application) authStatusHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		app.InternalServerErrorResponse(w, r, ErrUserContextNotFound)
		return
	}

	httpio.NoContentResponse(w)
}

// AuthMeHandler godoc
//
//	@Summary		Get authenticated user information
//	@Description	Retrieve details about the authenticated user based on the session token set by the backend via HTTPOnly cookies.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	UserResponse	"Authenticated user information"
//	@Failure		401		{object}	error			"User is not authenticated or token is invalid"
//	@Failure		500		{object}	error			"Internal server error"
//	@Router			/auth/me [post]
func (app *Application) authMeHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		app.InternalServerErrorResponse(w, r, ErrUserContextNotFound)
		return
	}

	response := &responses.AuthMeResponse{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		RoleName:  user.Role.Name,
	}
	if err := httpio.JsonResponse(w, http.StatusOK, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}
