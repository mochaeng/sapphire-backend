package app

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mochaeng/sapphire-backend/docs"
	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/ratelimiter"
	service "github.com/mochaeng/sapphire-backend/internal/services"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type Application struct {
	Service     *service.Service
	Config      *config.Cfg
	Logger      *zap.SugaredLogger
	RateLimiter ratelimiter.RateLimiter
}

func (app *Application) Mount() http.Handler {
	docsURL := fmt.Sprintf("%s/swagger/doc.json", app.Config.Addr)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:7777", "https://localhost:7777",
			"http://localhost:3000", "https://localhost:3000",
			"http://localhost:5173", "https://localhost:5173",
			"http://localhost:4173", "https://localhost:4173",
		},
		// AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(app.csrfMiddleware)
	if app.Config.RateLimiter.IsEnable {
		r.Use(app.rateLimiterMiddleware)
	}
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.With(app.basicAuthMiddleware).Get("/debug/vars", expvar.Handler().ServeHTTP)

		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/user", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				r.With(app.userContextMiddleware).Get("/", app.getUserHandler)
				r.Group(func(r chi.Router) {
					r.Use(app.authTokenMiddleware)
					r.Put("/follow", app.followUserHandler)
					r.Put("/unfollow", app.unfollowUserHandler)
				})
			})
			r.With(app.authTokenMiddleware).Get("/feed", app.GetUserFeedHandler)
			r.Route("/by", func(r chi.Router) {
				r.Get("/{username}", app.getUserByUsername)
			})
		})

		r.Route("/post", func(r chi.Router) {
			r.With(app.authTokenMiddleware).Post("/", app.createPostHandler)
			r.Route("/{postID}", func(r chi.Router) {
				r.With(app.postContextMiddleware).Get("/", app.getPostHandler)
				r.Group(func(r chi.Router) {
					r.Use(app.authTokenMiddleware)
					r.Use(app.postContextMiddleware)
					r.Patch("/", app.checkPostOwnership(config.Roles["moderator"].Level, app.updatePostHandler))
					r.Delete("/", app.checkPostOwnership(config.Roles["admin"].Level, app.deletePostHandler))
				})
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", app.signupHandler)
			r.Post("/signin", app.signinHandler)
			r.With(app.authTokenMiddleware).Post("/signout", app.signoutHandler)
			r.With(app.authTokenMiddleware).Get("/status", app.authStatusHandler)
		})

		r.Route("/verify-email", func(r chi.Router) {
			r.Put("/{token}", app.activateUserHandler)
		})
	})

	fs := http.FileServer(http.Dir(app.Config.MediaFolder))
	r.Handle(fmt.Sprintf("/%s/*", app.Config.MediaFolder), http.StripPrefix(fmt.Sprintf("/%s/", app.Config.MediaFolder), fs))
	return r
}

func (app *Application) Run(mux http.Handler) error {
	docs.SwaggerInfo.Version = app.Config.Version
	docs.SwaggerInfo.Host = app.Config.ApiURL
	docs.SwaggerInfo.BasePath = "/v1"

	server := http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		app.Logger.Infow("signal caught", "signal", s.String())
		shutdown <- server.Shutdown(ctx)
	}()

	app.Logger.Infow("server has started", "addr", app.Config.Addr, "env", app.Config.Env)
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdown
	if err != nil {
		return err
	}
	app.Logger.Infow("server has stopped", "addr", app.Config.Addr, "env", app.Config.Env)
	return nil
}
