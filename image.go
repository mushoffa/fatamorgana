package fatamorgana

import (
	"fmt"
	"image"
	"sync"

	"github.com/mushoffa/fatamorgana/grayscale"
)

// Image is a data struct which embeds Go standard image library.
type Image struct {
	data   image.Image
	format ImageType
}

// NewImage returns a new [Image] struct.
func NewImage(img image.Image) *Image {

	return &Image{data: img}
}

func (p *Image) Data() image.Image {
	return p.data
}

func (p *Image) Format() string {
	return p.format.String()
}

func (p *Image) MimeType() string {
	mimeEncoding := "data:image/%s;base64,"

	switch p.format.String() {
	case "png":
		return fmt.Sprintf(mimeEncoding, "png")
	case "jpg":
		return fmt.Sprintf(mimeEncoding, "jpeg")
	default:
		return ""
	}
}

func (p *Image) Grayscale(method grayscale.MethodType) *Image {
	var wg sync.WaitGroup

	bounds := p.data.Bounds()
	gray := image.NewGray(bounds)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		wg.Add(1)
		go func(x int, wg *sync.WaitGroup, g *image.Gray) {
			defer wg.Done()

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				pixel := p.data.At(x, y)
				c := method.Convert(pixel)
				g.Set(x, y, c)
			}
		}(x, &wg, gray)
	}

	wg.Wait()

	return &Image{gray, p.format}
}

// Monochrome converts given
func (p *Image) Inverse() *Image {
	var wg sync.WaitGroup

	bounds := p.data.Bounds()
	mono := image.NewGray(bounds)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		wg.Add(1)
		go func(x int, wg *sync.WaitGroup, g *image.Gray) {
			defer wg.Done()

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				offset := p.data.(*image.Gray).PixOffset(x, y)
				g.Pix[offset] = ^p.data.(*image.Gray).Pix[offset]
			}
		}(x, &wg, mono)
	}

	wg.Wait()

	return &Image{mono, p.format}
}
