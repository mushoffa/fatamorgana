package histogram

import (
	"image"
	"image/color"
)

type HistogramProfile int

const (
	Horizontal HistogramProfile = iota
	Vertical
)

// Binary returns the projection profile of a binary image: the number of black
// (gray value 0) pixels per row (Horizontal) or per column (Vertical).
// Any non-zero value counts as white.
func Binary(img *image.Gray, axis HistogramProfile) []int {
	if img == nil {
		return []int{}
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	size := h
	if axis == Vertical {
		size = w
	}
	histogram := make([]int, size)
	if w <= 0 || h <= 0 {
		return histogram
	}

	for y := 0; y < h; y++ {
		off := img.PixOffset(b.Min.X, b.Min.Y+y)
		for x, v := range img.Pix[off : off+w] {
			if v != 0 {
				continue
			}
			if axis == Vertical {
				histogram[x]++
			} else {
				histogram[y]++
			}
		}
	}

	return histogram
}

// Will be deprecated, monochrome is two tone, not only black & white
// Function name is ambiguous
func Monochrome(img image.Image, axis HistogramProfile) []uint16 {
	bounds := img.Bounds()

	minRow := bounds.Min.Y
	maxRow := bounds.Max.Y
	minColumn := bounds.Min.X
	maxColumn := bounds.Max.X

	if axis == Vertical {
		minRow, minColumn = minColumn, minRow
		maxRow, maxColumn = maxColumn, maxRow
	}

	histogram := make([]uint16, maxRow)
	for row := minRow; row < maxRow; row++ {
		for column := minColumn; column < maxColumn; column++ {
			c := column
			r := row

			if axis == Vertical {
				c, r = r, c
			}

			pixel := img.At(c, r).(color.Gray).Y
			// 8-bit grayscale color
			// 0 = black
			// 255 = white
			if pixel == 0 {
				histogram[row]++
			}
		}
	}

	return histogram
}
