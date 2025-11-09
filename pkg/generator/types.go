package generator

import (
	"image/color"
)

// ColorMode represents the background color mode
type ColorMode string

const (
	ColorModeSolid    ColorMode = "solid"
	ColorModeTiled    ColorMode = "tiled"
	ColorModeGradient ColorMode = "gradient"
	ColorModeNoise    ColorMode = "noise"
)

// ImageConfig holds the configuration for generating an image
type ImageConfig struct {
	Width         int
	Height        int
	ColorMode     ColorMode
	Colors        []color.Color
	GradientAngle float64
	TileSize      int
	Text          string
	TextSize      float64
	TextColor     *color.Color
	TextAngle     float64
	FontName      string
	BorderWidth   int
	BorderColor   color.Color
	Format        string // png, jpeg, webp
}

// DefaultConfig returns a default image configuration
func DefaultConfig() *ImageConfig {
	return &ImageConfig{
		Width:         256,
		Height:        192,
		ColorMode:     ColorModeSolid,
		Colors:        []color.Color{color.Gray{128}},
		GradientAngle: 0,
		TileSize:      16,
		Text:          "{w}x{h}",
		TextSize:      20,
		TextColor:     nil, // nil means auto (white or XOR)
		TextAngle:     0,
		FontName:      "",
		BorderWidth:   0,
		BorderColor:   color.Black,
		Format:        "png",
	}
}
