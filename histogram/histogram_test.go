package histogram

import (
	"image"
	"image/color"
	"reflect"
	"testing"
)

// newGray builds a white image of the given size with black pixels at the
// listed points.
func newGray(rect image.Rectangle, black ...image.Point) *image.Gray {
	g := image.NewGray(rect)
	for i := range g.Pix {
		g.Pix[i] = 255
	}
	for _, p := range black {
		g.SetGray(p.X, p.Y, color.Gray{Y: 0})
	}
	return g
}

func TestBinary(t *testing.T) {
	// 4x3 image, black pixels at (1,0) (2,0) (2,2):
	//
	//	. # # .
	//	. . . .
	//	. . # .
	base := newGray(image.Rect(0, 0, 4, 3),
		image.Pt(1, 0), image.Pt(2, 0), image.Pt(2, 2))

	tests := []struct {
		name string
		img  *image.Gray
		axis HistogramProfile
		want []int
	}{
		{"horizontal counts per row", base, Horizontal, []int{2, 0, 1}},
		{"vertical counts per column", base, Vertical, []int{0, 1, 2, 0}},
		{
			"all white horizontal",
			newGray(image.Rect(0, 0, 3, 2)), Horizontal, []int{0, 0},
		},
		{
			"all white vertical",
			newGray(image.Rect(0, 0, 3, 2)), Vertical, []int{0, 0, 0},
		},
		{
			"all black horizontal",
			func() *image.Gray {
				g := newGray(image.Rect(0, 0, 3, 2))
				for i := range g.Pix {
					g.Pix[i] = 0
				}
				return g
			}(),
			Horizontal, []int{3, 3},
		},
		{
			"gray values other than zero are not black",
			func() *image.Gray {
				g := newGray(image.Rect(0, 0, 3, 1))
				g.SetGray(0, 0, color.Gray{Y: 1})
				g.SetGray(1, 0, color.Gray{Y: 128})
				g.SetGray(2, 0, color.Gray{Y: 0})
				return g
			}(),
			Horizontal, []int{1},
		},
		{
			"single row",
			newGray(image.Rect(0, 0, 5, 1), image.Pt(0, 0), image.Pt(4, 0)),
			Horizontal, []int{2},
		},
		{
			"single column",
			newGray(image.Rect(0, 0, 1, 4), image.Pt(0, 1), image.Pt(0, 3)),
			Vertical, []int{2},
		},
		{"nil image horizontal", nil, Horizontal, []int{}},
		{"nil image vertical", nil, Vertical, []int{}},
		{"empty image horizontal", image.NewGray(image.Rect(0, 0, 0, 0)), Horizontal, []int{}},
		{"empty image vertical", image.NewGray(image.Rect(0, 0, 0, 0)), Vertical, []int{}},
		{
			"zero height keeps column count",
			image.NewGray(image.Rect(0, 0, 3, 0)), Vertical, []int{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			// An image and an axis from the table above

			// When
			// Binary is called with them
			got := Binary(tt.img, tt.axis)

			// Then
			// The result is a non-nil slice holding the expected counts
			if got == nil {
				t.Fatal("Binary returned nil, want a non-nil slice")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Binary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBinarySubImage(t *testing.T) {
	// Given
	// A sub-image whose origin is not (0, 0): rows 1-2 and columns 1-3 of
	//
	//	. # # .
	//	. . . .
	//	. . # .
	//
	// which is
	//
	//	. . .
	//	. # .
	base := newGray(image.Rect(0, 0, 4, 3),
		image.Pt(1, 0), image.Pt(2, 0), image.Pt(2, 2))
	sub := base.SubImage(image.Rect(1, 1, 4, 3)).(*image.Gray)

	// When
	// Binary is called for each axis
	horizontal := Binary(sub, Horizontal)
	vertical := Binary(sub, Vertical)

	// Then
	// The results are indexed from the sub-image's own origin, with one entry
	// per row or column of the sub-image
	if want := []int{0, 1}; !reflect.DeepEqual(horizontal, want) {
		t.Errorf("Horizontal = %v, want %v", horizontal, want)
	}
	if want := []int{0, 1, 0}; !reflect.DeepEqual(vertical, want) {
		t.Errorf("Vertical = %v, want %v", vertical, want)
	}
}

func TestBinaryDoesNotModifyImage(t *testing.T) {
	// Given
	// An image and a copy of its pixels
	img := newGray(image.Rect(0, 0, 4, 3), image.Pt(1, 0), image.Pt(2, 2))
	before := append([]uint8(nil), img.Pix...)

	// When
	// Binary is called for each axis
	Binary(img, Horizontal)
	Binary(img, Vertical)

	// Then
	// The image pixels are unchanged
	if !reflect.DeepEqual(img.Pix, before) {
		t.Error("Binary modified the image pixels")
	}
}

// For images whose origin is (0, 0), Binary must agree with Monochrome.
func TestBinaryMatchesMonochrome(t *testing.T) {
	// Given
	// An image whose origin is (0, 0)
	img := newGray(image.Rect(0, 0, 5, 4),
		image.Pt(0, 0), image.Pt(1, 0), image.Pt(4, 1),
		image.Pt(2, 3), image.Pt(3, 3), image.Pt(4, 3))

	for _, axis := range []HistogramProfile{Horizontal, Vertical} {
		// When
		// Both Monochrome and Binary are called for the same axis
		old := Monochrome(img, axis)
		got := Binary(img, axis)

		// Then
		// They return the same counts
		if len(old) != len(got) {
			t.Fatalf("axis %d: length %d, want %d", axis, len(got), len(old))
		}
		for i := range old {
			if int(old[i]) != got[i] {
				t.Errorf("axis %d index %d: got %d, want %d", axis, i, got[i], old[i])
			}
		}
	}
}

func BenchmarkBinary(b *testing.B) {
	// Given
	// An A4 page at 300 dpi with a regular pattern of black pixels
	img := newGray(image.Rect(0, 0, 2480, 3508))
	for y := 0; y < 3508; y += 3 {
		for x := 0; x < 2480; x += 7 {
			img.SetGray(x, y, color.Gray{Y: 0})
		}
	}

	for _, axis := range []struct {
		name string
		axis HistogramProfile
	}{{"horizontal", Horizontal}, {"vertical", Vertical}} {
		b.Run(axis.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// When
				// Binary is called repeatedly
				Binary(img, axis.axis)
			}
		})
	}
}
