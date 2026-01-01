//go:generate go run github.com/vektra/mockery/v2@latest --name=URLSaver

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/url/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/logger/handlers/slogpretty"

	//"url-shortener/internal/storage/sqlite"
	mwLogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/storage/postgres"

	//"url-shortener/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	configuration := config.MustLoad()

	log := setupLogger(configuration.Env)

	log.Info("starting url-shortener", slog.String("env", configuration.Env))
	log.Debug("debug messages are enabled")
	// TODO: DELETE AFTER DEBUG!!!!!!!!!!
	fmt.Println(configuration)

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		configuration.Database.Host,
		configuration.Database.Port,
		configuration.Database.User,
		configuration.Database.Password,
		configuration.Database.DBName,
		configuration.Database.SSLMode,
	)

	// ngl I couldn't figure out all the drivers shit with sqlite so I went with postgres
	// idk I use sqlite at work and I am so fed up with it so I'm biased as well

	storage, err := postgres.New(connString)
	if err != nil {
		log.Error("failed to init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	//id, err := storage.SaveURL("https://google.com", "google")
	//if err != nil {
	//	log.Error("failed to save url", slog.String("error", err.Error()))
	//	os.Exit(1)
	//}
	//
	//log.Info("saved url", slog.Int64("id", id))

	router := chi.NewRouter()
	// middleware - other handlers for like auth
	// this one adds request id to every request
	router.Use(middleware.RequestID)
	// why would you need user's IP but alright brodie
	router.Use(middleware.RealIP)
	// logging
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	// in case of panics
	router.Use(middleware.Recoverer)
	// /address/{id}
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			configuration.HTTPServer.User: configuration.HTTPServer.Password,
		}))
		r.Post("/", save.New(log, storage))
		// TODO: add DELETE /url/{id}
	})
	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", configuration.Address))

	server := &http.Server{
		Addr:         configuration.Address,
		Handler:      router,
		ReadTimeout:  configuration.HTTPServer.Timeout,
		WriteTimeout: configuration.HTTPServer.Timeout,
		IdleTimeout:  configuration.HTTPServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
