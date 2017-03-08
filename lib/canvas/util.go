package canvas

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type fileExtension int

func (f fileExtension) String() string {
	switch f {
	case JPG:
		return "JPG"
	case PNG:
		return "PNG"
	}
	return ""
}

func stringToFileExtension(f string) fileExtension {
	switch strings.ToUpper(f) {
	case "JPG", "JPEG":
		return JPG
	case "PNG":
		return PNG
	}
	return PNG
}

const (
	JPG fileExtension = iota
	PNG
)

func OpenImage(p string, c color.Model) (draw.Image, error) {
	file, err := Fopen(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	i, err := DecodeImage(file)
	if err != nil {
		return nil, err
	}

	ri := Clone(i, c)
	return ri, nil
}

var UnrecognizableExtension = Xrror("%s is not a recognized extension for image encoding.").Out

func splitPath(p string) (string, string, fileExtension) {
	d, f := filepath.Split(p)
	spl := strings.Split(f, ".")
	var n, x string
	if len(spl) == 2 {
		n, x = spl[0], spl[1]
	}
	return d, n, stringToFileExtension(x)
}

func SaveImage(p string, i image.Image) error {
	_, _, x := splitPath(p)
	f, err := Fopen(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return EncodeImage(f, i, x)
}

func DecodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func EncodeImage(w io.Writer, i image.Image, x fileExtension) error {
	switch x {
	case JPG:
		return EncodeJpg(w, i)
	case PNG:
		return EncodePng(w, i)
	}
	return UnrecognizableExtension(x)
}

func EncodeJpg(w io.Writer, i image.Image) error {
	if err := jpeg.Encode(w, i, &jpeg.Options{100}); err != nil {
		return err
	}
	return nil
}

func EncodePng(w io.Writer, i image.Image) error {
	if err := png.Encode(w, i); err != nil {
		return err
	}
	return nil
}

var OpenError = Xrror("unable to find or open file %s, provided %s").Out

func exist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir|0755)
	}
}

func Fopen(path string) (*os.File, error) {
	p := filepath.Clean(path)

	dir, name := filepath.Split(p)

	var fp string
	var err error
	switch dir {
	case "":
		fp, err = filepath.Abs(name)
	default:
		exist(dir)
		fp, err = filepath.Abs(p)
	}
	if err != nil {
		return nil, err
	}

	if file, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE, 0660); err == nil {
		return file, nil
	}

	return nil, OpenError(fp, path)
}

var (
	NewGray    = func(r image.Rectangle) draw.Image { return image.NewGray(r) }
	NewGray16  = func(r image.Rectangle) draw.Image { return image.NewGray16(r) }
	NewAlpha   = func(r image.Rectangle) draw.Image { return image.NewAlpha(r) }
	NewAlpha16 = func(r image.Rectangle) draw.Image { return image.NewAlpha16(r) }
	NewRGBA    = func(r image.Rectangle) draw.Image { return image.NewRGBA(r) }
	NewRGBA64  = func(r image.Rectangle) draw.Image { return image.NewRGBA64(r) }
	NewNRGBA   = func(r image.Rectangle) draw.Image { return image.NewNRGBA(r) }
	NewNRGBA64 = func(r image.Rectangle) draw.Image { return image.NewNRGBA64(r) }
	NewCMYK    = func(r image.Rectangle) draw.Image { return image.NewCMYK(r) }
)

func NewFrom(m color.Model, X, Y int) draw.Image {
	r := image.Rect(0, 0, X, Y)
	switch m {
	case color.GrayModel:
		return NewGray(r)
	case color.Gray16Model:
		return NewGray16(r)
	case color.AlphaModel:
		return NewAlpha(r)
	case color.Alpha16Model:
		return NewAlpha16(r)
	case color.RGBAModel:
		return NewRGBA(r)
	case color.RGBA64Model:
		return NewRGBA64(r)
	case color.NRGBAModel:
		return NewNRGBA(r)
	case color.NRGBA64Model:
		return NewNRGBA64(r)
	case color.CMYKModel:
		return NewCMYK(r)
	}
	return NewRGBA(r)
}

func Clone(i image.Image, m color.Model) draw.Image {
	switch m {
	case color.GrayModel:
		return cloneAs(i, NewGray)
	case color.Gray16Model:
		return cloneAs(i, NewGray16)
	case color.AlphaModel:
		return cloneAs(i, NewAlpha)
	case color.Alpha16Model:
		return cloneAs(i, NewAlpha16)
	case color.RGBAModel:
		return cloneAs(i, NewRGBA)
	case color.RGBA64Model:
		return cloneAs(i, NewRGBA64)
	case color.NRGBAModel:
		return cloneAs(i, NewNRGBA)
	case color.NRGBA64Model:
		return cloneAs(i, NewNRGBA64)
	case color.CMYKModel:
		return cloneAs(i, NewCMYK)
	}
	return cloneAs(i, NewRGBA)
}

func cloneAs(src image.Image, fn func(image.Rectangle) draw.Image) draw.Image {
	bounds := src.Bounds()
	img := fn(bounds)
	draw.Draw(img, bounds, src, bounds.Min, draw.Src)
	return img
}

func StringToColorModel(s string) color.Model {
	switch s {
	case "gray":
		return color.GrayModel
	case "gray16":
		return color.Gray16Model
	case "alpha":
		return color.AlphaModel
	case "alpha16":
		return color.Alpha16Model
	case "rgba":
		return color.RGBAModel
	case "rgba64":
		return color.RGBA64Model
	case "nrgba":
		return color.NRGBAModel
	case "nrgba64":
		return color.NRGBA64Model
	case "cmyk":
		return color.CMYKModel
	}
	return color.RGBAModel
}
