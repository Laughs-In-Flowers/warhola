package canvas

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/Laughs-In-Flowers/warhola/lib/star"
)

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

func New(path string, i image.Image, c color.Model) Canvas {
	_, tag, _ := splitPath(path)
	ni := Clone(i, c)
	SaveImage(path, ni)
	return &canvas{
		tag, path, c, ni,
	}
}

func Open(path string, c color.Model) (Canvas, error) {
	i, err := OpenImage(path, c)
	if err != nil {
		return nil, err
	}
	return New(path, i, c), nil
}

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

func (c *canvas) Path() string {
	return c.path
}

func (c *canvas) Apply(args *star.Args, rs ...star.Star) error {
	for _, st := range rs {
		err := st(args.Path, args.Debug, args.Args...)
		if err != nil {
			return err
		}
	}

	ni, err := OpenImage(c.path, c.model)
	if err != nil {
		return err
	}

	c.Image = ni

	return nil
}

func (c *canvas) Save() error {
	return SaveImage(c.path, c.Image)
}

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

func NewCanvaser() Canvaser {
	return &canvaser{New, Open, OpenAll}
}

func (c *canvaser) NewCanvas(p string, i image.Image, m color.Model) Canvas {
	return c.nfn(p, i, m)
}

func (c *canvaser) OpenCanvas(p string, m color.Model) (Canvas, error) {
	return c.ofn(p, m)
}

func (c *canvaser) OpenCanvases(p []string, m color.Model) ([]Canvas, error) {
	return c.afn(p, m)
}

var DefaultCanvaser Canvaser

func init() {
	DefaultCanvaser = NewCanvaser()
}
