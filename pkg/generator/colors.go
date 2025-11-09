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

// namedColors is a map of all 140 HTML color names to their RGB values
// Based on W3C HTML color specification: https://www.w3.org/TR/css-color-3/#html4
var namedColors = map[string]color.Color{
	// Red colors
	"indianred":  color.RGBA{0xCD, 0x5C, 0x5C, 255},
	"lightcoral": color.RGBA{0xF0, 0x80, 0x80, 255},
	"salmon":     color.RGBA{0xFA, 0x80, 0x72, 255},
	"darksalmon": color.RGBA{0xE9, 0x96, 0x7A, 255},
	"lightsalmon": color.RGBA{0xFF, 0xA0, 0x7A, 255},
	"crimson":    color.RGBA{0xDC, 0x14, 0x3C, 255},
	"red":        color.RGBA{0xFF, 0x00, 0x00, 255},
	"firebrick":  color.RGBA{0xB2, 0x22, 0x22, 255},
	"darkred":    color.RGBA{0x8B, 0x00, 0x00, 255},

	// Pink colors
	"pink":            color.RGBA{0xFF, 0xC0, 0xCB, 255},
	"lightpink":       color.RGBA{0xFF, 0xB6, 0xC1, 255},
	"hotpink":         color.RGBA{0xFF, 0x69, 0xB4, 255},
	"deeppink":        color.RGBA{0xFF, 0x14, 0x93, 255},
	"mediumvioletred": color.RGBA{0xC7, 0x15, 0x85, 255},
	"palevioletred":   color.RGBA{0xDB, 0x70, 0x93, 255},

	// Orange colors
	"coral":      color.RGBA{0xFF, 0x7F, 0x50, 255},
	"tomato":     color.RGBA{0xFF, 0x63, 0x47, 255},
	"orangered":  color.RGBA{0xFF, 0x45, 0x00, 255},
	"darkorange": color.RGBA{0xFF, 0x8C, 0x00, 255},
	"orange":     color.RGBA{0xFF, 0xA5, 0x00, 255},

	// Yellow colors
	"gold":                  color.RGBA{0xFF, 0xD7, 0x00, 255},
	"yellow":                color.RGBA{0xFF, 0xFF, 0x00, 255},
	"lightyellow":           color.RGBA{0xFF, 0xFF, 0xE0, 255},
	"lemonchiffon":          color.RGBA{0xFF, 0xFA, 0xCD, 255},
	"lightgoldenrodyellow":  color.RGBA{0xFA, 0xFA, 0xD2, 255},
	"papayawhip":            color.RGBA{0xFF, 0xEF, 0xD5, 255},
	"moccasin":              color.RGBA{0xFF, 0xE4, 0xB5, 255},
	"peachpuff":             color.RGBA{0xFF, 0xDA, 0xB9, 255},
	"palegoldenrod":         color.RGBA{0xEE, 0xE8, 0xAA, 255},
	"khaki":                 color.RGBA{0xF0, 0xE6, 0x8C, 255},
	"darkkhaki":             color.RGBA{0xBD, 0xB7, 0x6B, 255},

	// Purple colors
	"lavender":          color.RGBA{0xE6, 0xE6, 0xFA, 255},
	"thistle":           color.RGBA{0xD8, 0xBF, 0xD8, 255},
	"plum":              color.RGBA{0xDD, 0xA0, 0xDD, 255},
	"violet":            color.RGBA{0xEE, 0x82, 0xEE, 255},
	"orchid":            color.RGBA{0xDA, 0x70, 0xD6, 255},
	"fuchsia":           color.RGBA{0xFF, 0x00, 0xFF, 255},
	"magenta":           color.RGBA{0xFF, 0x00, 0xFF, 255},
	"mediumorchid":      color.RGBA{0xBA, 0x55, 0xD3, 255},
	"mediumpurple":      color.RGBA{0x93, 0x70, 0xDB, 255},
	"rebeccapurple":     color.RGBA{0x66, 0x33, 0x99, 255},
	"blueviolet":        color.RGBA{0x8A, 0x2B, 0xE2, 255},
	"darkviolet":        color.RGBA{0x94, 0x00, 0xD3, 255},
	"darkorchid":        color.RGBA{0x99, 0x32, 0xCC, 255},
	"darkmagenta":       color.RGBA{0x8B, 0x00, 0x8B, 255},
	"purple":            color.RGBA{0x80, 0x00, 0x80, 255},
	"indigo":            color.RGBA{0x4B, 0x00, 0x82, 255},
	"slateblue":         color.RGBA{0x6A, 0x5A, 0xCD, 255},
	"darkslateblue":     color.RGBA{0x48, 0x3D, 0x8B, 255},
	"mediumslateblue":   color.RGBA{0x7B, 0x68, 0xEE, 255},

	// Green colors
	"greenyellow":       color.RGBA{0xAD, 0xFF, 0x2F, 255},
	"chartreuse":        color.RGBA{0x7F, 0xFF, 0x00, 255},
	"lawngreen":         color.RGBA{0x7C, 0xFC, 0x00, 255},
	"lime":              color.RGBA{0x00, 0xFF, 0x00, 255},
	"limegreen":         color.RGBA{0x32, 0xCD, 0x32, 255},
	"palegreen":         color.RGBA{0x98, 0xFB, 0x98, 255},
	"lightgreen":        color.RGBA{0x90, 0xEE, 0x90, 255},
	"mediumspringgreen": color.RGBA{0x00, 0xFA, 0x9A, 255},
	"springgreen":       color.RGBA{0x00, 0xFF, 0x7F, 255},
	"mediumseagreen":    color.RGBA{0x3C, 0xB3, 0x71, 255},
	"seagreen":          color.RGBA{0x2E, 0x8B, 0x57, 255},
	"forestgreen":       color.RGBA{0x22, 0x8B, 0x22, 255},
	"green":             color.RGBA{0x00, 0x80, 0x00, 255},
	"darkgreen":         color.RGBA{0x00, 0x64, 0x00, 255},
	"yellowgreen":       color.RGBA{0x9A, 0xCD, 0x32, 255},
	"olivedrab":         color.RGBA{0x6B, 0x8E, 0x23, 255},
	"olive":             color.RGBA{0x80, 0x80, 0x00, 255},
	"darkolivegreen":    color.RGBA{0x55, 0x6B, 0x2F, 255},
	"mediumaquamarine":  color.RGBA{0x66, 0xCD, 0xAA, 255},
	"darkseagreen":      color.RGBA{0x8F, 0xBC, 0x8B, 255},
	"lightseagreen":     color.RGBA{0x20, 0xB2, 0xAA, 255},
	"darkcyan":          color.RGBA{0x00, 0x8B, 0x8B, 255},
	"teal":              color.RGBA{0x00, 0x80, 0x80, 255},

	// Blue/Cyan colors
	"aqua":            color.RGBA{0x00, 0xFF, 0xFF, 255},
	"cyan":            color.RGBA{0x00, 0xFF, 0xFF, 255},
	"lightcyan":       color.RGBA{0xE0, 0xFF, 0xFF, 255},
	"paleturquoise":   color.RGBA{0xAF, 0xEE, 0xEE, 255},
	"aquamarine":      color.RGBA{0x7F, 0xFF, 0xD4, 255},
	"turquoise":       color.RGBA{0x40, 0xE0, 0xD0, 255},
	"mediumturquoise": color.RGBA{0x48, 0xD1, 0xCC, 255},
	"darkturquoise":   color.RGBA{0x00, 0xCE, 0xD1, 255},
	"cadetblue":       color.RGBA{0x5F, 0x9E, 0xA0, 255},
	"steelblue":       color.RGBA{0x46, 0x82, 0xB4, 255},
	"lightsteelblue":  color.RGBA{0xB0, 0xC4, 0xDE, 255},
	"powderblue":      color.RGBA{0xB0, 0xE0, 0xE6, 255},
	"lightblue":       color.RGBA{0xAD, 0xD8, 0xE6, 255},
	"skyblue":         color.RGBA{0x87, 0xCE, 0xEB, 255},
	"lightskyblue":    color.RGBA{0x87, 0xCE, 0xFA, 255},
	"deepskyblue":     color.RGBA{0x00, 0xBF, 0xFF, 255},
	"dodgerblue":      color.RGBA{0x1E, 0x90, 0xFF, 255},
	"cornflowerblue":  color.RGBA{0x64, 0x95, 0xED, 255},
	"royalblue":       color.RGBA{0x41, 0x69, 0xE1, 255},
	"blue":            color.RGBA{0x00, 0x00, 0xFF, 255},
	"mediumblue":      color.RGBA{0x00, 0x00, 0xCD, 255},
	"darkblue":        color.RGBA{0x00, 0x00, 0x8B, 255},
	"navy":            color.RGBA{0x00, 0x00, 0x80, 255},
	"midnightblue":    color.RGBA{0x19, 0x19, 0x70, 255},

	// Brown colors
	"cornsilk":       color.RGBA{0xFF, 0xF8, 0xDC, 255},
	"blanchedalmond": color.RGBA{0xFF, 0xEB, 0xCD, 255},
	"bisque":         color.RGBA{0xFF, 0xE4, 0xC4, 255},
	"navajowhite":    color.RGBA{0xFF, 0xDE, 0xAD, 255},
	"wheat":          color.RGBA{0xF5, 0xDE, 0xB3, 255},
	"burlywood":      color.RGBA{0xDE, 0xB8, 0x87, 255},
	"tan":            color.RGBA{0xD2, 0xB4, 0x8C, 255},
	"rosybrown":      color.RGBA{0xBC, 0x8F, 0x8F, 255},
	"sandybrown":     color.RGBA{0xF4, 0xA4, 0x60, 255},
	"goldenrod":      color.RGBA{0xDA, 0xA5, 0x20, 255},
	"darkgoldenrod":  color.RGBA{0xB8, 0x86, 0x0B, 255},
	"peru":           color.RGBA{0xCD, 0x85, 0x3F, 255},
	"chocolate":      color.RGBA{0xD2, 0x69, 0x1E, 255},
	"saddlebrown":    color.RGBA{0x8B, 0x45, 0x13, 255},
	"sienna":         color.RGBA{0xA0, 0x52, 0x2D, 255},
	"brown":          color.RGBA{0xA5, 0x2A, 0x2A, 255},
	"maroon":         color.RGBA{0x80, 0x00, 0x00, 255},

	// White colors
	"white":         color.RGBA{0xFF, 0xFF, 0xFF, 255},
	"snow":          color.RGBA{0xFF, 0xFA, 0xFA, 255},
	"honeydew":      color.RGBA{0xF0, 0xFF, 0xF0, 255},
	"mintcream":     color.RGBA{0xF5, 0xFF, 0xFA, 255},
	"azure":         color.RGBA{0xF0, 0xFF, 0xFF, 255},
	"aliceblue":     color.RGBA{0xF0, 0xF8, 0xFF, 255},
	"ghostwhite":    color.RGBA{0xF8, 0xF8, 0xFF, 255},
	"whitesmoke":    color.RGBA{0xF5, 0xF5, 0xF5, 255},
	"seashell":      color.RGBA{0xFF, 0xF5, 0xEE, 255},
	"beige":         color.RGBA{0xF5, 0xF5, 0xDC, 255},
	"oldlace":       color.RGBA{0xFD, 0xF5, 0xE6, 255},
	"floralwhite":   color.RGBA{0xFF, 0xFA, 0xF0, 255},
	"ivory":         color.RGBA{0xFF, 0xFF, 0xF0, 255},
	"antiquewhite":  color.RGBA{0xFA, 0xEB, 0xD7, 255},
	"linen":         color.RGBA{0xFA, 0xF0, 0xE6, 255},
	"lavenderblush": color.RGBA{0xFF, 0xF0, 0xF5, 255},
	"mistyrose":     color.RGBA{0xFF, 0xE4, 0xE1, 255},

	// Gray colors
	"gainsboro":      color.RGBA{0xDC, 0xDC, 0xDC, 255},
	"lightgray":      color.RGBA{0xD3, 0xD3, 0xD3, 255},
	"lightgrey":      color.RGBA{0xD3, 0xD3, 0xD3, 255}, // British spelling
	"silver":         color.RGBA{0xC0, 0xC0, 0xC0, 255},
	"darkgray":       color.RGBA{0xA9, 0xA9, 0xA9, 255},
	"darkgrey":       color.RGBA{0xA9, 0xA9, 0xA9, 255}, // British spelling
	"gray":           color.RGBA{0x80, 0x80, 0x80, 255},
	"grey":           color.RGBA{0x80, 0x80, 0x80, 255}, // British spelling
	"dimgray":        color.RGBA{0x69, 0x69, 0x69, 255},
	"dimgrey":        color.RGBA{0x69, 0x69, 0x69, 255}, // British spelling
	"lightslategray": color.RGBA{0x77, 0x88, 0x99, 255},
	"lightslategrey": color.RGBA{0x77, 0x88, 0x99, 255}, // British spelling
	"slategray":      color.RGBA{0x70, 0x80, 0x90, 255},
	"slategrey":      color.RGBA{0x70, 0x80, 0x90, 255}, // British spelling
	"darkslategray":  color.RGBA{0x2F, 0x4F, 0x4F, 255},
	"darkslategrey":  color.RGBA{0x2F, 0x4F, 0x4F, 255}, // British spelling
	"black":          color.RGBA{0x00, 0x00, 0x00, 255},
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
