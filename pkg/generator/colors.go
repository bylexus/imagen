package generator

import (
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"strings"
)

// ParseColor parses a color string and returns a color.Color
// Supports: color names (blue, red, etc.), hex codes (RRGGBB or #RRGGBB), "random"
func ParseColor(colorStr string) (color.Color, error) {
	colorStr = strings.TrimSpace(strings.ToLower(colorStr))

	// Handle "random"
	if colorStr == "random" {
		return color.RGBA{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 255,
		}, nil
	}

	// Remove # prefix if present
	colorStr = strings.TrimPrefix(colorStr, "#")

	// Try to parse as hex
	if len(colorStr) == 6 {
		r, err1 := strconv.ParseUint(colorStr[0:2], 16, 8)
		g, err2 := strconv.ParseUint(colorStr[2:4], 16, 8)
		b, err3 := strconv.ParseUint(colorStr[4:6], 16, 8)
		if err1 == nil && err2 == nil && err3 == nil {
			return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
		}
	}

	// Try named colors
	if c, ok := namedColors[colorStr]; ok {
		return c, nil
	}

	return nil, fmt.Errorf("invalid color: %s", colorStr)
}

// namedColors is a map of common color names to their RGB values
var namedColors = map[string]color.Color{
	"black":   color.RGBA{0, 0, 0, 255},
	"white":   color.RGBA{255, 255, 255, 255},
	"red":     color.RGBA{255, 0, 0, 255},
	"green":   color.RGBA{0, 255, 0, 255},
	"blue":    color.RGBA{0, 0, 255, 255},
	"yellow":  color.RGBA{255, 255, 0, 255},
	"cyan":    color.RGBA{0, 255, 255, 255},
	"magenta": color.RGBA{255, 0, 255, 255},
	"gray":    color.RGBA{128, 128, 128, 255},
	"grey":    color.RGBA{128, 128, 128, 255},
	"orange":  color.RGBA{255, 165, 0, 255},
	"purple":  color.RGBA{128, 0, 128, 255},
	"pink":    color.RGBA{255, 192, 203, 255},
	"brown":   color.RGBA{165, 42, 42, 255},
}

// InterpolateColor interpolates between two colors based on factor (0.0 to 1.0)
func InterpolateColor(c1, c2 color.Color, factor float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	// Convert from 16-bit to 8-bit
	r1, g1, b1, a1 = r1>>8, g1>>8, b1>>8, a1>>8
	r2, g2, b2, a2 = r2>>8, g2>>8, b2>>8, a2>>8

	r := uint8(float64(r1) + factor*float64(int(r2)-int(r1)))
	g := uint8(float64(g1) + factor*float64(int(g2)-int(g1)))
	b := uint8(float64(b1) + factor*float64(int(b2)-int(b1)))
	a := uint8(float64(a1) + factor*float64(int(a2)-int(a1)))

	return color.RGBA{R: r, G: g, B: b, A: a}
}
