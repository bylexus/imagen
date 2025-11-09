package main

import (
	"fmt"
	"os"

	"github.com/bylexus/imagen/pkg/cli"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "generate":
		cmd := &cli.GenerateCommand{}
		err = cmd.Execute(args)
	case "serve":
		cmd := &cli.ServeCommand{}
		err = cmd.Execute(args)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`imagen - A small image creation utility

Usage:
  imagen generate [options]  Generate static placeholder images
  imagen serve [options]     Start web server to serve placeholder images
  imagen help                Show this help message

Generate Options:
  --size, -s WxH            Image size (width x height), can be repeated
  --color-mode, -m MODE     Color mode: solid, tiled, gradient, noise (can be repeated)
  --color, -c COLOR         Color value (name, hex, or 'random'), can be repeated
  --border-width, -b WIDTH  Border width in pixels
  --border-color COLOR      Border color
  --gradient-angle, -a DEG  Gradient angle in degrees (0=top-down, 180=bottom-up)
  --tile-size, -ts SIZE     Tile size in pixels for tiled/noise mode
  --text, -t TEXT           Text to display (use {w} and {h} for placeholders)
  --text-size SIZE          Text size in pt
  --text-color COLOR        Text color
  --filename, -f NAME       Output filename (use {w}, {h}, {nr} for placeholders)
  --format FORMAT           Output format: png, jpeg

Serve Options:
  --listen ADDR             Listen address(es), comma-separated (default: :3000)
                            Examples: ":3000", "192.168.1.20:5555", "[::1]:4567"

URL Format (for serve mode):
  http://[host]/[size]/c:[color]/t:[text]/f:[format]/b:[border]

  Example:
    http://localhost:3000/400x300/c:blue/t:"hello, world",s:26,c:yellow/f:png/b:5,ffffff

Examples:
  # Generate a 512x384 image with blue background
  imagen generate --size 512x384 --color blue

  # Generate multiple images with different sizes and gradients
  imagen generate -s 400x300 -s 800x600 -m gradient -c red -c yellow

  # Start server on port 8080
  imagen serve --listen :8080

  # Start server on multiple addresses
  imagen serve --listen ":3000,:8080"
`)
}
