package fatamorgana

import (
	"fmt"
	"os"
)

func Open(path string) (*Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return decode(file)
}

func (p *Image) Save(path, filename string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s.%s", path, filename, p.format.String())
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := encoders[p.format]

	return encoder(file, p.data)
}
