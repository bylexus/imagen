package server

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/bylexus/imagen/pkg/generator"
)

// Server represents the HTTP server for serving images
type Server struct {
	addresses []string
}

// NewServer creates a new server with the given listen addresses
func NewServer(addresses []string) *Server {
	if len(addresses) == 0 {
		addresses = []string{":3000"}
	}
	return &Server{addresses: addresses}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	http.HandleFunc("/", s.handleImageRequest)

	// Start listeners
	errChan := make(chan error, len(s.addresses))

	for _, addr := range s.addresses {
		addr := addr // capture for goroutine
		go func() {
			log.Printf("Starting server on %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				errChan <- fmt.Errorf("failed to start server on %s: %w", addr, err)
			}
		}()
	}

	// Wait for any error
	return <-errChan
}

// handleImageRequest handles HTTP requests and generates images based on URL parameters
func (s *Server) handleImageRequest(w http.ResponseWriter, r *http.Request) {
	config, err := parseURLConfig(r.URL.Path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid URL: %v", err), http.StatusBadRequest)
		return
	}

	// Generate image
	gen := generator.NewGenerator(config)
	img, err := gen.Generate()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate image: %v", err), http.StatusInternalServerError)
		return
	}

	// Set content type based on format
	switch config.Format {
	case "png":
		w.Header().Set("Content-Type", "image/png")
	case "jpeg", "jpg":
		w.Header().Set("Content-Type", "image/jpeg")
	default:
		w.Header().Set("Content-Type", "image/png")
	}

	// Write image
	if err := gen.WriteImage(w, img); err != nil {
		log.Printf("Failed to write image: %v", err)
	}
}

// ColorDefinition holds a parsed background color configuration
type ColorDefinition struct {
	Mode      generator.ColorMode
	Colors    []color.Color
	Angle     float64
	TileSize  int
	TextColor *color.Color
}

