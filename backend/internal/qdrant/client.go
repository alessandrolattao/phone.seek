// Package qdrant provides the Qdrant vector database client and seeding logic.
package qdrant

import (
	"fmt"

	qdrantclient "github.com/qdrant/go-client/qdrant"
)

// NewClient creates a new Qdrant gRPC client.
func NewClient(host string, port int) (*qdrantclient.Client, error) {
	client, err := qdrantclient.NewClient(&qdrantclient.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("creating qdrant client: %w", err)
	}

	return client, nil
}
