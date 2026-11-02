package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weatherwatch/internal/database"
	"weatherwatch/internal/forecast"
	"weatherwatch/internal/openmeteo"
	"weatherwatch/internal/watchlist"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := getenv("ADDR", ":8080")
	dbPath := getenv("DB_PATH", "weather.db")

	db, err := database.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	openMeteo := openmeteo.New()
	store := watchlist.NewStore(db)
	watchlistSvc := watchlist.NewService(store, openMeteo)
	forecastSvc := forecast.NewService(store, openMeteo, 10*time.Minute)

	srv := &http.Server{
		Addr:              addr,
		Handler:           newRouter(log, watchlist.NewHandler(watchlistSvc), forecast.NewHandler(forecastSvc)),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", addr, "db", dbPath)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
