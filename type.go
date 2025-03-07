package fatamorgana

type ImageType int

const (
	UNKOWN ImageType = iota
	JPEG
	PNG
)

func (e ImageType) String() string {
	return [...]string{"unknown", "jpeg", "png"}[e]
}

func formatType(format string) ImageType {
	switch format {
	case "jpeg","jpg":
		return JPEG
	case PNG.String():
		return PNG
	}

	return UNKOWN
}
