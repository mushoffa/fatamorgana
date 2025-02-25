package fatamorgana

import (
	"image"
	"image/color"
	"sync"

	"github.com/mushoffa/fatamorgana/grayscale"
)

// Image is a data struct which embeds Go standard image library.
type Image struct {
	data     image.Image
	mimetype ImageType
}

// NewImage returns a new [Image] struct.
func NewImage(img image.Image) *Image {

	return &Image{data: img}
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

	return &Image{gray, p.mimetype}
}

// Monochrome converts given
func (p *Image) Monochrome(method grayscale.MethodType, threshold uint8) *Image {
	var wg sync.WaitGroup

	bounds := p.data.Bounds()
	mono := image.NewGray(bounds)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		wg.Add(1)
		go func(x int, wg *sync.WaitGroup, g *image.Gray) {
			defer wg.Done()

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				pixel := p.data.At(x, y)
				c := method.Convert(pixel).(color.Gray).Y
				b := color.Black
				if c > threshold {
					b = color.White
				}
				g.Set(x, y, b)
			}
		}(x, &wg, mono)
	}

	wg.Wait()

	return &Image{mono, p.mimetype}
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

	return &Image{mono, p.mimetype}
}
