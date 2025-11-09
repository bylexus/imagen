package cli

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/bylexus/imagen/pkg/generator"
)

// GenerateCommand handles the 'generate' command
type GenerateCommand struct {
	sizes        []string
	colorModes   []string
	colors       []string
	borderWidth  int
	borderColor  string
	gradientAngle float64
	tileSize     int
	text         string
	textSize     float64
	textColor    string
	filename     string
	format       string
}

// Execute runs the generate command
func (c *GenerateCommand) Execute(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)

	// Size flag (can be repeated)
	fs.Func("size", "Image size (WxH), can be repeated", func(s string) error {
		c.sizes = append(c.sizes, s)
		return nil
	})
	fs.Func("s", "Image size (WxH), can be repeated (shorthand)", func(s string) error {
		c.sizes = append(c.sizes, s)
		return nil
	})

	// Color mode flag (can be repeated)
	fs.Func("color-mode", "Color mode: solid, tiled, gradient, noise (can be repeated)", func(s string) error {
		c.colorModes = append(c.colorModes, s)
		return nil
	})
	fs.Func("m", "Color mode (shorthand)", func(s string) error {
		c.colorModes = append(c.colorModes, s)
		return nil
	})

	// Color flag (can be repeated)
	fs.Func("color", "Color value (can be repeated)", func(s string) error {
		c.colors = append(c.colors, s)
		return nil
	})
	fs.Func("c", "Color value (shorthand)", func(s string) error {
		c.colors = append(c.colors, s)
		return nil
	})

	fs.IntVar(&c.borderWidth, "border-width", 0, "Border width in pixels")
	fs.IntVar(&c.borderWidth, "b", 0, "Border width (shorthand)")
	fs.StringVar(&c.borderColor, "border-color", "black", "Border color")

	fs.Float64Var(&c.gradientAngle, "gradient-angle", 0, "Gradient angle in degrees")
	fs.Float64Var(&c.gradientAngle, "a", 0, "Gradient angle (shorthand)")

	fs.IntVar(&c.tileSize, "tile-size", 16, "Tile size in pixels")
	fs.IntVar(&c.tileSize, "ts", 16, "Tile size (shorthand)")

	fs.StringVar(&c.text, "text", "{w}x{h}", "Text to display")
	fs.StringVar(&c.text, "t", "{w}x{h}", "Text to display (shorthand)")

	fs.Float64Var(&c.textSize, "text-size", 20, "Text size in pt")
	fs.StringVar(&c.textColor, "text-color", "", "Text color")

	fs.StringVar(&c.filename, "filename", "image.png", "Output filename")
	fs.StringVar(&c.filename, "f", "image.png", "Output filename (shorthand)")

	fs.StringVar(&c.format, "format", "png", "Output format (png, jpeg)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Set defaults if not provided
	if len(c.sizes) == 0 {
		c.sizes = []string{"256x192"}
	}
	if len(c.colorModes) == 0 {
		c.colorModes = []string{"solid"}
	}
	if len(c.colors) == 0 {
		c.colors = []string{"gray"}
	}

	// Generate images for each combination of size and color mode
	imageCount := 0
	totalImages := len(c.sizes) * len(c.colorModes)

	for _, sizeStr := range c.sizes {
		width, height, err := parseSize(sizeStr)
		if err != nil {
			return fmt.Errorf("invalid size %s: %w", sizeStr, err)
		}

		for _, mode := range c.colorModes {
			imageCount++

			// Create configuration
			config := generator.DefaultConfig()
			config.Width = width
			config.Height = height
			config.ColorMode = generator.ColorMode(mode)
			config.GradientAngle = c.gradientAngle
			config.TileSize = c.tileSize
			config.Text = c.text
			config.TextSize = c.textSize
			config.Format = c.format
			config.BorderWidth = c.borderWidth

			// Parse colors
			config.Colors = make([]color.Color, 0, len(c.colors))
			for _, colorStr := range c.colors {
				col, err := generator.ParseColor(colorStr)
				if err != nil {
					return fmt.Errorf("invalid color %s: %w", colorStr, err)
				}
				config.Colors = append(config.Colors, col)
			}

			// Parse border color
			if c.borderColor != "" {
				borderCol, err := generator.ParseColor(c.borderColor)
				if err != nil {
					return fmt.Errorf("invalid border color %s: %w", c.borderColor, err)
				}
				config.BorderColor = borderCol
			}

			// Parse text color
			if c.textColor != "" {
				textCol, err := generator.ParseColor(c.textColor)
				if err != nil {
					return fmt.Errorf("invalid text color %s: %w", c.textColor, err)
				}
				config.TextColor = &textCol
			}

			// Generate filename
			filename := c.filename
			if totalImages > 1 {
				// Add numbering for multiple images
				ext := ""
				if idx := strings.LastIndex(filename, "."); idx != -1 {
					ext = filename[idx:]
					filename = filename[:idx]
				}
				filename = fmt.Sprintf("%s-%04d%s", filename, imageCount, ext)
			}

			// Replace placeholders in filename
			filename = strings.ReplaceAll(filename, "{w}", strconv.Itoa(width))
			filename = strings.ReplaceAll(filename, "{h}", strconv.Itoa(height))
			filename = strings.ReplaceAll(filename, "{nr}", strconv.Itoa(imageCount))

			// Generate image
			gen := generator.NewGenerator(config)
			img, err := gen.Generate()
			if err != nil {
				return fmt.Errorf("failed to generate image: %w", err)
			}

			// Write to file
			file, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", filename, err)
			}

			if err := gen.WriteImage(file, img); err != nil {
				file.Close()
				return fmt.Errorf("failed to write image to %s: %w", filename, err)
			}

			file.Close()
			fmt.Printf("Generated: %s (%dx%d, %s)\n", filename, width, height, mode)
		}
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
