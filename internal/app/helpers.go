package app

import (
	"net/http"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/services"
)

func getUserFromContext(r *http.Request) *models.User {
	user, ok := r.Context().Value(userCtx).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func getSessionFromContext(r *http.Request) *models.Session {
	session, ok := r.Context().Value(sessionCtx).(*models.Session)
	if !ok {
		return nil
	}
	return session
}

func deleteCookie(w http.ResponseWriter, cookieName string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
}

func (app *Application) deleteUserSessionCookie(w http.ResponseWriter) {
	deleteCookie(w, services.AuthTokenKey)
}
