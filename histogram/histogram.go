package histogram

import (
	"image"
	"image/color"
)

type HistogramProfile int

const (
	HORIZONTAL HistogramProfile = iota
	VERTICAL
)

func Monochrome(img image.Image, axis HistogramProfile) []uint16 {
	bounds := img.Bounds()

	minRow := bounds.Min.Y
	maxRow := bounds.Max.Y
	minColumn := bounds.Min.X
	maxColumn := bounds.Max.X

	if axis == VERTICAL {
		minRow, minColumn = minColumn, minRow
		maxRow, maxColumn = maxColumn, maxRow
	}

	histogram := make([]uint16, maxRow)
	for row := minRow; row < maxRow; row++ {
		for column := minColumn; column < maxColumn; column++ {
			c := column
			r := row

			if axis == VERTICAL {
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
