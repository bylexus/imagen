# imagen

imagen is a small image creation utility to create placeholder images. It can either be started as a command line utility to generate static images, or as a web server to serve placeholder images on request.

## Supported image configurations

by default, imagen just creates a sample 256x192 gray image with a white text on it (drawing the actual size). But imagen can produce indivudualized images. The following parameters are supported:

- image size (width, height)
- background colors:
  - solid color (e.g. 'blue', or '#0000FF')
  - random solid color
  - random solid colors from a given set of colors
  - pixelated / tiled with multiple colors (e.g. black/white tiles)
  - color gradients with 2 or multiple colors and angles
- configurable text
  - text content
  - text color (default: white xor'ed with the background)
  - text size
  - text angle
- border: The image can also have a border:
  - border width
  - border color
- image output format: png, jpeg, webp

## Starting / using imagen

```
# cli mode:
imagen generate [parameters]

# web server mode:
imagen serve [parameters]

```

### generate parameters

- `--size=[WxH]`, `-s [WxH]: width/height. defaults to 256x192. Can be given multiple times to generate multiple images (see `--filename` below)
- `--color-mode=[mode]`, `-m [mode]`: The color mode. Valid modes are: `solid`, `tiled`, `gradient`, `noise`. The difference of "tiled" and "noise" is the randomness: "tiled" mode generates tile colors in order, while "noise" produces random tiles.
- `--color=[color]`, `-c [color]`: single color when using "solid". Can be a name (e.g. `blue`), an RGB code (e.g. `A0F34B`), `random` for a complete random color, or a comma-separated list of the values before: This list is used when using the `--nr` parameter to generate multiple images: It then randomly selects one entry from the list per image. This parameter can be repeated for defining multiple colors for the other modes.
- `--border-width=[nr]` `-b [width]`: The border with in pixels
- `--border-color=[color]`: The border color
- `--gradient-angle=angle`, `-a angle`: The gradient angle in degrees (0 = top-down, 180=bottom up)
- `--tile-size=width`, `-s width`: The size in pixel of a single tile in tile/noise mode
- `--text=[text]`, `-t text`: The text to output. You can use `{w}` and `{h}` as placeholder for the generated size.
- `--text-size=[size]`: The text size in pt
- `--text-color=[color]`: The text color (e.g. `white` or `ffffff`)
- `--nr=[nr]`, `-n nr`: Number of images to create. The filename is enhanced with a large enough counting number (e.g. `image-0001.png`)
- `--filename=[filename]`, `-f filename`: Output filename. You can use `{w}`, `{h}`, `{nr}` in the filename as placeholders for width, height, and image number

### serve parameters

the `serve` command starts a web server on port 3000 by default. You can configure its behaviour with the following parameters:

`--listen=[listen address]`: tcp ip/port to listen, e.g. `:3000` to list on all IPs on port 3000, or `192.168.1.20:5555` for a specific IPv4, `[::1]:4567` for an IPv6. Multiple listener addresses can be separated by comma.

## URL format

All the above options can be defined as URL parameters. The standard image can just be produced with

```
# produces the standard image:
http://[imagen-url]/
```

### url scheme

All parameters can be defined in the URL. The first parameter is always the size of the image,
while the following parameters can be placed in any order: They are prefixed with a single
character to indicate the parameter type:

```
http://[imagen-url]/[size]/c:[color-config]/t:[text]/f:[format]/b:[border]
```

**Example:**

http://[imagen-url]/400x300/c:blue/t:"hello, world",s:26,c:yellow/f:png/b:5,ffffff

This will create a 400x300px solid blue image with a yellow "hello world" text in size 26pt as PNG image. It has a 5 pixel white border.


## Software Architecture

The program consists of three main modules:

- the image generator module includes all logic to generate images
- the web server module manages the web server and parses the parameters from the url. It uses the generator module to create the images.
- the cli module offers the cli interface and parses the parameters from the command line. It uses the generator module to create the images.

