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
		app.Logger.Errorw("OAuth failed", "err", err)
		return
	}

	fmt.Println(gothUser)
}
