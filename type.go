package fatamorgana

type ImageType int

const (
	UNKOWN ImageType = iota
	JPEG
	PNG
)

func (e ImageType) String() string {
	return [...]string{"unkown", "jpg", "png"}[e]
}

func mimetype(format string) ImageType {
	switch format {
	case JPEG.String():
		return JPEG
	case PNG.String():
		return PNG
	}

	return UNKOWN
}
