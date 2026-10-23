package threshold

import (
	"errors"
	"image"
	"math/rand"
	"testing"
)

// filled returns a w×h image where every pixel has the value v.
func filled(w, h int, v uint8) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = v
	}
	return img
}

// referenceOtsu is a slow, obviously-correct implementation: it evaluates the
// between-class variance for every threshold from scratch and returns the
// midpoint of the first and last maximising thresholds. ok is false when no
// valid split exists (fewer than two occupied levels).
func referenceOtsu(h [Levels]int) (t int, ok bool) {
	best := -1.0
	var ts []int
	for th := 0; th < Levels-1; th++ {
		var w0, s0, w1, s1 float64
		for l, c := range h {
			if l <= th {
				w0 += float64(c)
				s0 += float64(l * c)
			} else {
				w1 += float64(c)
				s1 += float64(l * c)
			}
		}
		if w0 == 0 || w1 == 0 {
			continue
		}
		d := s0/w0 - s1/w1
		v := w0 * w1 * d * d
		switch {
		case v > best:
			best, ts = v, []int{th}
		case v == best:
			ts = append(ts, th)
		}
	}
	if len(ts) == 0 {
		return 0, false
	}
	return (ts[0] + ts[len(ts)-1]) / 2, true
}

func TestHistogram(t *testing.T) {
	t.Run("nil image", func(t *testing.T) {
		// Given a nil image
		// When Histogram is called
		got := Histogram(nil)

		// Then every bin is zero
		if got != [Levels]int{} {
			t.Errorf("Histogram(nil) = non-zero histogram")
		}
	})

	t.Run("empty image", func(t *testing.T) {
		// Given an image with no pixels
		img := image.NewGray(image.Rect(0, 0, 0, 0))

		// When Histogram is called
		got := Histogram(img)

		// Then every bin is zero
		if got != [Levels]int{} {
			t.Errorf("empty image gave non-zero histogram")
		}
	})

	t.Run("counts", func(t *testing.T) {
		// Given a 3x2 image with two 0s, three 7s and one 255
		img := image.NewGray(image.Rect(0, 0, 3, 2))
		copy(img.Pix, []uint8{0, 0, 255, 7, 7, 7})

		// When Histogram is called
		got := Histogram(img)

		// Then each bin holds the number of pixels with that value
		want := map[int]int{0: 2, 7: 3, 255: 1}
		total := 0
		for level, c := range got {
			total += c
			if c != want[level] {
				t.Errorf("hist[%d] = %d, want %d", level, c, want[level])
			}
		}
		if total != 6 {
			t.Errorf("total = %d, want 6", total)
		}
	})

	t.Run("sub-image with non-zero origin", func(t *testing.T) {
		// Given a 4x4 sub-image (value 200) with a non-zero origin inside a
		// larger image of zeros
		big := filled(10, 10, 0)
		sub := big.SubImage(image.Rect(2, 3, 6, 7)).(*image.Gray)
		for y := 3; y < 7; y++ {
			for x := 2; x < 6; x++ {
				big.Pix[big.PixOffset(x, y)] = 200
			}
		}

		// When Histogram is called on the sub-image
		got := Histogram(sub)

		// Then only the sub-image's 16 pixels are counted
		if got[200] != 16 || got[0] != 0 {
			t.Errorf("sub-image hist[200]=%d hist[0]=%d, want 16 and 0", got[200], got[0])
		}
	})
}

func TestOtsuEmpty(t *testing.T) {
	// Given a nil image
	// When Otsu is called
	// Then it returns ErrEmptyImage
	if _, err := Otsu(nil); !errors.Is(err, ErrEmptyImage) {
		t.Errorf("Otsu(nil) err = %v, want ErrEmptyImage", err)
	}

	// Given an image with no pixels
	// When Otsu is called
	// Then it returns ErrEmptyImage
	if _, err := Otsu(image.NewGray(image.Rect(0, 0, 0, 0))); !errors.Is(err, ErrEmptyImage) {
		t.Errorf("Otsu(empty) err = %v, want ErrEmptyImage", err)
	}

	// Given a histogram with no counts
	// When OtsuHistogram is called
	// Then it returns ErrEmptyImage
	if _, err := OtsuHistogram([Levels]int{}); !errors.Is(err, ErrEmptyImage) {
		t.Errorf("OtsuHistogram(zero) err = %v, want ErrEmptyImage", err)
	}
}

