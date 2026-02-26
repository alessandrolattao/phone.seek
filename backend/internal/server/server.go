// Package server provides the HTTP API for the smartphone search engine.
package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	appqdrant "github.com/alessandrolattao/qdrant-experiment/internal/qdrant"
)

const defaultLimit = 20

// Server handles HTTP requests for smartphone search.
type Server struct {
	searcher  *appqdrant.Searcher
	imagesDir string
	mux       *http.ServeMux
}

// New creates a new HTTP server.
func New(searcher *appqdrant.Searcher, imagesDir string) *Server {
	s := &Server{
		searcher:  searcher,
		imagesDir: imagesDir,
		mux:       http.NewServeMux(),
	}

	s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	s.mux.HandleFunc("GET /api/filters", s.handleFilters)
	s.mux.HandleFunc("GET /api/search", s.handleSearchText)
	s.mux.HandleFunc("POST /api/search/image", s.handleSearchImage)
	s.mux.Handle("GET /api/images/", http.StripPrefix("/api/images/", http.FileServer(http.Dir(imagesDir))))

	return s
}

// Handler returns the HTTP handler with CORS middleware.
func (s *Server) Handler() http.Handler {
	return corsMiddleware(s.mux)
}

func (s *Server) handleFilters(w http.ResponseWriter, r *http.Request) {
	brands, err := s.searcher.AvailableBrands(r.Context())
	if err != nil {
		slog.Error("failed to load brands", slog.String("error", err.Error()))
		brands = nil
	} else {
		sort.Strings(brands)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"brands":       brands,
		"nfc":          []string{"Yes", "No"},
		"network":      []string{"5G", "LTE", "HSPA", "GSM"},
		"os":           []string{"Android", "iOS", "Windows", "Other"},
		"display_type": []string{"AMOLED", "OLED", "IPS", "TFT", "LCD", "Other"},
	})
}

func (s *Server) handleSearchText(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing query parameter 'q'"})
		return
	}

	filters := parseFilters(r)

	start := time.Now()

	phones, err := s.searcher.SearchByText(r.Context(), query, defaultLimit, filters)
	if err != nil {
		slog.Error("text search failed", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"results":  phones,
		"total":    len(phones),
		"time_ms":  time.Since(start).Milliseconds(),
	})
}

func (s *Server) handleSearchImage(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 << 20 // 10MB

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, header, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing image file"})
		return
	}
	defer func() { _ = file.Close() }()

	filters := parseFilters(r)

	start := time.Now()

	phones, err := s.searcher.SearchByImage(r.Context(), file, header.Filename, defaultLimit, filters)
	if err != nil {
		slog.Error("image search failed", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"results":  phones,
		"total":    len(phones),
		"time_ms":  time.Since(start).Milliseconds(),
	})
}

func parseFilters(r *http.Request) appqdrant.SearchFilters {
	var filters appqdrant.SearchFilters

	filters.Brand = r.FormValue("brand")
	filters.NetGen = r.FormValue("network")
	filters.OS = r.FormValue("os")
	filters.DisplayType = r.FormValue("display_type")

	if v := r.FormValue("price_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filters.PriceMin = f
		}
	}

	if v := r.FormValue("price_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			filters.PriceMax = f
		}
	}

	switch r.FormValue("nfc") {
	case "Yes":
		t := true
		filters.NFC = &t
	case "No":
		f := false
		filters.NFC = &f
	}

	return filters
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
