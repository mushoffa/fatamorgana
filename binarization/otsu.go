package binarization

import (
	"image"
	"image/color"
	"runtime"
	"sync"
	"time"
)

// Otsu converts a grayscale image to a binary image using a global threshold.
// The threshold is expected to be computed with Otsu's method by the caller.
// Rows are processed concurrently by multiple goroutines.
// It returns the binary image and the time spent on the conversion.
// If img is nil or empty, it returns nil and a zero duration.
func Otsu(img *image.Gray, threshold uint8) (*image.Gray, time.Duration) {
	if img == nil {
		return nil, 0
	}

	bounds := img.Bounds()
	if bounds.Empty() {
		return nil, 0
	}

	start := time.Now()
	binary := image.NewGray(bounds)
	height := bounds.Dy()
	workers := min(runtime.NumCPU(), height)
	rowsPerWorker := (height + workers - 1) / workers

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		yStart := bounds.Min.Y + w*rowsPerWorker
		yEnd := min(yStart+rowsPerWorker, bounds.Max.Y)

		wg.Add(1)
		go func(yStart, yEnd int) {
			defer wg.Done()

			// Each goroutine writes only to its own rows, so no locking is needed.
			for y := yStart; y < yEnd; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					pixel := img.GrayAt(x, y).Y
					// Black pixel color
					c := color.Gray{Y: 0}
					if pixel > threshold {
						// White pixel color
						c = color.Gray{Y: 255}
					}
					binary.SetGray(x, y, c)
				}
			}
		}(yStart, yEnd)
	}
	wg.Wait()

	return binary, time.Since(start)
}
