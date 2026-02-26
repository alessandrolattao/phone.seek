package qdrant

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alessandrolattao/qdrant-experiment/internal/csvparser"
	"github.com/alessandrolattao/qdrant-experiment/internal/embedder"
	"github.com/alessandrolattao/qdrant-experiment/internal/model"
	qdrantclient "github.com/qdrant/go-client/qdrant"
)

const (
	collectionName = "smartphones"
	batchSize      = 64
	imageVectorSize = 512 // CLIP ViT-B/32
	textVectorSize  = 1024 // BAAI/bge-m3
	downloadConcurrency = 10
)

// Seeder handles loading smartphone data into Qdrant.
type Seeder struct {
	client    *qdrantclient.Client
	embedder  *embedder.Client
	csvPath   string
	imagesDir string
}

// NewSeeder creates a new Seeder.
func NewSeeder(client *qdrantclient.Client, embedder *embedder.Client, csvPath, imagesDir string) *Seeder {
	return &Seeder{
		client:    client,
		embedder:  embedder,
		csvPath:   csvPath,
		imagesDir: imagesDir,
	}
}

// SeedIfNeeded checks if data is already loaded, and imports from CSV if not.
func (s *Seeder) SeedIfNeeded() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := s.client.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("checking collection: %w", err)
	}

	if exists {
		info, err := s.client.GetCollectionInfo(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("getting collection info: %w", err)
		}

		var points uint64
		if info.PointsCount != nil {
			points = *info.PointsCount
		}

		slog.Info("collection already seeded, skipping",
			slog.String("collection", collectionName),
			slog.Uint64("points", points),
		)

		return nil
	}

	slog.Info("collection not found, starting seed", slog.String("collection", collectionName))

	if err := s.createCollection(); err != nil {
		return err
	}

	phones, err := csvparser.ParseFile(s.csvPath)
	if err != nil {
		return fmt.Errorf("parsing csv: %w", err)
	}

	slog.Info("parsed smartphones from csv", slog.Int("count", len(phones)))

	if err := os.MkdirAll(s.imagesDir, 0o755); err != nil {
		return fmt.Errorf("creating images dir: %w", err)
	}

	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer waitCancel()

	if err := s.embedder.WaitReady(waitCtx); err != nil {
		return fmt.Errorf("waiting for embedder: %w", err)
	}

	total := len(phones)

	for i := 0; i < total; i += batchSize {
		end := min(i+batchSize, total)
		batch := phones[i:end]

		slog.Info("processing",
			slog.String("embeddings", fmt.Sprintf("%d/%d", end, total)),
		)

		if err := s.processBatch(batch, uint64(i)); err != nil {
			return fmt.Errorf("processing batch %d-%d: %w", i, end, err)
		}

		slog.Info("processing",
			slog.String("imported", fmt.Sprintf("%d/%d", end, total)),
		)
	}

	slog.Info("seed complete", slog.Int("total", total))

	return nil
}

func (s *Seeder) createCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.client.CreateCollection(ctx, &qdrantclient.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrantclient.NewVectorsConfigMap(map[string]*qdrantclient.VectorParams{
			"image": {Size: imageVectorSize, Distance: qdrantclient.Distance_Cosine},
			"text":  {Size: textVectorSize, Distance: qdrantclient.Distance_Cosine},
		}),
	}); err != nil {
		return fmt.Errorf("creating collection: %w", err)
	}

	// Create payload indexes for filtering
	keywordType := qdrantclient.FieldType_FieldTypeKeyword
	textType := qdrantclient.FieldType_FieldTypeText
	floatType := qdrantclient.FieldType_FieldTypeFloat
	wait := true

	indexes := []struct {
		field     string
		fieldType *qdrantclient.FieldType
	}{
		{"brand", &keywordType},
		{"nfc", &textType},
		{"technology", &textType},
		{"os_family", &keywordType},
		{"display_type", &keywordType},
		{"price_eur", &floatType},
	}

	for _, idx := range indexes {
		idxCtx, idxCancel := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := s.client.CreateFieldIndex(idxCtx, &qdrantclient.CreateFieldIndexCollection{
			CollectionName: collectionName,
			FieldName:      idx.field,
			FieldType:      idx.fieldType,
			Wait:           &wait,
		})

		idxCancel()

		if err != nil {
			return fmt.Errorf("creating index on %s: %w", idx.field, err)
		}

		slog.Info("created payload index", slog.String("field", idx.field))
	}

	return nil
}

