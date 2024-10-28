package app

import (
	"net/http"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
)

func (app *Application) InternalServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("internal server error", "path", r.URL.Path, "method", r.Method, "error", err)
	httpio.WriteJSONWithError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *Application) ConflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Errorw("conflict error", "path", r.URL.Path, "method", r.Method, "error", err)
	httpio.WriteJSONWithError(w, http.StatusConflict, err.Error())
}

func (app *Application) BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnw("bad request error", "path", r.URL.Path, "method", r.Method, "error", err)
	httpio.WriteJSONWithError(w, http.StatusBadRequest, err.Error())
}

func (app *Application) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnw("not found error", "path", r.URL.Path, "method", r.Method, "error", err)
	httpio.WriteJSONWithError(w, http.StatusNotFound, "not found")
}

func (app *Application) UnauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnw("unauthorized error", "path", r.URL.Path, "method", r.Method, "error", err)
	httpio.WriteJSONWithError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) UnauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnw("unauthorized basic error", "path", r.URL.Path, "method", r.Method, "error", err)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	httpio.WriteJSONWithError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *Application) ForbiddenErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Warnw("forbidden error", "path", r.URL.Path, "method", r.Method, "error", err)
	httpio.WriteJSONWithError(w, http.StatusForbidden, "forbidden")
}

func (app *Application) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.Logger.Warn("rate limit exceed", "method", r.Method, "path", r.URL.Path)
	w.Header().Set("Retry-After", retryAfter)
	httpio.WriteJSON(w, http.StatusTooManyRequests, "rate limit exceed, retry after: "+retryAfter)
}
