package threshold

import (
	"errors"
	"image"
	"math"
	"math/rand"
	"testing"
)

// referenceSauvola recomputes every window from scratch.
func referenceSauvola(img *image.Gray, window int, k, r float64) [][]uint8 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	half := window / 2
	out := make([][]uint8, h)
	for y := 0; y < h; y++ {
		out[y] = make([]uint8, w)
		for x := 0; x < w; x++ {
			var sum, sq, n float64
			for yy := max(y-half, 0); yy <= min(y+half, h-1); yy++ {
				for xx := max(x-half, 0); xx <= min(x+half, w-1); xx++ {
					v := float64(img.GrayAt(b.Min.X+xx, b.Min.Y+yy).Y)
					sum += v
					sq += v * v
					n++
				}
			}
			mean := sum / n
			std := math.Sqrt(math.Max(sq/n-mean*mean, 0))
			out[y][x] = clampUint8(mean * (1 + k*(std/r-1)))
		}
	}
	return out
}

func randomGray(rect image.Rectangle, seed int64) *image.Gray {
	rng := rand.New(rand.NewSource(seed))
	img := image.NewGray(rect)
	for i := range img.Pix {
		img.Pix[i] = uint8(rng.Intn(Levels))
	}
	return img
}

func TestSauvolaErrors(t *testing.T) {
	// Given a nil image
	// When Sauvola is called
	// Then it returns ErrEmptyImage
	if _, err := Sauvola(nil, SauvolaOptions{}); !errors.Is(err, ErrEmptyImage) {
		t.Errorf("nil image err = %v, want ErrEmptyImage", err)
	}

	// Given an image with no pixels
	// When Sauvola is called
	// Then it returns ErrEmptyImage
	if _, err := Sauvola(image.NewGray(image.Rect(0, 0, 0, 0)), SauvolaOptions{}); !errors.Is(err, ErrEmptyImage) {
		t.Errorf("empty image err = %v, want ErrEmptyImage", err)
	}

	// Given a valid image and a window that is negative or even
	// When Sauvola is called
	// Then it returns ErrInvalidWindow
	for _, win := range []int{-3, -1, 2, 10} {
		if _, err := Sauvola(filled(4, 4, 10), SauvolaOptions{Window: win}); !errors.Is(err, ErrInvalidWindow) {
			t.Errorf("window %d err = %v, want ErrInvalidWindow", win, err)
		}
	}
}

