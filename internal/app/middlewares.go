package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

type postKey string
type userKey string

const (
	postCtx postKey = "post"
	userCtx userKey = "user"
)

func (app *Application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")
		postID, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.InternalServerErrorResponse(w, r, err)
			return
		}
		ctx := r.Context()
		post, err := app.Service.Post.GetWithUser(ctx, postID)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.NotFoundResponse(w, r, err)
			default:
				app.InternalServerErrorResponse(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil || userID < 1 {
			app.BadRequestResponse(w, r, err)
			return
		}
		user, err := app.Service.User.GetCached(r.Context(), userID)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.NotFoundResponse(w, r, err)
			default:
				app.InternalServerErrorResponse(w, r, err)
			}
			return
		}
		ctx := context.WithValue(r.Context(), userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.UnauthorizedBasicErrorResponse(w, r, ErrAuthorizationHeaderMissing)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Basic" {
			app.UnauthorizedBasicErrorResponse(w, r, ErrAuthorizationHeaderMalformed)
			return
		}
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			app.UnauthorizedBasicErrorResponse(w, r, err)
			return
		}
		username := app.Config.Auth.Basic.Username
		password := app.Config.Auth.Basic.Password
		creds := strings.SplitN(string(decoded), ":", 2)
		if len(creds) != 2 || creds[0] != username || creds[1] != password {
			app.UnauthorizedBasicErrorResponse(w, r, ErrInvalidCredentials)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *Application) authTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.UnauthorizedErrorResponse(w, r, ErrAuthorizationHeaderMissing)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.UnauthorizedErrorResponse(w, r, ErrAuthorizationHeaderMalformed)
			return
		}
		jwtToken, err := app.Service.Auth.ValidateToken(parts[1])
		if err != nil {
			app.UnauthorizedErrorResponse(w, r, err)
			return
		}
		claims := jwtToken.Claims.(jwt.MapClaims)
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.UnauthorizedErrorResponse(w, r, err)
			return
		}
		ctx := r.Context()
		user, err := app.Service.User.GetCached(ctx, userID)
		if err != nil {
			app.UnauthorizedErrorResponse(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) checkPostOwnership(requiredLevel int, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromCtx(r)
		if post.User.ID == user.ID {
			next.ServeHTTP(w, r)
			return
		}
		if user.Role.Level >= requiredLevel {
			next.ServeHTTP(w, r)
			return
		}
		app.ForbiddenErrorResponse(w, r, fmt.Errorf("no required level to own a post"))
	})
}

func (app *Application) rateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Config.RateLimiter.IsEnable {
			if allow, retryAfter := app.RateLimiter.Allow(r.RemoteAddr); !allow {
				app.RateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
