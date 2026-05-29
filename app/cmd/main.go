package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/peterzhang086/crypto-asset-api/internal/handlers"
	"github.com/peterzhang086/crypto-asset-api/internal/middleware"
)

func main() {
	// Structured JSON logs — required for log aggregation / SIEM ingestion
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stdout).With().
		Str("service", "crypto-asset-api").
		Logger()

	port := getEnv("PORT", "8080")

	r := chi.NewRouter()

	// Middleware order matters
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.SecurityHeaders)
	r.Use(chimw.Timeout(15 * time.Second))

	// Health endpoints — separate paths for k8s liveness vs readiness
	r.Get("/healthz", handlers.Healthz)
	r.Get("/readyz", handlers.Readyz)

	// Business API — intentionally simple; rich threat modeling targets
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/wallets/{address}/balance", handlers.GetBalance)
		r.Post("/transactions/sign", handlers.SignTransaction)
		r.Get("/markets/{pair}", handlers.GetMarket)
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("port", port).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
