package qdrant

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alessandrolattao/qdrant-experiment/internal/embedder"
	"github.com/alessandrolattao/qdrant-experiment/internal/model"
	qdrantclient "github.com/qdrant/go-client/qdrant"
)

// SearchFilters holds optional filters for narrowing search results.
type SearchFilters struct {
	Brand       string
	NFC         *bool   // nil = no filter, true = has NFC, false = no NFC
	NetGen      string  // "5G", "LTE", "3G", "2G" or ""
	OS          string  // "Android", "iOS", "Windows", "Other" or ""
	DisplayType string  // "AMOLED", "OLED", "IPS", "TFT", "LCD", "Other" or ""
	PriceMin    float64 // 0 = no lower bound
	PriceMax    float64 // 0 = no upper bound
}

// Searcher performs vector search in Qdrant using CLIP and MiniLM embeddings.
type Searcher struct {
	client   *qdrantclient.Client
	embedder *embedder.Client
}

// NewSearcher creates a new Searcher.
func NewSearcher(client *qdrantclient.Client, embedder *embedder.Client) *Searcher {
	return &Searcher{
		client:   client,
		embedder: embedder,
	}
}

// SearchByText embeds the query with MiniLM and searches the "text" named vector.
func (s *Searcher) SearchByText(ctx context.Context, query string, limit uint64, filters SearchFilters) ([]model.Smartphone, error) {
	embedding, err := s.embedder.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embedding text: %w", err)
	}

	using := "text"

	return s.searchByVector(ctx, embedding, &using, limit, filters)
}

// SearchByImage embeds the image with CLIP and searches the "image" named vector.
func (s *Searcher) SearchByImage(ctx context.Context, imageData io.Reader, filename string, limit uint64, filters SearchFilters) ([]model.Smartphone, error) {
	embedding, err := s.embedder.EmbedImage(ctx, imageData, filename)
	if err != nil {
		return nil, fmt.Errorf("embedding image: %w", err)
	}

	using := "image"

	return s.searchByVector(ctx, embedding, &using, limit, filters)
}

// AvailableBrands returns all unique brand values from the collection.
func (s *Searcher) AvailableBrands(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	brands := map[string]struct{}{}
	var offset *qdrantclient.PointId

	scrollLimit := uint32(1000)

	for {
		points, err := s.client.Scroll(ctx, &qdrantclient.ScrollPoints{
			CollectionName:  collectionName,
			Limit:           &scrollLimit,
			Offset:          offset,
			WithPayload:     qdrantclient.NewWithPayloadInclude("brand"),
			WithVectors:     qdrantclient.NewWithVectors(false),
		})
		if err != nil {
			return nil, fmt.Errorf("scrolling brands: %w", err)
		}

		for _, p := range points {
			if b := payloadString(p.Payload, "brand"); b != "" {
				brands[b] = struct{}{}
			}
		}

		if len(points) < 1000 {
			break
		}

		offset = points[len(points)-1].Id
	}

	result := make([]string, 0, len(brands))
	for b := range brands {
		result = append(result, b)
	}

	return result, nil
}

func (s *Searcher) searchByVector(ctx context.Context, vector []float32, using *string, limit uint64, filters SearchFilters) ([]model.Smartphone, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	qp := &qdrantclient.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrantclient.NewQuery(vector...),
		Using:          using,
		Limit:          &limit,
		WithPayload:    qdrantclient.NewWithPayload(true),
	}

	if f := buildFilter(filters); f != nil {
		qp.Filter = f
	}

	results, err := s.client.Query(ctx, qp)
	if err != nil {
		return nil, fmt.Errorf("querying qdrant: %w", err)
	}

	phones := make([]model.Smartphone, 0, len(results))

	for _, point := range results {
		phone := payloadToSmartphone(point.Payload)
		phone.Score = point.Score
		phones = append(phones, phone)
	}

	return phones, nil
}

