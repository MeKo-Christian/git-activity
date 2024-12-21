package internal

import (
	"image/color"
	"math"
)

var colorPalette []color.Color

// GenerateColorPalette generates a palette of n distinct colors.
func GenerateColorPalette(n int) []color.Color {
	if n <= 0 {
		return nil
	}

	colors := make([]color.Color, n)

	// Degree step to ensure non-gradual jumps (e.g., 180 degrees in hue space)
	step := 1 / float64(n)
	for i := 0; i < n; i++ {
		hue := math.Mod(float64(i)*step, 1.0)
		colors[i] = hsvToRGBA(hue, 0.7, 0.9) // Saturation and value are adjustable
	}

	return colors
}

// hsvToRGBA converts HSV color values to RGBA.
func hsvToRGBA(h, s, v float64) color.Color {
	var r, g, b float64
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h*6, 2)-1))
	m := v - c

	switch {
	case h < 1.0/6:
		r, g, b = c, x, 0
	case h < 2.0/6:
		r, g, b = x, c, 0
	case h < 3.0/6:
		r, g, b = 0, c, x
	case h < 4.0/6:
		r, g, b = 0, x, c
	case h < 5.0/6:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

func init() {
	colorPalette = GenerateColorPalette(16)
}
