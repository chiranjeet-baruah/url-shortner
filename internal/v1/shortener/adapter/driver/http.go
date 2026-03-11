// Package driver contains "driver" adapters — the entry points that *drive*
// (call into) the application core. In this project, the only driver is HTTP.
// This file sets up the HTTP routes and handler functions.
package driver

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"url-shortener/internal/v1/shortener/dto"
	"url-shortener/internal/v1/shortener/service"
)

// NewRouter creates an http.ServeMux with all the application routes:
//
//	POST /api/v1/shorten            → Create a short URL
//	GET  /r/{shortCode}              → Redirect to the original URL
//	GET  /api/v1/stats/{shortCode}   → Get click statistics
//	GET  /                           → Serve the web UI (embedded HTML files)
//
// Go 1.22+ supports method+pattern routing in the standard library,
// so no third-party router is needed.
func NewRouter(svc *service.Service, webFS fs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth())
	mux.HandleFunc("POST /api/v1/shorten", handleShorten(svc))
	mux.HandleFunc("GET /r/{shortCode}", handleRedirect(svc))
	mux.HandleFunc("GET /api/v1/stats/{shortCode}", handleStats(svc))
	mux.Handle("GET /", http.FileServerFS(webFS))

	return mux
}

// handleShorten returns an HTTP handler that reads a JSON body containing
// a URL, calls the service to shorten it, and responds with the short URL.
func handleShorten(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode the JSON request body into a ShortenRequest DTO.
		var req dto.ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		// Delegate to the service layer for validation and persistence.
		resp, err := svc.Shorten(r.Context(), req)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("encoding shorten response: %v", err)
		}
	}
}

// handleRedirect returns an HTTP handler that looks up a short code and
// sends an HTTP 301 (Moved Permanently) redirect to the original URL.
func handleRedirect(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// r.PathValue extracts the {shortCode} from the URL pattern (Go 1.22+).
		code := r.PathValue("shortCode")

		originalURL, err := svc.Resolve(r.Context(), code)
		if err != nil {
			http.Error(w, `{"error":"short URL not found"}`, http.StatusNotFound)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
	}
}

// handleHealth returns a simple liveness check handler that always responds
// with HTTP 200 and {"status":"ok"}. It has no dependencies on the service,
// database, or cache, so it can be used as a lightweight ping target.
func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			log.Printf("writing health response: %v", err)
		}
	}
}

// handleStats returns an HTTP handler that fetches and returns click
// statistics (original URL, click count) for a given short code.
func handleStats(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("shortCode")

		resp, err := svc.Stats(r.Context(), code)
		if err != nil {
			http.Error(w, `{"error":"short URL not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("encoding stats response: %v", err)
		}
	}
}
