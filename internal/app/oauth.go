package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

func (app *Application) OAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	r = setProviderInContext(r, provider)

	gothic.BeginAuthHandler(w, r)
}

func (app *Application) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	r = setProviderInContext(r, provider)

	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		app.BadRequestResponse(w, r, fmt.Errorf("oauth failed. Error: %w", err))
		return
	}

	// 1. check if email exists and it's verified by the provider
	// 2. check if there's an already user in the database with that email
	// 3. if the user exists:
	// 		  if the user already has oauth account:
	// 			- create session
	// 		  else:
	// 			- create oauth record (provider, providerUserID, userID)
	// 			- activated user if it's not yet
	// 			- create session
	// 	  else:
	// 		- create user record
	// 		- create oauth record
	// 		- create session

	ctx := context.Background()
	user, err := app.Service.User.LinkOrCreateUserFromOAuth(ctx, &gothUser)
	if err != nil {
		app.BadRequestResponse(w, r, fmt.Errorf("oauth creation failed. Error: %w", err))
		return
	}

	cookie, err := app.Service.Auth.GetCookieSession(user.ID)
	if err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}

	redirectURL := fmt.Sprintf("%s/auth/oauth-success", app.Config.FrontedURL)
	http.SetCookie(w, cookie)
	http.Redirect(w, r, redirectURL, http.StatusPermanentRedirect)

	// app.Logger.Infow("provider", provider)
	// app.Logger.Infow("gothUser", gothUser)
}

func setProviderInContext(r *http.Request, provider string) *http.Request {
	//lint:ignore SA1029 gothic expects a string
	return r.WithContext(context.WithValue(r.Context(), "provider", provider))
}
