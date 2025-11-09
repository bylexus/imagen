package generator

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Generator handles image generation
type Generator struct {
	config *ImageConfig
}

// NewGenerator creates a new image generator with the given configuration
func NewGenerator(config *ImageConfig) *Generator {
	return &Generator{config: config}
}

// Generate creates the image based on the configuration
func (g *Generator) Generate() (image.Image, error) {
	// Create the base image
	img := image.NewRGBA(image.Rect(0, 0, g.config.Width, g.config.Height))

	// Draw background
	if err := g.drawBackground(img); err != nil {
		return nil, err
	}

	// Draw border
	if g.config.BorderWidth > 0 {
		g.drawBorder(img)
	}

	// Draw text
	if g.config.Text != "" {
		g.drawText(img)
	}

	return img, nil
}

// WriteImage writes the image to the given writer in the specified format
func (g *Generator) WriteImage(w io.Writer, img image.Image) error {
	switch strings.ToLower(g.config.Format) {
	case "png":
		return png.Encode(w, img)
	case "jpeg", "jpg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
	default:
		return fmt.Errorf("unsupported format: %s", g.config.Format)
	}
}

// drawBackground draws the background based on the color mode
func (g *Generator) drawBackground(img *image.RGBA) error {
	switch g.config.ColorMode {
	case ColorModeSolid:
		g.drawSolidBackground(img)
	case ColorModeTiled:
		g.drawTiledBackground(img, false)
	case ColorModeNoise:
		g.drawTiledBackground(img, true)
	case ColorModeGradient:
		g.drawGradientBackground(img)
	default:
		return fmt.Errorf("unsupported color mode: %s", g.config.ColorMode)
	}
	return nil
}

// drawSolidBackground fills the image with a solid color
func (g *Generator) drawSolidBackground(img *image.RGBA) {
	c := g.config.Colors[0]
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
}

// drawTiledBackground draws a tiled/pixelated background
func (g *Generator) drawTiledBackground(img *image.RGBA, random bool) {
	tileSize := g.config.TileSize
	if tileSize <= 0 {
		tileSize = 16
	}

	colors := g.config.Colors
	if len(colors) == 0 {
		colors = []color.Color{color.Black, color.White}
	}

	colorIndex := 0
	for y := 0; y < g.config.Height; y += tileSize {
		for x := 0; x < g.config.Width; x += tileSize {
			// Select color
			var c color.Color
			if random {
				c = colors[rand.Intn(len(colors))]
			} else {
				c = colors[colorIndex%len(colors)]
				colorIndex++
			}

			// Draw tile
			rect := image.Rect(x, y, min(x+tileSize, g.config.Width), min(y+tileSize, g.config.Height))
			draw.Draw(img, rect, &image.Uniform{c}, image.Point{}, draw.Src)
		}
	}
}

// drawGradientBackground draws a gradient background
func (g *Generator) drawGradientBackground(img *image.RGBA) {
	colors := g.config.Colors
	if len(colors) < 2 {
		// Fall back to solid color
		g.drawSolidBackground(img)
		return
	}

	angle := g.config.GradientAngle
	width := float64(g.config.Width)
	height := float64(g.config.Height)

	// Convert angle to radians
	angleRad := angle * math.Pi / 180.0

	// Calculate gradient direction vector
	// 0 degrees = top to bottom (dy positive)
	// 90 degrees = left to right (dx positive)
	// 180 degrees = bottom to top (dy negative)
	dx := math.Sin(angleRad)
	dy := math.Cos(angleRad)

	// Calculate the maximum distance along the gradient direction
	corners := []struct{ x, y float64 }{
		{0, 0},
		{width, 0},
		{0, height},
		{width, height},
	}

	var minProj, maxProj float64
	for i, corner := range corners {
		proj := corner.x*dx + corner.y*dy
		if i == 0 {
			minProj = proj
			maxProj = proj
		} else {
			minProj = math.Min(minProj, proj)
			maxProj = math.Max(maxProj, proj)
		}
	}

	gradientLength := maxProj - minProj

	// Draw the gradient
	for y := 0; y < g.config.Height; y++ {
		for x := 0; x < g.config.Width; x++ {
			// Calculate position along gradient (0.0 to 1.0)
			proj := float64(x)*dx + float64(y)*dy
			t := (proj - minProj) / gradientLength

			// Handle multiple colors
			var c color.Color
			if len(colors) == 2 {
				c = InterpolateColor(colors[0], colors[1], t)
			} else {
				// Multi-color gradient
				segment := t * float64(len(colors)-1)
				idx := int(segment)
				if idx >= len(colors)-1 {
					c = colors[len(colors)-1]
				} else {
					localT := segment - float64(idx)
					c = InterpolateColor(colors[idx], colors[idx+1], localT)
				}
			}

			img.Set(x, y, c)
		}
	}
}

