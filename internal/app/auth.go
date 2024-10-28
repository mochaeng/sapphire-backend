package app

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/mailer"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type RegisterUserPayload struct {
	Username  string `json:"username" validate:"required,max=16,min=3"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=3,max=72"`
	FirstName string `json:"first_name" validate:"required,min=2,max=30"`
	LastName  string `json:"last_name" validate:"max=30"`
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*models.User
	Token string `json:"token"`
}

// RegisterUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/register/user [post]
func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := httpio.ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	user := &models.User{
		Username:  payload.Username,
		Email:     payload.Email,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Role: models.Role{
			ID: roles["user"].id,
		},
	}
	if err := user.Password.Set(payload.Password); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}

	plainToken := uuid.NewString()
	sha256Token := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(sha256Token[:])
	invitation := &models.UserInvitation{
		User:    user,
		Token:   hashToken,
		Expired: app.Config.Mail.Expired,
	}

	ctx := r.Context()
	err := app.Store.User.CreateAndInvite(ctx, invitation)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail, store.ErrDuplicateUsername:
			app.BadRequestResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}

	isSandBox := app.Config.Env == "dev"
	activationURL := fmt.Sprintf("%s/confirm/%s", app.Config.FrontedURL, plainToken)
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}
	status, err := app.Mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, isSandBox)
	if err != nil {
		app.Logger.Errorw("error sending welcome email", "error", err)
		if err := app.Store.User.Delete(ctx, user.ID); err != nil {
			app.Logger.Errorw("error deleting user", "error", err)
		}
		app.InternalServerErrorResponse(w, r, err)
		return
	}
	app.Logger.Infow("Email sent", "status code", status)

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	if err := httpio.JsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.InternalServerErrorResponse(w, r, err)
	}
}

// CegisterUserTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		201		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/token [post]
func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserTokenPayload
	if err := httpio.ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	user, err := app.Store.User.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.UnauthorizedErrorResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.Authenticator.GenerateToken(user.ID)
	if err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}

	if err := httpio.JsonResponse(w, http.StatusCreated, token); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}