func TestSauvolaShape(t *testing.T) {
	// Given a 7x5 image
	img := randomGray(image.Rect(0, 0, 7, 5), 1)

	// When Sauvola is called with default options
	got, err := Sauvola(img, SauvolaOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Then the threshold map has 5 rows of 7 columns, indexed [y][x]
	if len(got) != 5 {
		t.Fatalf("rows = %d, want 5", len(got))
	}
	for y, row := range got {
		if len(row) != 7 {
			t.Errorf("row %d has %d columns, want 7", y, len(row))
		}
	}
}

func TestSauvolaUniformImage(t *testing.T) {
	// Given a uniform image of value 200, so the local std is 0 everywhere
	img := filled(9, 9, 200)

	// When Sauvola is called with K = 0.5
	got, err := Sauvola(img, SauvolaOptions{Window: 5, K: 0.5})
	if err != nil {
		t.Fatal(err)
	}

	// Then every threshold is m*(1-K) = 100
	for y, row := range got {
		for x, v := range row {
			if v != 100 {
				t.Fatalf("T(%d,%d) = %d, want 100", x, y, v)
			}
		}
	}
}

func TestSauvolaDefaults(t *testing.T) {
	// Given a random image
	img := randomGray(image.Rect(0, 0, 20, 20), 2)

	// When Sauvola is called with zero-value options and with the default
	// values spelled out explicitly
	a, err := Sauvola(img, SauvolaOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Sauvola(img, SauvolaOptions{Window: DefaultSauvolaWindow, K: DefaultSauvolaK, R: DefaultSauvolaR})
	if err != nil {
		t.Fatal(err)
	}

	// Then both produce identical threshold maps
	for y := range a {
		for x := range a[y] {
			if a[y][x] != b[y][x] {
				t.Fatalf("defaults differ at (%d,%d): %d vs %d", x, y, a[y][x], b[y][x])
			}
		}
	}
}

func TestSauvolaMatchesReference(t *testing.T) {
	cases := []struct {
		w, h, win int
		k, r      float64
	}{
		{1, 1, 3, 0.34, 128},
		{8, 6, 1, 0.34, 128},
		{8, 6, 3, 0.2, 128},
		{17, 11, 5, 0.5, 128},
		{30, 25, 15, 0.34, 128},
		{10, 10, 31, 0.34, 100}, // window larger than the image
		{12, 40, 7, 0.1, 255},
	}
	for i, tc := range cases {
		// Given a random image and a set of Sauvola parameters
		img := randomGray(image.Rect(0, 0, tc.w, tc.h), int64(i))

		// When the integral-image implementation is run
		got, err := Sauvola(img, SauvolaOptions{Window: tc.win, K: tc.k, R: tc.r})
		if err != nil {
			t.Fatal(err)
		}
		want := referenceSauvola(img, tc.win, tc.k, tc.r)

		// Then it matches the brute-force reference at every pixel
		for y := range want {
			for x := range want[y] {
				// Allow off-by-one from float rounding at the .5 boundary.
				if d := int(got[y][x]) - int(want[y][x]); d < -1 || d > 1 {
					t.Fatalf("case %d: T(%d,%d) = %d, want %d", i, x, y, got[y][x], want[y][x])
				}
			}
		}
	}
}

func TestSauvolaSubImage(t *testing.T) {
	// Given 
	// A sub-image with a non-zero origin and a zero-origin copy of the same region
	big := randomGray(image.Rect(0, 0, 40, 40), 3)
	r := image.Rect(6, 4, 29, 25)
	sub := big.SubImage(r).(*image.Gray)

	cp := image.NewGray(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			cp.SetGray(x, y, big.GrayAt(r.Min.X+x, r.Min.Y+y))
		}
	}

	// When Sauvola is run on both
	a, err := Sauvola(sub, SauvolaOptions{Window: 9})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Sauvola(cp, SauvolaOptions{Window: 9})
	if err != nil {
		t.Fatal(err)
	}

	// Then the threshold maps are identical, indexed from the region's corner
	for y := range a {
		for x := range a[y] {
			if a[y][x] != b[y][x] {
				t.Fatalf("sub-image T(%d,%d) = %d, copy = %d", x, y, a[y][x], b[y][x])
			}
		}
	}
}

func TestSauvolaHandlesUnevenIllumination(t *testing.T) {
	// Given dark text (value base-60) on a background that brightens left to
	// right, where a single global threshold would fail
	const w, h = 120, 40
	img := image.NewGray(image.Rect(0, 0, w, h))
	text := func(x, y int) bool { return y%8 < 2 && x%10 < 6 } // stripes
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			base := 90 + x*130/w // 90 .. ~220
			v := base
			if text(x, y) {
				v = base - 60
			}
			img.Pix[img.PixOffset(x, y)] = clampUint8(float64(v))
		}
	}

	// When Sauvola thresholds the image
	th, err := Sauvola(img, SauvolaOptions{Window: 15, K: 0.34})
	if err != nil {
		t.Fatal(err)
	}

	// Then at most 5% of pixels are classified as the wrong class
	wrong := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			white := img.GrayAt(x, y).Y > th[y][x]
			if white == text(x, y) {
				wrong++
			}
		}
	}
	if frac := float64(wrong) / (w * h); frac > 0.05 {
		t.Errorf("misclassified %.1f%% of pixels, want <= 5%%", frac*100)
	}
}

func BenchmarkSauvola(b *testing.B) {
	img := randomGray(image.Rect(0, 0, 1024, 1024), 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Sauvola(img, SauvolaOptions{Window: 25}); err != nil {
			b.Fatal(err)
		}
	}
}