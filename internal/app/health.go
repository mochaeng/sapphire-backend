package app

import (
	"net/http"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
)

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.Config.Env,
		"version": app.Config.Version,
	}
	if err := httpio.JsonResponse(w, 200, data); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}
