package threshold

import (
	"errors"
	"image"
	"math"
)

// Default Sauvola parameters.
const (
	// DefaultSauvolaWindow is the default side length (in pixels) of the square local window.
	DefaultSauvolaWindow = 15

	// DefaultSauvolaK is the default sensitivity. 
	// Sauvola and Pietikäinen suggest 0.5; 0.2–0.34 is common for document images.
	DefaultSauvolaK = 0.34

	// DefaultSauvolaR is the default dynamic range of the standard deviation for an 8-bit image.
	DefaultSauvolaR = 128.0
)

// ErrInvalidWindow is returned when the window size is not a positive odd number.
var ErrInvalidWindow = errors.New("threshold: window size must be a positive odd number")

// SauvolaOptions configures Sauvola. Zero values select the defaults.
type SauvolaOptions struct {
	// Window is the side length of the square neighborhood. It must be odd.
	// Zero means DefaultSauvolaWindow.
	Window int

	// K controls how strongly the local standard deviation lowers the threshold below the local mean. 
	// Zero means DefaultSauvolaK.
	K float64

	// R is the dynamic range of the standard deviation. 
	// Zero means DefaultSauvolaR.
	R float64
}

// Sauvola computes a per-pixel adaptive threshold map using Sauvola's method:
//
//	T(x, y) = m(x, y) * (1 + K * (s(x, y)/R - 1))
//
// where m and s are the mean and standard deviation of the gray values in a
// Window×Window neighborhood centered on (x, y). The neighborhood is clipped
// at the image borders, so edge pixels use a smaller window.
//
// The result is indexed [y][x] relative to the image's top-left corner
// (img.Bounds().Min), so it is correct for sub-images too. A pixel is
// considered foreground (white) when its value is greater than its threshold,
// matching the convention of Otsu and Image.Monochrome.
//
// Runtime is O(width*height) regardless of the window size, using integral
// images.
func Sauvola(img *image.Gray, opt SauvolaOptions) ([][]uint8, error) {
	if img == nil || img.Bounds().Empty() {
		return nil, ErrEmptyImage
	}
	if opt.Window == 0 {
		opt.Window = DefaultSauvolaWindow
	}
	if opt.K == 0 {
		opt.K = DefaultSauvolaK
	}
	if opt.R == 0 {
		opt.R = DefaultSauvolaR
	}
	if opt.Window < 1 || opt.Window%2 == 0 {
		return nil, ErrInvalidWindow
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	// Integral images with one row/column of zero padding:
	// sum[y+1][x+1] holds the sum over rows 0..y and columns 0..x.
	stride := w + 1
	sum := make([]int64, stride*(h+1))
	sqSum := make([]int64, stride*(h+1))
	for y := 0; y < h; y++ {
		off := img.PixOffset(b.Min.X, b.Min.Y+y)
		row := img.Pix[off : off+w]
		var rowSum, rowSq int64
		for x, v := range row {
			p := int64(v)
			rowSum += p
			rowSq += p * p
			i := (y+1)*stride + x + 1
			sum[i] = sum[i-stride] + rowSum
			sqSum[i] = sqSum[i-stride] + rowSq
		}
	}

	half := opt.Window / 2
	out := make([][]uint8, h)
	buf := make([]uint8, w*h)
	for y := 0; y < h; y++ {
		out[y] = buf[y*w : (y+1)*w : (y+1)*w]

		y0, y1 := max(y-half, 0), min(y+half+1, h)
		for x := 0; x < w; x++ {
			x0, x1 := max(x-half, 0), min(x+half+1, w)
			n := float64((x1 - x0) * (y1 - y0))

			s := windowSum(sum, stride, x0, y0, x1, y1)
			sq := windowSum(sqSum, stride, x0, y0, x1, y1)

			mean := float64(s) / n
			variance := float64(sq)/n - mean*mean
			if variance < 0 {
				variance = 0 // guard against rounding error
			}
			std := math.Sqrt(variance)

			t := mean * (1 + opt.K*(std/opt.R-1))
			out[y][x] = clampUint8(t)
		}
	}

	return out, nil
}

// windowSum returns the sum over [x0,x1)×[y0,y1) from an integral image.
func windowSum(integral []int64, stride, x0, y0, x1, y1 int) int64 {
	return integral[y1*stride+x1] - integral[y0*stride+x1] -
		integral[y1*stride+x0] + integral[y0*stride+x0]
}

func clampUint8(v float64) uint8 {
	switch {
	case v <= 0:
		return 0
	case v >= 255:
		return 255
	}
	return uint8(v + 0.5)
}