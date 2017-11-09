package canvas

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/Laughs-In-Flowers/warhola/lib/util/xrr"
)

type Image interface {
	image.Image
	Set(x, y int, c color.Color)
}

// Opens an Image from the provided string path and color.Model.
func OpenImage(p string, c color.Model) (Image, string, error) {
	file, err := Fopen(p)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	i, ext, err := DecodeImage(file)
	if err != nil {
		return nil, "", err
	}

	ri := Clone(i, c)
	return ri, ext, nil
}

// Given an io.reader, decodes an image returning it and any error.
func DecodeImage(r io.Reader) (image.Image, string, error) {
	img, ext, err := image.Decode(r)
	if err != nil {
		return nil, "", err
	}
	return img, ext, nil
}

// Saves a image.Image to the provided path.
func SaveImage(path string, k Kind, i image.Image) error {
	f, err := Fopen(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return EncodeImage(f, i, k)
}

// Given an io.Writer, an image.Image and a FileExtension, attempts to encode the image,
// returning any error.
func EncodeImage(w io.Writer, i image.Image, k Kind) error {
	switch k {
	case JPG:
		return EncodeJpg(w, i)
	case PNG:
		return EncodePng(w, i)
	}
	return KindError(k)
}

// Encode a jpeg given an io.Writer and image.Image.
func EncodeJpg(w io.Writer, i image.Image) error {
	if err := jpeg.Encode(w, i, &jpeg.Options{100}); err != nil {
		return err
	}
	return nil
}

// Encode a png given an io.Writer and image.Image.
func EncodePng(w io.Writer, i image.Image) error {
	if err := png.Encode(w, i); err != nil {
		return err
	}
	return nil
}

var (
	OpenError = xrr.Xrror("unable to find or open file %s, provided %s").Out
	KindError = xrr.Xrror("%s is not a recognized extension for image encoding.").Out
)

func exist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir|0755)
	}
}

// Return an *os.File pointer from the given path.
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
	newGray    = func(r image.Rectangle) Image { return image.NewGray(r) }
	newGray16  = func(r image.Rectangle) Image { return image.NewGray16(r) }
	newAlpha   = func(r image.Rectangle) Image { return image.NewAlpha(r) }
	newAlpha16 = func(r image.Rectangle) Image { return image.NewAlpha16(r) }
	newRGBA    = func(r image.Rectangle) Image { return image.NewRGBA(r) }
	newRGBA64  = func(r image.Rectangle) Image { return image.NewRGBA64(r) }
	newNRGBA   = func(r image.Rectangle) Image { return image.NewNRGBA(r) }
	newNRGBA64 = func(r image.Rectangle) Image { return image.NewNRGBA64(r) }
	newCMYK    = func(r image.Rectangle) Image { return image.NewCMYK(r) }
)

// A new Image from the provided color.Model of the specified X,Y size.
func NewFrom(m color.Model, X, Y int) Image {
	r := image.Rect(0, 0, X, Y)
	switch m {
	case color.GrayModel:
		return newGray(r)
	case color.Gray16Model:
		return newGray16(r)
	case color.AlphaModel:
		return newAlpha(r)
	case color.Alpha16Model:
		return newAlpha16(r)
	case color.RGBAModel:
		return newRGBA(r)
	case color.RGBA64Model:
		return newRGBA64(r)
	case color.NRGBAModel:
		return newNRGBA(r)
	case color.NRGBA64Model:
		return newNRGBA64(r)
	case color.CMYKModel:
		return newCMYK(r)
	}
	return newRGBA(r)
}

// Clone an image.Image with the specified color.Model and return as Image.
func Clone(src image.Image, m color.Model) Image {
	switch m {
	case color.GrayModel:
		return clone(src, newGray)
	case color.Gray16Model:
		return clone(src, newGray16)
	case color.AlphaModel:
		return clone(src, newAlpha)
	case color.Alpha16Model:
		return clone(src, newAlpha16)
	case color.RGBAModel:
		return clone(src, newRGBA)
	case color.RGBA64Model:
		return clone(src, newRGBA64)
	case color.NRGBAModel:
		return clone(src, newNRGBA)
	case color.NRGBA64Model:
		return clone(src, newNRGBA64)
	case color.CMYKModel:
		return clone(src, newCMYK)
	}
	return clone(src, newRGBA)
}

func clone(src image.Image, fn func(image.Rectangle) Image) Image {
	bounds := src.Bounds()
	dst := fn(bounds)
	draw(dst, dst.Bounds(), src, bounds.Min, nil, image.Point{}, Src)
	return dst
}

func paste(src, dst Image, p Point) Image {
	sp := p.IPoint()
	draw(dst, dst.Bounds(), src, sp, nil, image.Point{}, Src)
	return dst
}

func overlay(src, dst Image, p Point, o float64) Image {
	sp := p.IPoint()
	m := image.NewUniform(color.Alpha{uint8(255 * (o / 100))})
	draw(dst, dst.Bounds(), src, sp, m, image.Point{}, Over)
	return dst
}
