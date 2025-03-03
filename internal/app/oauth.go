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

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	gothic.BeginAuthHandler(w, r)
}

func (app *Application) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

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
	err = app.Service.User.LinkOrCreateUserFromOAuth(ctx, &gothUser)
	if err != nil {
		app.BadRequestResponse(w, r, fmt.Errorf("oauth creation failed. Error: %w", err))
		return
	}

	// return a user from the linkorcreateuserfromoauth
	app.Logger.Infow("frontend URL", app.Config.FrontedURL)

	// token, err := app.Service.Auth.GenerateSessionToken()
	// if err != nil {
	// 	app.InternalServerErrorResponse(w, r, err)
	// 	return
	// }

	// session, err := app.Service.Auth.CreateSession(token, user.ID)
	// if err != nil {
	// 	app.InternalServerErrorResponse(w, r, err)
	// 	return
	// }

	// cookie := http.Cookie{
	// 	Name:     AuthTokenKey,
	// 	Value:    token,
	// 	Path:     "/",
	// 	HttpOnly: true,
	// 	Secure:   true,
	// 	MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
	// 	SameSite: http.SameSiteLaxMode,
	// }
	// http.SetCookie(w, &cookie)

	http.Redirect(w, r, app.Config.FrontedURL, http.StatusPermanentRedirect)

	fmt.Println(provider)
	fmt.Println(gothUser)
}

func setProviderInContext(r *http.Request, provider string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), "provider", provider))
}
