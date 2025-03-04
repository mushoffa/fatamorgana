package fatamorgana

import (
	"image"
)

type CropRectangle interface {
	SubImage(image.Rectangle) image.Image
}

func (p *Image) Crop(minX, minY, maxX, maxY int) image.Image {
	r := image.Rect(minX, minY, maxX, maxY)
	return p.data.(CropRectangle).SubImage(r)
}