// parseURLConfig parses the URL path and returns an ImageConfig
// URL format: /[size]/[c|g|t|n]:[color-config]/t:[text]/f:[format]/b:[border]
func parseURLConfig(path string) (*generator.ImageConfig, error) {
	config := generator.DefaultConfig()

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Empty path means default image
	if path == "" {
		return config, nil
	}

	// Split by /
	parts := strings.Split(path, "/")

	// Collect all color definitions for random selection
	var colorDefs []ColorDefinition

	for i, part := range parts {
		if part == "" {
			continue
		}

		// First part without prefix is size
		if i == 0 && !strings.Contains(part, ":") {
			width, height, err := parseSize(part)
			if err != nil {
				return nil, fmt.Errorf("invalid size: %w", err)
			}
			config.Width = width
			config.Height = height
			continue
		}

		// Parse prefixed parameters
		if len(part) < 2 || part[1] != ':' {
			return nil, fmt.Errorf("invalid parameter format: %s", part)
		}

		prefix := part[0]
		value := part[2:]

		switch prefix {
		case 'c': // solid color background
			colorDef, err := parseSolidBackground(value)
			if err != nil {
				return nil, fmt.Errorf("invalid solid background: %w", err)
			}
			colorDefs = append(colorDefs, colorDef)
		case 'g': // gradient background
			colorDef, err := parseGradientBackground(value)
			if err != nil {
				return nil, fmt.Errorf("invalid gradient background: %w", err)
			}
			colorDefs = append(colorDefs, colorDef)
		case 't': // tiled background OR text
			// Determine if this is a tiled background or text
			// Text starts with a quote, tiled starts with a color
			if strings.HasPrefix(value, "\"") {
				// This is text
				if err := parseTextConfig(config, value); err != nil {
					return nil, fmt.Errorf("invalid text config: %w", err)
				}
			} else {
				// This is tiled background
				colorDef, err := parseTiledBackground(value)
				if err != nil {
					return nil, fmt.Errorf("invalid tiled background: %w", err)
				}
				colorDefs = append(colorDefs, colorDef)
			}
		case 'n': // noise background
			colorDef, err := parseNoiseBackground(value)
			if err != nil {
				return nil, fmt.Errorf("invalid noise background: %w", err)
			}
			colorDefs = append(colorDefs, colorDef)
		case 'f': // format
			config.Format = value
		case 'b': // border
			if err := parseBorderConfig(config, value); err != nil {
				return nil, fmt.Errorf("invalid border config: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown parameter prefix: %c", prefix)
		}
	}

	// If we have color definitions, randomly select one
	if len(colorDefs) > 0 {
		selectedDef := colorDefs[rand.Intn(len(colorDefs))]
		config.ColorMode = selectedDef.Mode
		config.Colors = selectedDef.Colors
		config.GradientAngle = selectedDef.Angle
		config.TileSize = selectedDef.TileSize
		if selectedDef.TextColor != nil {
			config.TextColor = selectedDef.TextColor
		}
	}

	return config, nil
}

// parseSolidBackground parses solid color background
// Format: c:[color][:t:[textcolor]]
func parseSolidBackground(value string) (ColorDefinition, error) {
	def := ColorDefinition{
		Mode:     generator.ColorModeSolid,
		Colors:   []color.Color{},
		TileSize: 16,
	}

	// Split by : to separate color from optional text color
	parts := strings.Split(value, ":t:")
	if len(parts) > 2 {
		return def, fmt.Errorf("invalid solid background format")
	}

	// Parse the main color
	col, err := generator.ParseColor(parts[0])
	if err != nil {
		return def, fmt.Errorf("invalid color: %w", err)
	}
	def.Colors = append(def.Colors, col)

	// Parse optional text color
	if len(parts) == 2 {
		textCol, err := generator.ParseColor(parts[1])
		if err != nil {
			return def, fmt.Errorf("invalid text color: %w", err)
		}
		def.TextColor = &textCol
	}

	return def, nil
}

// parseGradientBackground parses gradient background
// Format: g:[color1],[color2][,[color3]...][:angle][:t:[textcolor]]
func parseGradientBackground(value string) (ColorDefinition, error) {
	def := ColorDefinition{
		Mode:     generator.ColorModeGradient,
		Colors:   []color.Color{},
		Angle:    0,
		TileSize: 16,
	}

	// Split by :t: to separate main config from optional text color
	parts := strings.Split(value, ":t:")
	if len(parts) > 2 {
		return def, fmt.Errorf("invalid gradient background format")
	}

	mainPart := parts[0]

	// Parse optional text color
	if len(parts) == 2 {
		textCol, err := generator.ParseColor(parts[1])
		if err != nil {
			return def, fmt.Errorf("invalid text color: %w", err)
		}
		def.TextColor = &textCol
	}

	// Split main part by : to separate colors from optional angle
	colorAndAngle := strings.Split(mainPart, ":")
	if len(colorAndAngle) > 2 {
		return def, fmt.Errorf("invalid gradient format")
	}

	// Parse colors (comma-separated)
	colorsPart := colorAndAngle[0]
	colorStrs := strings.Split(colorsPart, ",")
	if len(colorStrs) < 2 {
		return def, fmt.Errorf("gradient requires at least 2 colors")
	}

	for _, colorStr := range colorStrs {
		col, err := generator.ParseColor(strings.TrimSpace(colorStr))
		if err != nil {
			return def, fmt.Errorf("invalid color %s: %w", colorStr, err)
		}
		def.Colors = append(def.Colors, col)
	}

	// Parse optional angle
	if len(colorAndAngle) == 2 {
		angle, err := strconv.ParseFloat(colorAndAngle[1], 64)
		if err != nil {
			return def, fmt.Errorf("invalid angle: %w", err)
		}
		def.Angle = angle
	}

	return def, nil
}

// parseTiledBackground parses tiled background
// Format: t:[color1],[color2][,[color3]...][:tilesize][:t:[textcolor]]
func parseTiledBackground(value string) (ColorDefinition, error) {
	def := ColorDefinition{
		Mode:     generator.ColorModeTiled,
		Colors:   []color.Color{},
		TileSize: 36, // default from README
	}

	// Split by :t: to separate main config from optional text color
	parts := strings.Split(value, ":t:")
	if len(parts) > 2 {
		return def, fmt.Errorf("invalid tiled background format")
	}

	mainPart := parts[0]

	// Parse optional text color
	if len(parts) == 2 {
		textCol, err := generator.ParseColor(parts[1])
		if err != nil {
			return def, fmt.Errorf("invalid text color: %w", err)
		}
		def.TextColor = &textCol
	}

	// Split main part by : to separate colors from optional tile size
	colorAndSize := strings.Split(mainPart, ":")
	if len(colorAndSize) > 2 {
		return def, fmt.Errorf("invalid tiled format")
	}

	// Parse colors (comma-separated)
	colorsPart := colorAndSize[0]
	colorStrs := strings.Split(colorsPart, ",")
	if len(colorStrs) < 2 {
		return def, fmt.Errorf("tiled requires at least 2 colors")
	}

	for _, colorStr := range colorStrs {
		col, err := generator.ParseColor(strings.TrimSpace(colorStr))
		if err != nil {
			return def, fmt.Errorf("invalid color %s: %w", colorStr, err)
		}
		def.Colors = append(def.Colors, col)
	}

	// Parse optional tile size
	if len(colorAndSize) == 2 {
		size, err := strconv.Atoi(colorAndSize[1])
		if err != nil {
			return def, fmt.Errorf("invalid tile size: %w", err)
		}
		def.TileSize = size
	}

	return def, nil
}

// parseNoiseBackground parses noise background
// Format: n:[color1],[color2][,[color3]...][:tilesize][:t:[textcolor]]
func parseNoiseBackground(value string) (ColorDefinition, error) {
	def := ColorDefinition{
		Mode:     generator.ColorModeNoise,
		Colors:   []color.Color{},
		TileSize: 36, // default from README
	}

	// Split by :t: to separate main config from optional text color
	parts := strings.Split(value, ":t:")
	if len(parts) > 2 {
		return def, fmt.Errorf("invalid noise background format")
	}

	mainPart := parts[0]

	// Parse optional text color
	if len(parts) == 2 {
		textCol, err := generator.ParseColor(parts[1])
		if err != nil {
			return def, fmt.Errorf("invalid text color: %w", err)
		}
		def.TextColor = &textCol
	}

	// Split main part by : to separate colors from optional tile size
	colorAndSize := strings.Split(mainPart, ":")
	if len(colorAndSize) > 2 {
		return def, fmt.Errorf("invalid noise format")
	}

	// Parse colors (comma-separated)
	colorsPart := colorAndSize[0]
	colorStrs := strings.Split(colorsPart, ",")
	if len(colorStrs) < 2 {
		return def, fmt.Errorf("noise requires at least 2 colors")
	}

	for _, colorStr := range colorStrs {
		col, err := generator.ParseColor(strings.TrimSpace(colorStr))
		if err != nil {
			return def, fmt.Errorf("invalid color %s: %w", colorStr, err)
		}
		def.Colors = append(def.Colors, col)
	}

	// Parse optional tile size
	if len(colorAndSize) == 2 {
		size, err := strconv.Atoi(colorAndSize[1])
		if err != nil {
			return def, fmt.Errorf("invalid tile size: %w", err)
		}
		def.TileSize = size
	}

	return def, nil
}

// parseTextConfig parses text configuration
// Format: t:"text"[,s:size][,c:color][,a:angle]
func parseTextConfig(config *generator.ImageConfig, value string) error {
	// Split by comma while respecting quotes
	parts := splitRespectingQuotes(value)
	if len(parts) == 0 {
		return fmt.Errorf("empty text config")
	}

	// First part is the text (remove quotes if present)
	config.Text = strings.Trim(parts[0], "\"")

	// Parse remaining parts
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if len(part) < 2 || part[1] != ':' {
			return fmt.Errorf("invalid text parameter: %s", part)
		}

		prefix := part[0]
		val := part[2:]

		switch prefix {
		case 's': // size
			size, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("invalid text size: %w", err)
			}
			config.TextSize = size
		case 'c': // color
			col, err := generator.ParseColor(val)
			if err != nil {
				return fmt.Errorf("invalid text color: %w", err)
			}
			config.TextColor = &col
		case 'a': // angle
			angle, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("invalid text angle: %w", err)
			}
			config.TextAngle = angle
		default:
			return fmt.Errorf("unknown text parameter: %c", prefix)
		}
	}

	return nil
}

// splitRespectingQuotes splits a string by commas while respecting quoted sections
func splitRespectingQuotes(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		char := s[i]

		if char == '"' {
			inQuotes = !inQuotes
			current.WriteByte(char)
		} else if char == ',' && !inQuotes {
			// Split here
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(char)
		}
	}

	// Add the last part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseBorderConfig parses border configuration
// Format: b:width,color
func parseBorderConfig(config *generator.ImageConfig, value string) error {
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return fmt.Errorf("empty border config")
	}

	// First part is width
	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid border width: %w", err)
	}
	config.BorderWidth = width

	// Second part (if present) is color
	if len(parts) > 1 {
		col, err := generator.ParseColor(parts[1])
		if err != nil {
			return fmt.Errorf("invalid border color: %w", err)
		}
		config.BorderColor = col
	}

	return nil
}

// parseSize parses a size string in format "WxH"
func parseSize(sizeStr string) (width, height int, err error) {
	parts := strings.Split(sizeStr, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("size must be in format WxH")
	}

	width, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width: %w", err)
	}

	height, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height: %w", err)
	}

	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("width and height must be positive")
	}

	return width, height, nil
}