func (s *Seeder) processBatch(batch []model.Smartphone, offset uint64) error {
	// Phase 1: download images concurrently
	var wg sync.WaitGroup
	sem := make(chan struct{}, downloadConcurrency)

	for i := range batch {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			batch[idx].ImageFile = s.downloadImage(&batch[idx])
		}(i)
	}

	wg.Wait()

	// Phase 2: text embeddings (batch)
	descriptions := make([]string, len(batch))
	for i, phone := range batch {
		descriptions[i] = phone.Description()
	}

	embedCtx, embedCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	textEmbeddings, err := s.embedder.EmbedTexts(embedCtx, descriptions)
	embedCancel()

	if err != nil {
		return fmt.Errorf("text embeddings: %w", err)
	}

	// Phase 3: image embeddings (batch via file paths)
	var imagePaths []string
	imageIndexes := map[int]int{}

	for i, phone := range batch {
		if phone.ImageFile == "" {
			continue
		}

		imgPath := filepath.Join(s.imagesDir, phone.ImageFile)
		imageIndexes[i] = len(imagePaths)
		imagePaths = append(imagePaths, imgPath)
	}

	var imageEmbeddings [][]float32

	if len(imagePaths) > 0 {
		imgCtx, imgCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		imageEmbeddings, err = s.embedder.EmbedImagePaths(imgCtx, imagePaths)
		imgCancel()

		if err != nil {
			slog.Warn("image embeddings failed, continuing with text only", slog.String("error", err.Error()))
			imageEmbeddings = nil
		}
	}

	// Phase 4: build points and upsert
	points := make([]*qdrantclient.PointStruct, 0, len(batch))

	for i, phone := range batch {
		id := offset + uint64(i) + 1

		vectors := map[string]*qdrantclient.Vector{
			"text": {Data: textEmbeddings[i]},
		}

		if imgIdx, ok := imageIndexes[i]; ok && imageEmbeddings != nil && imgIdx < len(imageEmbeddings) {
			vectors["image"] = &qdrantclient.Vector{Data: imageEmbeddings[imgIdx]}
		}

		points = append(points, &qdrantclient.PointStruct{
			Id:      qdrantclient.NewIDNum(id),
			Vectors: qdrantclient.NewVectorsMap(vectors),
			Payload: qdrantclient.NewValueMap(phone.PayloadMap()),
		})
	}

	upsertCtx, upsertCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer upsertCancel()

	_, err = s.client.Upsert(upsertCtx, &qdrantclient.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Seeder) downloadImage(phone *model.Smartphone) string {
	if phone.ImageURL == "" {
		return ""
	}

	filename := phone.ImageFilename()
	if filename == "" {
		return ""
	}

	dest := filepath.Join(s.imagesDir, filename)

	// Skip if already downloaded
	if _, err := os.Stat(dest); err == nil {
		return filename
	}

	resp, err := http.Get(phone.ImageURL) //nolint:noctx // fire-and-forget download during seed
	if err != nil {
		slog.Warn("failed to download image", slog.String("url", phone.ImageURL), slog.String("error", err.Error()))
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("image download bad status", slog.String("file", filename), slog.Int("status", resp.StatusCode))
		return ""
	}

	f, err := os.Create(dest)
	if err != nil {
		slog.Warn("failed to create image file", slog.String("path", dest), slog.String("error", err.Error()))
		return ""
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, resp.Body); err != nil {
		slog.Warn("failed to write image", slog.String("path", dest), slog.String("error", err.Error()))
		_ = os.Remove(dest)

		return ""
	}

	return filename
}
