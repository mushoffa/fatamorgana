package fatamorgana

import (
	"encoding/base64"
	"image"
	"io"
	"strings"
)

func DecodeBase64(decoded string) (*Image, error) {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(decoded))
	return decode(reader)
}

func decode(reader io.Reader) (*Image, error) {
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	imgType := formatType(format)

	return &Image{
		data:   img,
		format: imgType,
	}, nil
}
