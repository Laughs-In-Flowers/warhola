package canvas

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/Laughs-In-Flowers/xrr"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

// Indicates the type of file the canvas is.
type FileType int

const (
	FILETYPENOOP FileType = iota
	BMP
	JPG
	PNG
	TIFF
)

//A variable containing a listing of available and fully functional FileType.
var AvailableFileType = []FileType{
	BMP,
	JPG,
	PNG,
	TIFF,
}

func stringToFileType(s string) FileType {
	switch strings.ToUpper(s) {
	case "BMP":
		return BMP
	case "JPEG", "JPG":
		return JPG
	case "PNG":
		return PNG
	case "TIFF":
		return TIFF
	}
	return FILETYPENOOP
}

// Provides a string of this FileType.
func (t FileType) String() string {
	switch t {
	case BMP:
		return "bmp"
	case JPG:
		return "jpg"
	case PNG:
		return "png"
	case TIFF:
		return "tiff"
	}
	return "FileTypeNoop"
}

var (
	encodeFileTypeError       = xrr.Xrror("unable to encode: FileTypeNoop")
	unrecognizedFileTypeError = xrr.Xrror("%s is not a recognized filetype").Out
)

func (t FileType) encode(f *os.File, p *pxl) error {
	switch t {
	case BMP:
		return encodeBmp(f, p)
	case JPG:
		return encodeJpg(f, p)
	case PNG:
		return encodePng(f, p)
	case TIFF:
		return encodeTiff(f, p)
	}
	return encodeFileTypeError
}

func encodeBmp(w io.Writer, i image.Image) error {
	if err := bmp.Encode(w, i); err != nil {
		return err
	}
	return nil
}

var jpgQuality int = 100

func encodeJpg(w io.Writer, i image.Image) error {
	if err := jpeg.Encode(w, i, &jpeg.Options{jpgQuality}); err != nil {
		return err
	}
	return nil
}

func encodePng(w io.Writer, i image.Image) error {
	if err := png.Encode(w, i); err != nil {
		return err
	}
	return nil
}

var (
	tiffCompression tiff.CompressionType = tiff.Uncompressed
	tiffPredictor   bool                 = false
)

func encodeTiff(w io.Writer, i image.Image) error {
	if err := tiff.Encode(w, i, &tiff.Options{tiffCompression, tiffPredictor}); err != nil {
		return err
	}
	return nil
}
