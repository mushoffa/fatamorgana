package fatamorgana

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"github.com/mushoffa/fatamorgana/grayscale"
)

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

	return &Image{mono, p.format}
}

// Monochrome converts given
func (p *Image) MonochromeAdaptive(thresholds [][]uint8) *Image {
	var wg sync.WaitGroup

	bounds := p.data.Bounds()
	mono := image.NewGray(bounds)

	fmt.Println("Max_X: ", bounds.Max.X)
	fmt.Println("Max_Y: ", bounds.Max.Y)
	fmt.Println("Max_T1: ", len(thresholds))
	fmt.Println("Max_T2: ", len(thresholds[0]))

	// fmt.Println(thresholds)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		wg.Add(1)
		go func(x int, wg *sync.WaitGroup, g *image.Gray) {
			defer wg.Done()

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				threshold := thresholds[y][x]
				pixel := p.data.At(x, y)
				c := pixel.(color.Gray).Y
				b := color.Black
				if c > threshold {
					b = color.White
				}
				g.Set(x, y, b)
			}
		}(x, &wg, mono)
	}

	wg.Wait()

	return &Image{mono, p.format}
}
