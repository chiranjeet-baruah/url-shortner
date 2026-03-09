// Package main is the entry point of the URL shortener application.
// It wires together all components (database, cache, tracker, HTTP server)
// and handles graceful shutdown on SIGINT/SIGTERM.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// The underscore import registers the PostgreSQL driver with database/sql.
	// Without it, sql.Open("postgres", ...) would fail.
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"url-shortener/internal/v1/shortener/adapter/driven"
	"url-shortener/internal/v1/shortener/adapter/driver"
	"url-shortener/internal/v1/shortener/service"
	"url-shortener/web"
)

func main() {
	// --------------- Configuration ---------------
	// Read settings from environment variables, falling back to defaults
	// suitable for local development.
	databaseURL := envOrDefault("DATABASE_URL", "postgres://shortener:shortener@localhost:54321/shortener?sslmode=disable")
	redisAddr := envOrDefault("REDIS_ADDR", "localhost:6379")
	baseURL := envOrDefault("BASE_URL", "http://localhost:8080")
	addr := envOrDefault("ADDR", ":8080")

	// --------------- PostgreSQL ---------------
	// sql.Open doesn't actually connect — it just validates the DSN.
	// The real connection happens on the first query (or Ping below).
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer db.Close() // Ensure the connection pool is closed when main() exits.

	ctx := context.Background()
	// Ping verifies we can actually reach the database.
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("pinging database: %v", err)
	}

	// Create the "urls" table if it doesn't exist yet.
	repo := driven.NewPostgresRepo(db)
	if err := repo.AutoMigrate(ctx); err != nil {
		log.Fatalf("auto-migrate: %v", err)
	}

	// --------------- Redis ---------------
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("pinging redis: %v", err)
	}

	cache := driven.NewRedisCache(rdb)

	// --------------- Click Tracker ---------------
	// Starts a background goroutine that writes click counts to the DB.
	tracker := driven.NewTracker(repo)
	tracker.Start(ctx)

	// --------------- Service + HTTP Router ---------------
	// Wire everything together: repository, cache, tracker → service → router.
	svc := service.New(repo, cache, tracker, baseURL)
	router := driver.NewRouter(svc, web.StaticFS)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,   // Max time to read the entire request
		WriteTimeout: 10 * time.Second,  // Max time to write the response
		IdleTimeout:  120 * time.Second, // Max time to keep idle keep-alive connections
	}

	// --------------- Graceful Shutdown ---------------
	// Listen for OS signals (Ctrl+C or `kill`) to shut down cleanly.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a separate goroutine so we can listen for signals.
	go func() {
		log.Printf("server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-done // Block here until a shutdown signal is received.
	log.Println("shutting down...")

	// Give in-flight requests up to 10 seconds to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	// Flush remaining click events before exiting.
	tracker.Stop()
	log.Println("server stopped")
}

// envOrDefault reads an environment variable by key. If it's empty or not set,
// it returns the provided fallback value instead.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
