// Package embedder provides an HTTP client for the embedding service.
package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"
)

// Client communicates with the embedding service (CLIP + BGE-M3).
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new embedder client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type textRequest struct {
	Text string `json:"text"`
}

type textsRequest struct {
	Texts []string `json:"texts"`
}

type imagePathsRequest struct {
	Paths []string `json:"paths"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type embeddingsResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// EmbedText returns the BGE-M3 embedding for a text query (1024d).
func (c *Client) EmbedText(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(textRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	return c.postEmbedding(ctx, "/embed/text", body)
}

// EmbedTexts returns BGE-M3 embeddings for a batch of texts (1024d each).
func (c *Client) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(textsRequest{Texts: texts})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	return c.postEmbeddings(ctx, "/embed/texts", body)
}

// EmbedImage returns the CLIP embedding for an uploaded image (512d).
func (c *Client) EmbedImage(ctx context.Context, imageData io.Reader, filename string) ([]float32, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	if _, err := io.Copy(part, imageData); err != nil {
		return nil, fmt.Errorf("copying image data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embed/image", &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedder returned status %d", resp.StatusCode)
	}

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Embedding, nil
}

// EmbedImagePaths returns CLIP embeddings for images at the given file paths (512d each).
func (c *Client) EmbedImagePaths(ctx context.Context, paths []string) ([][]float32, error) {
	body, err := json.Marshal(imagePathsRequest{Paths: paths})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	return c.postEmbeddings(ctx, "/embed/image-paths", body)
}

// WaitReady polls the embedder health endpoint until it responds.
func (c *Client) WaitReady(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
		if err != nil {
			return fmt.Errorf("creating health request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				slog.Info("embedder service is ready")
				return nil
			}
		}

		slog.Info("waiting for embedder service...")
		time.Sleep(2 * time.Second)
	}
}

func (c *Client) postEmbedding(ctx context.Context, path string, body []byte) ([]float32, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedder returned status %d", resp.StatusCode)
	}

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Embedding, nil
}

func (c *Client) postEmbeddings(ctx context.Context, path string, body []byte) ([][]float32, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedder returned status %d", resp.StatusCode)
	}

	var result embeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Embeddings, nil
}
