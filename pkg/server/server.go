package server

import (
	"fmt"
	"image/color"
	"log"
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

// parseURLConfig parses the URL path and returns an ImageConfig
// URL format: /[size]/c:[color-config]/t:[text]/f:[format]/b:[border]
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
		case 'c': // color configuration
			if err := parseColorConfig(config, value); err != nil {
				return nil, fmt.Errorf("invalid color config: %w", err)
			}
		case 't': // text configuration
			if err := parseTextConfig(config, value); err != nil {
				return nil, fmt.Errorf("invalid text config: %w", err)
			}
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

	return config, nil
}

// parseColorConfig parses color configuration
// Format: [mode],[color1],[color2],...,a:[angle],ts:[tilesize]
func parseColorConfig(config *generator.ImageConfig, value string) error {
	// Unquote if quoted
	value = strings.Trim(value, "\"")

	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return fmt.Errorf("empty color config")
	}

	// First part might be mode or color
	firstPart := parts[0]

	// Check if it's a mode
	if firstPart == "solid" || firstPart == "tiled" || firstPart == "gradient" || firstPart == "noise" {
		config.ColorMode = generator.ColorMode(firstPart)
		parts = parts[1:]
	}

	// Parse remaining parts
	config.Colors = []color.Color{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for special parameters
		if strings.HasPrefix(part, "a:") {
			angle, err := strconv.ParseFloat(part[2:], 64)
			if err != nil {
				return fmt.Errorf("invalid angle: %w", err)
			}
			config.GradientAngle = angle
			continue
		}

		if strings.HasPrefix(part, "ts:") {
			tileSize, err := strconv.Atoi(part[3:])
			if err != nil {
				return fmt.Errorf("invalid tile size: %w", err)
			}
			config.TileSize = tileSize
			continue
		}

		// Parse as color
		col, err := generator.ParseColor(part)
		if err != nil {
			return fmt.Errorf("invalid color %s: %w", part, err)
		}
		config.Colors = append(config.Colors, col)
	}

	// Default to gray if no colors specified
	if len(config.Colors) == 0 {
		col, _ := generator.ParseColor("gray")
		config.Colors = []color.Color{col}
	}

	return nil
}

// parseTextConfig parses text configuration
// Format: "text",s:[size],c:[color],a:[angle]
func parseTextConfig(config *generator.ImageConfig, value string) error {
	// Unquote if quoted
	value = strings.Trim(value, "\"")

	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return fmt.Errorf("empty text config")
	}

	// First part is the text
	config.Text = parts[0]

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

// parseBorderConfig parses border configuration
// Format: [width],[color]
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
