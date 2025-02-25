package fatamorgana

import (
	"fmt"
	"image"
	"os"
)

func Open(path string) (*Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	imgType := mimetype(format)

	return &Image{
		data:     img,
		mimetype: imgType,
	}, nil
}

func (p *Image) Save(path, filename string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s.%s", path, filename, p.mimetype.String())
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := encoders[p.mimetype]

	return encoder(file, p.data)
}