func buildFilter(filters SearchFilters) *qdrantclient.Filter {
	var conditions []*qdrantclient.Condition

	if filters.Brand != "" {
		conditions = append(conditions, qdrantclient.NewMatch("brand", filters.Brand))
	}

	if filters.NFC != nil {
		if *filters.NFC {
			conditions = append(conditions, matchPrefix("nfc", "Yes"))
		} else {
			conditions = append(conditions, qdrantclient.NewMatch("nfc", "No"))
		}
	}

	if filters.NetGen != "" {
		conditions = append(conditions, matchContains("technology", filters.NetGen))
	}

	if filters.OS != "" {
		conditions = append(conditions, qdrantclient.NewMatch("os_family", filters.OS))
	}

	if filters.DisplayType != "" {
		conditions = append(conditions, qdrantclient.NewMatch("display_type", filters.DisplayType))
	}

	if filters.PriceMin > 0 || filters.PriceMax > 0 {
		r := &qdrantclient.Range{}
		if filters.PriceMin > 0 {
			r.Gte = &filters.PriceMin
		}
		if filters.PriceMax > 0 {
			r.Lte = &filters.PriceMax
		}
		conditions = append(conditions, qdrantclient.NewRange("price_eur", r))
	}

	if len(conditions) == 0 {
		return nil
	}

	return &qdrantclient.Filter{Must: conditions}
}

func matchPrefix(field, prefix string) *qdrantclient.Condition {
	return &qdrantclient.Condition{
		ConditionOneOf: &qdrantclient.Condition_Field{
			Field: &qdrantclient.FieldCondition{
				Key: field,
				Match: &qdrantclient.Match{
					MatchValue: &qdrantclient.Match_Text{
						Text: prefix,
					},
				},
			},
		},
	}
}

func matchContains(field, substring string) *qdrantclient.Condition {
	return &qdrantclient.Condition{
		ConditionOneOf: &qdrantclient.Condition_Field{
			Field: &qdrantclient.FieldCondition{
				Key: field,
				Match: &qdrantclient.Match{
					MatchValue: &qdrantclient.Match_Text{
						Text: strings.ToUpper(substring),
					},
				},
			},
		},
	}
}


func payloadToSmartphone(payload map[string]*qdrantclient.Value) model.Smartphone {
	return model.Smartphone{
		Brand:      payloadString(payload, "brand"),
		Model:      payloadString(payload, "model"),
		ImageURL:   payloadString(payload, "image_url"),
		ImageFile:  payloadString(payload, "image_file"),
		Technology: payloadString(payload, "technology"),
		Announced:  payloadString(payload, "announced"),
		Status:     payloadString(payload, "status"),
		Dimensions: payloadString(payload, "dimensions"),
		Weight:     payloadString(payload, "weight"),
		SIM:        payloadString(payload, "sim"),
		Display:    payloadString(payload, "display"),
		ScreenSize: payloadString(payload, "screen_size"),
		Resolution: payloadString(payload, "resolution"),
		Protection: payloadString(payload, "protection"),
		OS:         payloadString(payload, "os"),
		Chipset:    payloadString(payload, "chipset"),
		CPU:        payloadString(payload, "cpu"),
		GPU:        payloadString(payload, "gpu"),
		CardSlot:   payloadString(payload, "card_slot"),
		Storage:    payloadString(payload, "storage"),
		Camera:     payloadString(payload, "camera"),
		Video:      payloadString(payload, "video"),
		Selfie:     payloadString(payload, "selfie"),
		Battery:    payloadString(payload, "battery"),
		Charging:   payloadString(payload, "charging"),
		WLAN:       payloadString(payload, "wlan"),
		Bluetooth:  payloadString(payload, "bluetooth"),
		GPS:        payloadString(payload, "gps"),
		NFC:        payloadString(payload, "nfc"),
		USB:        payloadString(payload, "usb"),
		Sensors:    payloadString(payload, "sensors"),
		Colors:     payloadString(payload, "colors"),
		Price:      payloadString(payload, "price"),
	}
}

func payloadString(payload map[string]*qdrantclient.Value, key string) string {
	v, ok := payload[key]
	if !ok || v == nil {
		return ""
	}

	return v.GetStringValue()
}
