// Package model contains domain types for the smartphone search engine.
package model

import (
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// Smartphone represents a phone from the GSMArena dataset.
type Smartphone struct {
	Brand      string `json:"brand"`
	Model      string `json:"model"`
	ImageURL   string `json:"image_url"`
	ImageFile  string `json:"image_file"`
	Technology string `json:"technology"`
	Announced  string `json:"announced"`
	Status     string `json:"status"`
	Dimensions string `json:"dimensions"`
	Weight     string `json:"weight"`
	SIM        string `json:"sim"`
	Display    string `json:"display"`
	ScreenSize string `json:"screen_size"`
	Resolution string `json:"resolution"`
	Protection string `json:"protection"`
	OS         string `json:"os"`
	Chipset    string `json:"chipset"`
	CPU        string `json:"cpu"`
	GPU        string `json:"gpu"`
	CardSlot   string `json:"card_slot"`
	Storage    string `json:"storage"`
	Camera     string `json:"camera"`
	Video      string `json:"video"`
	Selfie     string `json:"selfie"`
	Battery    string `json:"battery"`
	Charging   string `json:"charging"`
	WLAN       string `json:"wlan"`
	Bluetooth  string `json:"bluetooth"`
	GPS        string `json:"gps"`
	NFC        string `json:"nfc"`
	USB        string `json:"usb"`
	Sensors    string `json:"sensors"`
	Colors     string  `json:"colors"`
	Price      string  `json:"price"`
	Score      float32 `json:"score,omitempty"`
}

var eurPriceRe = regexp.MustCompile(`(\d+(?:\.\d{1,2})?)\s*EUR`)

// parseEURPrice extracts the first EUR price from a string like "About 130 EUR".
func parseEURPrice(s string) float64 {
	m := eurPriceRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return 0
	}

	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}

	return v
}

// classifyOS normalizes the raw OS string into a family bucket.
func classifyOS(s string) string {
	low := strings.ToLower(s)

	switch {
	case strings.Contains(low, "android"):
		return "Android"
	case strings.Contains(low, "ios"):
		return "iOS"
	case strings.Contains(low, "windows"):
		return "Windows"
	default:
		return "Other"
	}
}

// classifyDisplay normalizes the raw display string into a type bucket.
func classifyDisplay(s string) string {
	up := strings.ToUpper(s)

	switch {
	case strings.Contains(up, "AMOLED"):
		return "AMOLED"
	case strings.Contains(up, "OLED"):
		return "OLED"
	case strings.Contains(up, "IPS"):
		return "IPS"
	case strings.Contains(up, "TFT"):
		return "TFT"
	case strings.Contains(up, "LCD"):
		return "LCD"
	default:
		return "Other"
	}
}

// ImageFilename extracts the filename from the image URL.
func (s Smartphone) ImageFilename() string {
	u, err := url.Parse(s.ImageURL)
	if err != nil || u.Path == "" {
		return ""
	}

	return path.Base(u.Path)
}

// Description builds a searchable text from the phone specs.
func (s Smartphone) Description() string {
	var b strings.Builder

	b.WriteString(s.Brand)
	b.WriteString(" ")
	b.WriteString(s.Model)
	b.WriteString(". Network: ")
	b.WriteString(s.Technology)
	b.WriteString(". Display: ")
	b.WriteString(s.Display)
	b.WriteString(" ")
	b.WriteString(s.ScreenSize)
	b.WriteString(" ")
	b.WriteString(s.Resolution)
	b.WriteString(". Protection: ")
	b.WriteString(s.Protection)
	b.WriteString(". OS: ")
	b.WriteString(s.OS)
	b.WriteString(". Chipset: ")
	b.WriteString(s.Chipset)
	b.WriteString(" ")
	b.WriteString(s.CPU)
	b.WriteString(". GPU: ")
	b.WriteString(s.GPU)
	b.WriteString(". Storage: ")
	b.WriteString(s.Storage)
	b.WriteString(". Card slot: ")
	b.WriteString(s.CardSlot)
	b.WriteString(". Camera: ")
	b.WriteString(s.Camera)
	b.WriteString(". Selfie: ")
	b.WriteString(s.Selfie)
	b.WriteString(". Battery: ")
	b.WriteString(s.Battery)
	b.WriteString(" ")
	b.WriteString(s.Charging)
	b.WriteString(". Dimensions: ")
	b.WriteString(s.Dimensions)
	b.WriteString(". Weight: ")
	b.WriteString(s.Weight)
	b.WriteString(". SIM: ")
	b.WriteString(s.SIM)
	b.WriteString(". NFC: ")
	b.WriteString(s.NFC)
	b.WriteString(". Colors: ")
	b.WriteString(s.Colors)
	b.WriteString(". Price: ")
	b.WriteString(s.Price)

	return b.String()
}

// PayloadMap returns the smartphone data as a map for Qdrant payload.
func (s Smartphone) PayloadMap() map[string]any {
	return map[string]any{
		"brand":       s.Brand,
		"model":       s.Model,
		"image_url":   s.ImageURL,
		"image_file":  s.ImageFile,
		"technology":  s.Technology,
		"announced":   s.Announced,
		"status":      s.Status,
		"dimensions":  s.Dimensions,
		"weight":      s.Weight,
		"sim":         s.SIM,
		"display":     s.Display,
		"screen_size": s.ScreenSize,
		"resolution":  s.Resolution,
		"protection":  s.Protection,
		"os":          s.OS,
		"chipset":     s.Chipset,
		"cpu":         s.CPU,
		"gpu":         s.GPU,
		"card_slot":   s.CardSlot,
		"storage":     s.Storage,
		"camera":      s.Camera,
		"video":       s.Video,
		"selfie":      s.Selfie,
		"battery":     s.Battery,
		"charging":    s.Charging,
		"wlan":        s.WLAN,
		"bluetooth":   s.Bluetooth,
		"gps":         s.GPS,
		"nfc":         s.NFC,
		"usb":         s.USB,
		"sensors":     s.Sensors,
		"colors":      s.Colors,
		"price":       s.Price,
		"description":  s.Description(),
		"os_family":    classifyOS(s.OS),
		"display_type": classifyDisplay(s.Display),
		"price_eur":    parseEURPrice(s.Price),
	}
}
