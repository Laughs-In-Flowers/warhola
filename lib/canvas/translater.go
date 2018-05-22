package canvas

import (
	"image"
	"strings"
)

type Translater interface {
	Rotate(float64, bool, image.Point) error
	RotateTo(float64, bool, image.Point) (Canvas, error)
	Flip(string) error
	FlipTo(string) (Canvas, error)
	Shear(string, float64) error
	ShearTo(string, float64) (Canvas, error)
	Translate(int, int) error
	TranslateTo(int, int) (Canvas, error)
}

type tDir int

func stringToTDir(s string) tDir {
	switch strings.ToLower(s) {
	case "horizontal":
		return horizontal
	case "vertical":
		return vertical
	}
	return noDir
}

const (
	noDir tDir = iota
	horizontal
	vertical
)

func (c *canvas) Rotate(angle float64, preserve bool, at image.Point) error {
	return nil
}

func (c *canvas) RotateTo(angle float64, preserve bool, at image.Point) (Canvas, error) {
	nc := c.Clone()
	err := nc.Rotate(angle, preserve, at)
	return nc, err
}

func (c *canvas) Flip(dir string) error {
	return nil
}

func (c *canvas) FlipTo(dir string) (Canvas, error) {
	nc := c.Clone()
	err := nc.Flip(dir)
	return nc, err
}

func (c *canvas) Shear(dir string, angle float64) error {
	return nil
}

func (c *canvas) ShearTo(dir string, angle float64) (Canvas, error) {
	nc := c.Clone()
	err := nc.Shear(dir, angle)
	return nc, err
}

func (c *canvas) Translate(dx, dy int) error {
	return nil
}

func (c *canvas) TranslateTo(dx, dy int) (Canvas, error) {
	nc := c.Clone()
	err := nc.Translate(dx, dy)
	return nc, err
}
