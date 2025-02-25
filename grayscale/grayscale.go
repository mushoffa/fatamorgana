package grayscale

import (
	"image/color"
)

// MethodType represents a grayscale algorithm method type.
type MethodType color.Model

// Common grayscale algorithm methods.
var (
	// AVERAGING represents the averaging conversion method.
	AVERAGING MethodType = color.ModelFunc(averaging)

	// LUMINOSITY_601 represents the luminosity conversion method using ITU-R 601 coefficient format.
	LUMINOSITY_601 MethodType = color.ModelFunc(luminosity_601)

	// LUMINOSITY_709 represents the luminosity conversion method ITU-R 709 coefficient format.
	LUMINOSITY_709 MethodType = color.ModelFunc(luminosity_709)

	// LUMINOSITY_2100 represents the luminosity conversion method ITU-R 1200 coefficient format.
	LUMINOSITY_2100 MethodType = color.ModelFunc(luminosity_2100)

	// SINGLE_CHANNEL represents the single channel conversion method.
	SINGLE_CHANNEL MethodType = color.ModelFunc(single_channel)

	// WEIGHTED represents the weighted conversion method.
	WEIGHTED MethodType = color.ModelFunc(weighted)
)

// MethodTypes stores as pairs of common grayscale algorithm method types and its alias names.
var MethodTypes = map[MethodType]string{
	AVERAGING:       "AVERAGING",
	LUMINOSITY_601:  "LUMINOSITY_601",
	LUMINOSITY_709:  "LUMINOSITY_709",
	LUMINOSITY_2100: "LUMINOSITY_2100",
	SINGLE_CHANNEL:  "SINGLE_CHANNEL",
	WEIGHTED:        "WEIGHTED",
}

// single_channel returns a copy of the image in Grayscale using averaging method.
func averaging(c color.Color) color.Color {
	if _, ok := c.(color.Gray); ok {
		return c
	}

	r, g, b := rgb(c)
	y := (r + g + b) / 3

	return color.Gray{uint8(y)}
}

// single_channel returns a copy of the image in Grayscale using luminosity method.
func luminosity_601(c color.Color) color.Color {
	if isGray(c) {
		return c
	}

	return color.GrayModel.Convert(c)
}

// single_channel returns a copy of the image in Grayscale using luminosity method.
func luminosity_709(c color.Color) color.Color {
	if isGray(c) {
		return c
	}

	r, g, b := rgb(c)
	y := 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b)

	return color.Gray{uint8(y)}
}

// single_channel returns a copy of the image in Grayscale using luminosity method.
func luminosity_2100(c color.Color) color.Color {
	if isGray(c) {
		return c
	}

	r, g, b := rgb(c)
	y := 0.2672*float32(r) + 0.6780*float32(g) + 0.0593*float32(b)

	return color.Gray{uint8(y)}
}

// single_channel returns a copy of the image in Grayscale using singel channel method.
func single_channel(c color.Color) color.Color {
	if isGray(c) {
		return c
	}

	r, _, _ := rgb(c)

	return color.Gray{uint8(r)}
}

// single_channel returns a copy of the image in Grayscale using weighted method.
func weighted(c color.Color) color.Color {
	if isGray(c) {
		return c
	}

	r, g, b := rgb(c)
	y := 0.3*float32(r) + 0.59*float32(g) + 0.11*float32(b)

	return color.Gray{uint8(y)}
}

// isGray checks and returns true if the given image is gray.
func isGray(c color.Color) bool {
	if _, ok := c.(color.Gray); ok {
		return true
	}

	return false
}

func rgb(c color.Color) (r, g, b uint32) {
	r, g, b, _ = c.RGBA()
	r = r >> 8
	g = g >> 8
	b = b >> 8

	return
}