// drawBorder draws a border around the image
func (g *Generator) drawBorder(img *image.RGBA) {
	bounds := img.Bounds()
	borderColor := g.config.BorderColor
	width := g.config.BorderWidth

	// Top and bottom borders
	for i := 0; i < width; i++ {
		draw.Draw(img, image.Rect(0, i, bounds.Max.X, i+1), &image.Uniform{borderColor}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(0, bounds.Max.Y-i-1, bounds.Max.X, bounds.Max.Y-i), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	}

	// Left and right borders
	for i := 0; i < width; i++ {
		draw.Draw(img, image.Rect(i, 0, i+1, bounds.Max.Y), &image.Uniform{borderColor}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(bounds.Max.X-i-1, 0, bounds.Max.X-i, bounds.Max.Y), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	}
}

// drawText draws text on the image
func (g *Generator) drawText(img *image.RGBA) {
	// Replace placeholders in text
	text := g.config.Text
	text = strings.ReplaceAll(text, "{w}", fmt.Sprintf("%d", g.config.Width))
	text = strings.ReplaceAll(text, "{h}", fmt.Sprintf("%d", g.config.Height))

	// Determine text color
	var textColor color.Color = color.RGBA{255, 255, 255, 255}
	if g.config.TextColor != nil {
		textColor = *g.config.TextColor
	}

	// Calculate inverted color for text border
	borderColor := invertColor(textColor)

	// Try to load TrueType font, fall back to basicfont
	face := g.loadFont()

	// Measure text
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	bounds, _ := drawer.BoundString(text)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

	// Center the text
	x := (g.config.Width - textWidth) / 2
	y := (g.config.Height + textHeight) / 2

	basePoint := fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	// Draw text border (outline) by drawing the text in 8 directions with border color
	borderOffsets := []struct{ dx, dy int }{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}

	borderDrawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(borderColor),
		Face: face,
	}

	for _, offset := range borderOffsets {
		borderDrawer.Dot = fixed.Point26_6{
			X: basePoint.X + fixed.I(offset.dx),
			Y: basePoint.Y + fixed.I(offset.dy),
		}
		borderDrawer.DrawString(text)
	}

	// Draw the main text on top
	drawer.Dot = basePoint
	drawer.DrawString(text)
}

// invertColor returns the inverted (complementary) color
func invertColor(c color.Color) color.Color {
	r, g, b, a := c.RGBA()
	// Convert from 16-bit to 8-bit
	r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)
	// Invert RGB components
	return color.RGBA{
		R: 255 - r8,
		G: 255 - g8,
		B: 255 - b8,
		A: a8,
	}
}

// loadFont attempts to load a TrueType font from the system, falls back to basicfont
func (g *Generator) loadFont() font.Face {
	// Try to load system TrueType font
	fontPaths := getSystemFontPaths()

	for _, fontPath := range fontPaths {
		if face := tryLoadTTF(fontPath, g.config.TextSize); face != nil {
			return face
		}
	}

	// Fall back to basicfont
	return basicfont.Face7x13
}

// getSystemFontPaths returns common system font paths based on OS
func getSystemFontPaths() []string {
	switch runtime.GOOS {
	case "darwin": // macOS
		return []string{
			"/System/Library/Fonts/Helvetica.ttc",
			"/System/Library/Fonts/SFNSText.ttf",
			"/System/Library/Fonts/SFNS.ttf",
			"/Library/Fonts/Arial.ttf",
			"/System/Library/Fonts/Supplemental/Arial.ttf",
		}
	case "linux":
		return []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
			"/usr/share/fonts/TTF/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/freefont/FreeSans.ttf",
		}
	case "windows":
		return []string{
			"C:\\Windows\\Fonts\\arial.ttf",
			"C:\\Windows\\Fonts\\calibri.ttf",
		}
	default:
		return []string{}
	}
}

// tryLoadTTF attempts to load a TrueType font from the given path
func tryLoadTTF(fontPath string, size float64) font.Face {
	// Check if file exists
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		return nil
	}

	// Read font file
	fontData, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return nil
	}

	// Parse font
	// Try as TrueType collection first (for .ttc files)
	f, err := opentype.ParseCollection(fontData)
	if err == nil && f.NumFonts() > 0 {
		// Use first font in collection
		font, err := f.Font(0)
		if err == nil {
			face, err := opentype.NewFace(font, &opentype.FaceOptions{
				Size: size,
				DPI:  72,
			})
			if err == nil {
				return face
			}
		}
	}

	// Try as single TrueType font
	ttf, err := opentype.Parse(fontData)
	if err != nil {
		return nil
	}

	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})
	if err != nil {
		return nil
	}

	return face
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
