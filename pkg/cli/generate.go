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

// ColorDefinition holds a parsed color parameter
type ColorDefinition struct {
	Mode         generator.ColorMode // solid, gradient, tiled, noise
	Colors       []color.Color
	ColorStrings []string // original color strings for regenerating random colors
	Angle        float64  // for gradient
	TileSize     int      // for tiled/noise
	TextColor    *color.Color // optional text color override
}

// GenerateCommand handles the 'generate' command
type GenerateCommand struct {
	sizes           []string
	colorDefs       []ColorDefinition
	border          string
	text            string
	textSize        float64
	textColor       string
	textAngle       float64
	filename        string
	format          string
	rounds          int
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

	// Color flags (can be repeated) - solid color
	fs.Func("color", "Solid color: color[:t:textcolor]", func(s string) error {
		def, err := parseColorParameter(s, generator.ColorModeSolid)
		if err != nil {
			return err
		}
		c.colorDefs = append(c.colorDefs, def)
		return nil
	})
	fs.Func("c", "Solid color (shorthand)", func(s string) error {
		def, err := parseColorParameter(s, generator.ColorModeSolid)
		if err != nil {
			return err
		}
		c.colorDefs = append(c.colorDefs, def)
		return nil
	})

	// Gradient flags (can be repeated)
	fs.Func("gradient", "Gradient: color1,color2[,...][:angle][:t:textcolor]", func(s string) error {
		def, err := parseColorParameter(s, generator.ColorModeGradient)
		if err != nil {
			return err
		}
		c.colorDefs = append(c.colorDefs, def)
		return nil
	})
	fs.Func("g", "Gradient (shorthand)", func(s string) error {
		def, err := parseColorParameter(s, generator.ColorModeGradient)
		if err != nil {
			return err
		}
		c.colorDefs = append(c.colorDefs, def)
		return nil
	})

	// Tiles flags (can be repeated)
	fs.Func("tiles", "Tiles: color1,color2[,...][:tilesize][:t:textcolor]", func(s string) error {
		def, err := parseColorParameter(s, generator.ColorModeTiled)
		if err != nil {
			return err
		}
		c.colorDefs = append(c.colorDefs, def)
		return nil
	})
	fs.Func("t", "Tiles (shorthand)", func(s string) error {
		def, err := parseColorParameter(s, generator.ColorModeTiled)
		if err != nil {
			return err
		}
		c.colorDefs = append(c.colorDefs, def)
		return nil
	})

	// Noise flags (can be repeated)
	fs.Func("noise", "Noise: color1,color2[,...][:tilesize][:t:textcolor]", func(s string) error {
		def, err := parseColorParameter(s, generator.ColorModeNoise)
		if err != nil {
			return err
		}
		c.colorDefs = append(c.colorDefs, def)
		return nil
	})
	fs.Func("n", "Noise (shorthand)", func(s string) error {
		def, err := parseColorParameter(s, generator.ColorModeNoise)
		if err != nil {
			return err
		}
		c.colorDefs = append(c.colorDefs, def)
		return nil
	})

	// Border parameter (combined width and color)
	fs.StringVar(&c.border, "border", "", "Border: width,color")
	fs.StringVar(&c.border, "b", "", "Border (shorthand)")

	// Text parameters
	fs.StringVar(&c.text, "text", "{w}x{h}", "Text to display")
	fs.Float64Var(&c.textSize, "text-size", 20, "Text size in pt")
	fs.StringVar(&c.textColor, "text-color", "", "Default text color")
	fs.Float64Var(&c.textAngle, "text-angle", 0, "Text angle in degrees")

	// Output parameters
	fs.StringVar(&c.filename, "filename", "image.png", "Output filename")
	fs.StringVar(&c.filename, "f", "image.png", "Output filename (shorthand)")
	fs.StringVar(&c.format, "format", "png", "Output format (png, jpeg)")

	// Rounds parameter
	fs.IntVar(&c.rounds, "nr", 1, "Number of runs")
	fs.IntVar(&c.rounds, "r", 1, "Number of runs (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Set defaults if not provided
	if len(c.sizes) == 0 {
		c.sizes = []string{"256x192"}
	}
	if len(c.colorDefs) == 0 {
		// Default: solid gray
		c.colorDefs = []ColorDefinition{
			{
				Mode:   generator.ColorModeSolid,
				Colors: []color.Color{color.Gray{128}},
			},
		}
	}

	// Parse border if provided
	var borderWidth int
	var borderColor color.Color = color.Black
	if c.border != "" {
		parts := strings.Split(c.border, ",")
		if len(parts) != 2 {
			return fmt.Errorf("border must be in format width,color")
		}
		var err error
		borderWidth, err = strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return fmt.Errorf("invalid border width: %w", err)
		}
		borderColor, err = generator.ParseColor(strings.TrimSpace(parts[1]))
		if err != nil {
			return fmt.Errorf("invalid border color: %w", err)
		}
	}

	// Parse default text color if provided
	var defaultTextColor *color.Color
	if c.textColor != "" {
		col, err := generator.ParseColor(c.textColor)
		if err != nil {
			return fmt.Errorf("invalid text color: %w", err)
		}
		defaultTextColor = &col
	}

	// Generate images: sizes * color definitions * rounds
	imageCount := 0
	totalImages := len(c.sizes) * len(c.colorDefs) * c.rounds

	for round := 1; round <= c.rounds; round++ {
		for _, sizeStr := range c.sizes {
			width, height, err := parseSize(sizeStr)
			if err != nil {
				return fmt.Errorf("invalid size %s: %w", sizeStr, err)
			}

			for _, colorDef := range c.colorDefs {
				imageCount++

				// Regenerate random colors for each round (but not for the first round)
				actualColorDef := colorDef
				if round > 1 && hasRandomColor(colorDef) {
					// Re-parse the original color definition to get new random colors
					var err error
					actualColorDef, err = regenerateRandomColors(colorDef)
					if err != nil {
						return fmt.Errorf("failed to regenerate random colors: %w", err)
					}
				}

				// Create configuration
				config := generator.DefaultConfig()
				config.Width = width
				config.Height = height
				config.ColorMode = actualColorDef.Mode
				config.Colors = actualColorDef.Colors
				config.GradientAngle = actualColorDef.Angle
				config.TileSize = actualColorDef.TileSize
				config.Text = c.text
				config.TextSize = c.textSize
				config.TextAngle = c.textAngle
				config.Format = c.format
				config.BorderWidth = borderWidth
				config.BorderColor = borderColor

				// Text color priority: color parameter > default text color > auto
				if actualColorDef.TextColor != nil {
					config.TextColor = actualColorDef.TextColor
				} else if defaultTextColor != nil {
					config.TextColor = defaultTextColor
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
				fmt.Printf("Generated: %s (%dx%d, %s)\n", filename, width, height, colorDef.Mode)
			}
		}
	}

	return nil
}

// parseColorParameter parses a color parameter string based on the mode
// Format examples:
//   - solid: "blue" or "blue:t:white"
//   - gradient: "red,blue" or "red,blue:45" or "red,blue:45:t:white"
//   - tiles: "red,blue" or "red,blue:10" or "red,blue:10:t:white"
//   - noise: "red,blue,green" or "red,blue,green:10" or "red,blue,green:10:t:white"
func parseColorParameter(param string, mode generator.ColorMode) (ColorDefinition, error) {
	def := ColorDefinition{
		Mode:     mode,
		Angle:    0,
		TileSize: 36, // default tile size
	}

	// Check for text color override at the end (:t:color)
	textColorIdx := strings.LastIndex(param, ":t:")
	if textColorIdx != -1 {
		textColorStr := param[textColorIdx+3:]
		param = param[:textColorIdx]

		textCol, err := generator.ParseColor(textColorStr)
		if err != nil {
			return def, fmt.Errorf("invalid text color %s: %w", textColorStr, err)
		}
		def.TextColor = &textCol
	}

	switch mode {
	case generator.ColorModeSolid:
		// Format: color
		colorStr := strings.TrimSpace(param)
		def.ColorStrings = []string{colorStr}
		col, err := generator.ParseColor(colorStr)
		if err != nil {
			return def, fmt.Errorf("invalid color %s: %w", colorStr, err)
		}
		def.Colors = []color.Color{col}

	case generator.ColorModeGradient:
		// Format: color1,color2[,color3...][:angle]
		parts := strings.Split(param, ":")
		colorsPart := parts[0]

		// Parse angle if provided
		if len(parts) > 1 {
			angle, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return def, fmt.Errorf("invalid gradient angle %s: %w", parts[1], err)
			}
			def.Angle = angle
		}

		// Parse colors
		colorStrs := strings.Split(colorsPart, ",")
		if len(colorStrs) < 2 {
			return def, fmt.Errorf("gradient requires at least 2 colors")
		}

		def.ColorStrings = make([]string, len(colorStrs))
		def.Colors = make([]color.Color, 0, len(colorStrs))
		for i, colorStr := range colorStrs {
			colorStr = strings.TrimSpace(colorStr)
			def.ColorStrings[i] = colorStr
			col, err := generator.ParseColor(colorStr)
			if err != nil {
				return def, fmt.Errorf("invalid color %s: %w", colorStr, err)
			}
			def.Colors = append(def.Colors, col)
		}

	case generator.ColorModeTiled, generator.ColorModeNoise:
		// Format: color1,color2[,color3...][:tilesize]
		parts := strings.Split(param, ":")
		colorsPart := parts[0]

		// Parse tile size if provided
		if len(parts) > 1 {
			tileSize, err := strconv.Atoi(parts[1])
			if err != nil {
				return def, fmt.Errorf("invalid tile size %s: %w", parts[1], err)
			}
			def.TileSize = tileSize
		}

		// Parse colors
		colorStrs := strings.Split(colorsPart, ",")
		if len(colorStrs) < 2 {
			return def, fmt.Errorf("tiles/noise requires at least 2 colors")
		}

		def.ColorStrings = make([]string, len(colorStrs))
		def.Colors = make([]color.Color, 0, len(colorStrs))
		for i, colorStr := range colorStrs {
			colorStr = strings.TrimSpace(colorStr)
			def.ColorStrings[i] = colorStr
			col, err := generator.ParseColor(colorStr)
			if err != nil {
				return def, fmt.Errorf("invalid color %s: %w", colorStr, err)
			}
			def.Colors = append(def.Colors, col)
		}
	}

	return def, nil
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

// hasRandomColor checks if a color definition contains any "random" color strings
func hasRandomColor(def ColorDefinition) bool {
	for _, colorStr := range def.ColorStrings {
		if strings.ToLower(strings.TrimSpace(colorStr)) == "random" {
			return true
		}
	}
	return false
}

// regenerateRandomColors regenerates random colors in a color definition
func regenerateRandomColors(def ColorDefinition) (ColorDefinition, error) {
	newDef := def
	newDef.Colors = make([]color.Color, len(def.ColorStrings))

	for i, colorStr := range def.ColorStrings {
		col, err := generator.ParseColor(colorStr)
		if err != nil {
			return newDef, fmt.Errorf("invalid color %s: %w", colorStr, err)
		}
		newDef.Colors[i] = col
	}

	return newDef, nil
}
