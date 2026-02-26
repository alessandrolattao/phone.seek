// Package csvparser handles parsing of GSMArena smartphone CSV datasets.
package csvparser

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alessandrolattao/qdrant-experiment/internal/model"
)

// columnMapping maps CSV header names to Smartphone struct fields.
var columnMapping = []struct {
	csvColumn string
	setter    func(*model.Smartphone, string)
}{
	{"Brand", func(s *model.Smartphone, v string) { s.Brand = v }},
	{"Model Name", func(s *model.Smartphone, v string) { s.Model = v }},
	{"Model Image", func(s *model.Smartphone, v string) { s.ImageURL = v }},
	{"Technology", func(s *model.Smartphone, v string) { s.Technology = v }},
	{"Announced", func(s *model.Smartphone, v string) { s.Announced = v }},
	{"Status", func(s *model.Smartphone, v string) { s.Status = v }},
	{"Dimensions", func(s *model.Smartphone, v string) { s.Dimensions = v }},
	{"Weight", func(s *model.Smartphone, v string) { s.Weight = v }},
	{"SIM", func(s *model.Smartphone, v string) { s.SIM = v }},
	{"Type", func(s *model.Smartphone, v string) { s.Display = v }},
	{"Size", func(s *model.Smartphone, v string) { s.ScreenSize = v }},
	{"Resolution", func(s *model.Smartphone, v string) { s.Resolution = v }},
	{"Protection", func(s *model.Smartphone, v string) { s.Protection = v }},
	{"OS", func(s *model.Smartphone, v string) { s.OS = v }},
	{"Chipset", func(s *model.Smartphone, v string) { s.Chipset = v }},
	{"CPU", func(s *model.Smartphone, v string) { s.CPU = v }},
	{"GPU", func(s *model.Smartphone, v string) { s.GPU = v }},
	{"Card slot", func(s *model.Smartphone, v string) { s.CardSlot = v }},
	{"Internal", func(s *model.Smartphone, v string) { s.Storage = v }},
	{"Quad", func(s *model.Smartphone, v string) { s.Camera = v }},
	{"Video", func(s *model.Smartphone, v string) { s.Video = v }},
	{"Single", func(s *model.Smartphone, v string) { s.Selfie = v }},
	{"Type_1", func(s *model.Smartphone, v string) { s.Battery = v }},
	{"Charging", func(s *model.Smartphone, v string) { s.Charging = v }},
	{"WLAN", func(s *model.Smartphone, v string) { s.WLAN = v }},
	{"Bluetooth", func(s *model.Smartphone, v string) { s.Bluetooth = v }},
	{"GPS", func(s *model.Smartphone, v string) { s.GPS = v }},
	{"NFC", func(s *model.Smartphone, v string) { s.NFC = v }},
	{"USB", func(s *model.Smartphone, v string) { s.USB = v }},
	{"Sensors", func(s *model.Smartphone, v string) { s.Sensors = v }},
	{"Colors", func(s *model.Smartphone, v string) { s.Colors = v }},
	{"Price", func(s *model.Smartphone, v string) { s.Price = v }},
}

// ParseFile reads a GSMArena CSV and returns parsed smartphones.
func ParseFile(path string) ([]model.Smartphone, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening csv: %w", err)
	}
	defer func() { _ = f.Close() }()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading csv header: %w", err)
	}

	// Strip UTF-8 BOM from the first column header if present.
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\xEF\xBB\xBF")
	}

	colIndex := buildColumnIndex(header)

	var phones []model.Smartphone

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			continue
		}

		phone := parseRecord(record, colIndex)
		if phone.Brand == "" || phone.Model == "" {
			continue
		}

		phones = append(phones, phone)
	}

	return phones, nil
}

func parseRecord(record []string, colIndex map[string]int) model.Smartphone {
	var phone model.Smartphone

	for _, m := range columnMapping {
		idx, ok := colIndex[m.csvColumn]
		if !ok || idx >= len(record) {
			continue
		}

		m.setter(&phone, cleanValue(record[idx]))
	}

	return phone
}

func buildColumnIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, col := range header {
		idx[col] = i
	}

	return idx
}

func cleanValue(s string) string {
	return strings.TrimSpace(s)
}
