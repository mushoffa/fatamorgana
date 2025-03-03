package fatamorgana

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

type Encode func(io.Writer, image.Image) error

var (
	JPEGEncoder Encode = jpegEncoder()
	PNGEncoder  Encode = pngEncoder()
)

var encoders = map[ImageType]Encode{
	JPEG: jpegEncoder(),
	PNG:  pngEncoder(),
}

func (p *Image) Base64() (string, error) {
	bytes, err := p.Bytes()
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(bytes)

	return encoded, nil
}

func (p *Image) Bytes() ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := encoders[p.format]

	if err := encoder(buffer, p.data); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func jpegEncoder() Encode {
	return func(w io.Writer, img image.Image) error {
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 100})
	}
}

func pngEncoder() Encode {
	return func(w io.Writer, img image.Image) error {
		return png.Encode(w, img)
	}
}
