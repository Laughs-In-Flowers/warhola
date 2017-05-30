package canvas

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/Laughs-In-Flowers/warhola/lib/star"
)

// An interface encompassing facilities for manipulating a single draw.Image.
type Canvas interface {
	draw.Image
	Path() string
	Apply(*star.Args, ...star.Star) error
	Save() error
}

type canvas struct {
	tag, path string
	model     color.Model
	draw.Image
}

// Given a string path, image.Image, and a color.Model, produces a Canvas interface.
func New(path string, i image.Image, c color.Model) Canvas {
	_, tag, _ := splitPath(path)
	ni := Clone(i, c)
	SaveImage(path, ni)
	return &canvas{
		tag, path, c, ni,
	}
}

// Opens a Canvas from the provided string path and color.Model.
func Open(path string, c color.Model) (Canvas, error) {
	i, err := OpenImage(path, c)
	if err != nil {
		return nil, err
	}
	return New(path, i, c), nil
}

// Opens an array of Canvases from the provided string path and color.Model.
func OpenAll(paths []string, m color.Model) ([]Canvas, error) {
	var ret []Canvas
	for _, p := range paths {
		c, err := Open(p, m)
		if err != nil {
			return nil, err
		}
		ret = append(ret, c)
	}
	return ret, nil
}

// The string path of the canvas.
func (c *canvas) Path() string {
	return c.path
}

// Applies the provided star.Stars to the canvas with the provided star.Args.
func (c *canvas) Apply(args *star.Args, rs ...star.Star) error {
	var i draw.Image = c.Image
	var err error

	for _, st := range rs {
		i, err = st(i, args.Debug, args.Args...)
		if err != nil {
			return err
		}
	}

	c.Image = i

	return nil
}

// Saves the canvas to its path.
func (c *canvas) Save() error {
	return SaveImage(c.path, c.Image)
}

// An interface for Canvas i/o manipulation: New, Open, & Open many.
type Canvaser interface {
	NewCanvas(string, image.Image, color.Model) Canvas
	OpenCanvas(string, color.Model) (Canvas, error)
	OpenCanvases([]string, color.Model) ([]Canvas, error)
}

type canvaser struct {
	nfn func(string, image.Image, color.Model) Canvas
	ofn func(string, color.Model) (Canvas, error)
	afn func([]string, color.Model) ([]Canvas, error)
}

func defaultCanvaser() Canvaser {
	return &canvaser{New, Open, OpenAll}
}

// Creates new canvas from the provided path, image.Image and color.Model.
func (c *canvaser) NewCanvas(p string, i image.Image, m color.Model) Canvas {
	return c.nfn(p, i, m)
}

// Opens a canvas from the provided path with the provided color.Model.
func (c *canvaser) OpenCanvas(p string, m color.Model) (Canvas, error) {
	return c.ofn(p, m)
}

// Opens any number of canvases from the provided string lsit with the provided color.Model.
func (c *canvaser) OpenCanvases(p []string, m color.Model) ([]Canvas, error) {
	return c.afn(p, m)
}

// A default Canvaser.
var DefaultCanvaser Canvaser

func init() {
	DefaultCanvaser = defaultCanvaser()
}
