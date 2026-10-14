package threshold

import (
	"errors"
	"image"
)

// Levels is the number of gray levels in an 8-bit image.
const Levels = 256

// ErrEmptyImage is returned when the image (or histogram) contains no pixels.
var ErrEmptyImage = errors.New("threshold: image has no pixels")

// Otsu returns the global threshold chosen by Otsu's method for img.
func Otsu(img *image.Gray) (uint8, error) {
	return OtsuHistogram(Histogram(img))
}

// Histogram returns the gray-level histogram of img: Histogram(img)[v] is the
// number of pixels with value v. Sub-images with a non-zero origin are handled
// correctly. A nil image yields an all-zero histogram.
func Histogram(img *image.Gray) [Levels]int {
	var hist [Levels]int
	if img == nil {
		return hist
	}

	b := img.Bounds()
	w := b.Dx()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		off := img.PixOffset(b.Min.X, y)
		for _, v := range img.Pix[off : off+w] {
			hist[v]++
		}
	}
	return hist
}

func OtsuHistogram(hist [Levels]int) (uint8, error) {
	var total, sumAll int64
	for level, count := range hist {
		total += int64(count)
		sumAll += int64(level) * int64(count)
	}
	if total == 0 {
		return 0, ErrEmptyImage
	}

	// Background pixel count and gray-level sum for levels <= t.
	var w0, sum0 int64

	// Highest between-class variance seen so far, and the first and last t of
	// the contiguous run of thresholds that reaches it.
	best := -1.0
	var first, last int

	// t = Levels-1 would leave the foreground empty, so it is never a candidate.
	for t := 0; t < Levels-1; t++ {
		count := int64(hist[t])
		w0 += count
		sum0 += int64(t) * count

		w1 := total - w0
		if w0 == 0 || w1 == 0 {
			continue // one class is empty: not a valid split
		}

		mean0 := float64(sum0) / float64(w0)
		mean1 := float64(sumAll-sum0) / float64(w1)
		d := mean0 - mean1
		variance := float64(w0) * float64(w1) * d * d

		switch {
		case variance > best:
			best, first, last = variance, t, t
		case variance == best && t == last+1:
			// Empty bins leave w0 and sum0 unchanged, so the value repeats exactly.
			last = t
		}
	}

	if best < 0 {
		// Single gray level: no valid split exists.
		for level, count := range hist {
			if count > 0 {
				return uint8(level), nil
			}
		}
	}

	return uint8((first + last) / 2), nil
}