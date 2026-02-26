// Package main is the entry point for the smartphone search engine.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/alessandrolattao/qdrant-experiment/internal/embedder"
	appqdrant "github.com/alessandrolattao/qdrant-experiment/internal/qdrant"
	"github.com/alessandrolattao/qdrant-experiment/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting qdrant smartphone search engine")

	qdrantHost := getEnv("QDRANT_HOST", "localhost")
	qdrantPort := getEnvInt("QDRANT_PORT", 6334)
	embedderURL := getEnv("EMBEDDER_URL", "http://localhost:8000")
	listenAddr := getEnv("LISTEN_ADDR", ":8080")
	imagesDir := getEnv("IMAGES_DIR", "images")

	client, err := appqdrant.NewClient(qdrantHost, qdrantPort)
	if err != nil {
		slog.Error("failed to connect to qdrant", slog.String("error", err.Error()))
		os.Exit(1)
	}

	embedClient := embedder.NewClient(embedderURL)
	seeder := appqdrant.NewSeeder(client, embedClient, "data/smartphones.csv", imagesDir)

	go func() {
		if err := seeder.SeedIfNeeded(); err != nil {
			slog.Error("seed failed", slog.String("error", err.Error()))
		}
	}()

	searcher := appqdrant.NewSearcher(client, embedClient)
	srv := server.New(searcher, imagesDir)

	slog.Info("server listening", slog.String("addr", listenAddr))

	if err := http.ListenAndServe(listenAddr, srv.Handler()); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
}
