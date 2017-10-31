package canvas

import (
	"image/color"
	"image/draw"
	"os"

	"github.com/Laughs-In-Flowers/log"
)

// An interface for manipulating a single draw.Image.
type Canvas interface {
	draw.Image
	Factory
	Noop() bool
	GetImage() draw.Image
	SetImage(draw.Image)
	Save() error
}

type canvas struct {
	draw.Image
	path, extension string
	model           color.Model
	action          Action
	*factory
}

type options struct {
	ColorM    color.Model
	Path      string
	Extension string
	X, Y      int
	PP        int
	PPUnit    string
}

// Provide options by function when creating a canvas
// (color model, path, extension, x, y, points per, points per unit).
func Options(colorm string, path, extension string, x, y, pp int, ppu string) options {
	return options{
		ColorM:    StringToColorModel(colorm),
		Path:      path,
		Extension: extension,
		X:         x,
		Y:         y,
		PP:        pp,
		PPUnit:    ppu,
	}
}

const DefaultPath = "DEFAULT"

type Action int

const (
	ImageNoop Action = iota
	ImageNew
	ImageOpen
)

func newAction(o options) Action {
	if o.Path != DefaultPath {
		_, err := os.Stat(o.Path)
		switch {
		case os.IsNotExist(err):
			return ImageNew
		default:
			return ImageOpen
		}
	}
	return ImageNoop
}

// Given options(wrapped via Options function), produces a Canvas interface.
func New(o options, l log.Logger) Canvas {
	var i draw.Image
	var err error
	action := newAction(o)
	switch action {
	case ImageNoop:
		i = NewFrom(o.ColorM, 1, 1)
		l.Println("image is noop")
	case ImageNew:
		i = NewFrom(o.ColorM, o.X, o.Y)
		err = SaveImage(o.Path, o.Extension, i)
		if err != nil {
			l.Fatalln(err)
		}
		l.Printf("image is new: %s", o.Path)
	case ImageOpen:
		i, _, err = OpenImage(o.Path, o.ColorM)
		if err != nil {
			l.Fatalln(err)
		}
		l.Printf("image exists: %s", o.Path)
	}
	ni := Clone(i, o.ColorM)
	c := &canvas{
		Image:     ni,
		path:      o.Path,
		extension: o.Extension,
		model:     o.ColorM,
		action:    action,
	}
	c.factory = newFactory(c, o.PP, o.PPUnit)
	return c
}

// The canvas has no useful image to work with.
func (c *canvas) Noop() bool {
	if c.action == ImageNoop {
		return true
	}
	return false
}

// Saves the canvas image to its path.
func (c *canvas) Save() error {
	if c.action != ImageNoop {
		return SaveImage(c.path, c.extension, c.Image)
	}
	return nil
}

// Extract the draw.Image of this canvas
func (c *canvas) GetImage() draw.Image {
	return c.Image
}

// Replace the image of this canvas with the supplied image
// Note: the only anchors that get changed are center, origin, and bound
// as of now you are responsible for updating your custom anchors if you
// change your image dimensions
func (c *canvas) SetImage(i draw.Image) {
	c.action = ImageOpen
	c.Image = i
	updateFactory(c, c.factory)
}