func TestOtsuSingleLevel(t *testing.T) {
	for _, v := range []uint8{0, 1, 77, 128, 254, 255} {
		// Given an image where every pixel has the same value v
		img := filled(5, 5, v)

		// When Otsu is called
		got, err := Otsu(img)

		// Then no split exists and the threshold is v itself
		if err != nil {
			t.Errorf("level %d: unexpected error %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("level %d: threshold = %d, want %d", v, got, v)
		}
	}
}

func TestOtsuKnownValues(t *testing.T) {
	tests := []struct {
		name string
		hist map[int]int
		want uint8
	}{
		// Variance is flat for t in [lo, hi-1]; expect the midpoint.
		{"two levels 50/200", map[int]int{50: 10, 200: 10}, 124},
		{"two levels 0/255", map[int]int{0: 8, 255: 8}, 127},
		{"two adjacent levels", map[int]int{100: 5, 101: 5}, 100},
		{"two levels 0/1", map[int]int{0: 3, 1: 9}, 0},
		{"two levels 254/255", map[int]int{254: 3, 255: 9}, 254},
		// Unequal masses still split between the two clusters.
		{"unbalanced 30/220", map[int]int{30: 90, 220: 10}, 124},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given a histogram with the listed levels occupied
			var h [Levels]int
			for l, c := range tc.hist {
				h[l] = c
			}

			// When OtsuHistogram is called
			got, err := OtsuHistogram(h)
			if err != nil {
				t.Fatal(err)
			}

			// Then the threshold is the midpoint of the flat variance plateau
			if got != tc.want {
				t.Errorf("threshold = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOtsuSeparatesClusters(t *testing.T) {
	// Given an image of two noisy clusters around 60 and 190
	rng := rand.New(rand.NewSource(1))
	img := image.NewGray(image.Rect(0, 0, 64, 64))
	for i := range img.Pix {
		center := 60
		if i%2 == 0 {
			center = 190
		}
		img.Pix[i] = uint8(center + rng.Intn(21) - 10)
	}

	// When Otsu is called
	got, err := Otsu(img)
	if err != nil {
		t.Fatal(err)
	}

	// Then the threshold falls between the two clusters
	if got < 70 || got >= 180 {
		t.Errorf("threshold = %d, want within [70, 180)", got)
	}
}

func TestOtsuMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	checked := 0
	for i := 0; i < 3000; i++ {
		// Given a random sparse histogram with at least two occupied levels
		var h [Levels]int
		for j, n := 0, rng.Intn(60)+2; j < n; j++ {
			h[rng.Intn(Levels)] += rng.Intn(100) + 1
		}
		want, ok := referenceOtsu(h)
		if !ok {
			continue // single occupied level; covered separately
		}
		checked++

		// When OtsuHistogram is called
		got, err := OtsuHistogram(h)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}

		// Then it matches the brute-force reference
		if int(got) != want {
			t.Fatalf("iteration %d: got %d, want %d", i, got, want)
		}
	}
	if checked == 0 {
		t.Fatal("no histograms were compared")
	}
}

func TestOtsuSubImageMatchesCopy(t *testing.T) {
	// Given a sub-image with a non-zero origin and a zero-origin copy of the
	// same region
	rng := rand.New(rand.NewSource(7))
	big := image.NewGray(image.Rect(0, 0, 40, 40))
	for i := range big.Pix {
		big.Pix[i] = uint8(rng.Intn(Levels))
	}
	r := image.Rect(5, 9, 31, 28)
	sub := big.SubImage(r).(*image.Gray)

	// Copy the region into a zero-origin image.
	cp := image.NewGray(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			cp.Pix[cp.PixOffset(x, y)] = big.Pix[big.PixOffset(r.Min.X+x, r.Min.Y+y)]
		}
	}

	// When Histogram and Otsu are run on both
	// Then the histograms and thresholds are identical
	if Histogram(sub) != Histogram(cp) {
		t.Fatal("sub-image and copy histograms differ")
	}
	a, _ := Otsu(sub)
	b, _ := Otsu(cp)
	if a != b {
		t.Errorf("Otsu(sub) = %d, Otsu(copy) = %d", a, b)
	}
}

func TestOtsuIgnoresPixelsOutsideSubImage(t *testing.T) {
	// Given a uniform sub-image (value 40) surrounded by a bright border
	big := filled(10, 10, 255)
	sub := big.SubImage(image.Rect(2, 2, 8, 8)).(*image.Gray)
	for y := 2; y < 8; y++ {
		for x := 2; x < 8; x++ {
			big.Pix[big.PixOffset(x, y)] = 40
		}
	}

	// When Otsu is called on the sub-image
	got, err := Otsu(sub)
	if err != nil {
		t.Fatal(err)
	}

	// Then the border is ignored and the single level 40 is returned
	if got != 40 {
		t.Errorf("threshold = %d, want 40 (single level in sub-image)", got)
	}
}

func BenchmarkOtsu(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	img := image.NewGray(image.Rect(0, 0, 1024, 1024))
	for i := range img.Pix {
		img.Pix[i] = uint8(rng.Intn(Levels))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Otsu(img); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOtsuHistogram(b *testing.B) {
	var h [Levels]int
	for i := range h {
		h[i] = i + 1
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := OtsuHistogram(h); err != nil {
			b.Fatal(err)
		}
	}
}