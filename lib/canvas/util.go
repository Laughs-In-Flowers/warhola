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

// A package level representation of a file type.
type FileExtension int

// A method satisfying any String interface for a FileExtension.
func (f FileExtension) String() string {
	switch f {
	case JPG:
		return "JPG"
	case PNG:
		return "PNG"
	}
	return ""
}

func stringToFileExtension(f string) FileExtension {
	switch strings.ToUpper(f) {
	case "JPG", "JPEG":
		return JPG
	case "PNG":
		return PNG
	}
	return PNG
}

const (
	JPG FileExtension = iota
	PNG
)

// Opens a draw.Image from the provided string path and color.Model.
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

func splitPath(p string) (string, string, FileExtension) {
	d, f := filepath.Split(p)
	spl := strings.Split(f, ".")
	var n, x string
	if len(spl) == 2 {
		n, x = spl[0], spl[1]
	}
	return d, n, stringToFileExtension(x)
}

// Saves a image.Image to the provided path.
func SaveImage(p string, i image.Image) error {
	_, _, x := splitPath(p)
	f, err := Fopen(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return EncodeImage(f, i, x)
}

// Given an io.reader, decodes an image returning it and any error.
func DecodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// Given an io.Writer, an image.Image and a FileExtension, attempts to encode the image,
// returning any error.
func EncodeImage(w io.Writer, i image.Image, x FileExtension) error {
	switch x {
	case JPG:
		return EncodeJpg(w, i)
	case PNG:
		return EncodePng(w, i)
	}
	return UnrecognizableExtension(x)
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

var OpenError = Xrror("unable to find or open file %s, provided %s").Out

func exist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir|0755)
	}
}

//
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
	newGray    = func(r image.Rectangle) draw.Image { return image.NewGray(r) }
	newGray16  = func(r image.Rectangle) draw.Image { return image.NewGray16(r) }
	newAlpha   = func(r image.Rectangle) draw.Image { return image.NewAlpha(r) }
	newAlpha16 = func(r image.Rectangle) draw.Image { return image.NewAlpha16(r) }
	newRGBA    = func(r image.Rectangle) draw.Image { return image.NewRGBA(r) }
	newRGBA64  = func(r image.Rectangle) draw.Image { return image.NewRGBA64(r) }
	newNRGBA   = func(r image.Rectangle) draw.Image { return image.NewNRGBA(r) }
	newNRGBA64 = func(r image.Rectangle) draw.Image { return image.NewNRGBA64(r) }
	newCMYK    = func(r image.Rectangle) draw.Image { return image.NewCMYK(r) }
)

// A new draw.Image from the provided color.Model of the specified X,Y size.
func NewFrom(m color.Model, X, Y int) draw.Image {
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

// Clone a image.Image with the specified color.Model, return a draw.Image.
func Clone(i image.Image, m color.Model) draw.Image {
	switch m {
	case color.GrayModel:
		return cloneAs(i, newGray)
	case color.Gray16Model:
		return cloneAs(i, newGray16)
	case color.AlphaModel:
		return cloneAs(i, newAlpha)
	case color.Alpha16Model:
		return cloneAs(i, newAlpha16)
	case color.RGBAModel:
		return cloneAs(i, newRGBA)
	case color.RGBA64Model:
		return cloneAs(i, newRGBA64)
	case color.NRGBAModel:
		return cloneAs(i, newNRGBA)
	case color.NRGBA64Model:
		return cloneAs(i, newNRGBA64)
	case color.CMYKModel:
		return cloneAs(i, newCMYK)
	}
	return cloneAs(i, newRGBA)
}

func cloneAs(src image.Image, fn func(image.Rectangle) draw.Image) draw.Image {
	bounds := src.Bounds()
	img := fn(bounds)
	draw.Draw(img, bounds, src, bounds.Min, draw.Src)
	return img
}

// Returns a color.Model from the priovided string, defaulting to color.RGBAModel.
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
