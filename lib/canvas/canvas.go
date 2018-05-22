package canvas

import (
	"image/color"

	"github.com/Laughs-In-Flowers/log"
)

//
func New(cnf ...Config) (Canvas, error) {
	c := &canvas{
		identity: newIdentity(),
		pxl:      newPxl(),
	}
	cc := newConfiguration(c, cnf...)
	c.Configuration = cc
	err := c.Configure()
	if err != nil {
		return nil, err
	}
	return c, nil
}

type Canvas interface {
	log.Logger
	Configuration
	Identity
	Pxl
	Nooper
	Saver
	Cloner
	Operator
}

type Operator interface {
	Adjuster
	Blender
	//Blurer
	//Convoluter
	//Histogramer
	//Noiser
	//Painter
	//Segmenter
	Transformer
	Translater
}

type canvas struct {
	log.Logger
	Configuration
	*identity
	*pxl
	debug bool
}

// An interface for denoting a non operational Canvas.
type Nooper interface {
	Noop() bool
}

// returned boolean indicates generally if canvas is operational
func (c *canvas) Noop() bool {
	switch {
	case c.pxl.m == COLORNOOP,
		c.path == PATHNOOP,
		c.kind == KINDNOOP,
		c.action == ACTIONNOOP:
		return true
	default:
		return false
	}
	return true
}

//
type Saver interface {
	Save() error
	SaveTo(string) error
}

// save the canvas
func (c *canvas) Save() error {
	c.Printf("canvas %s saving...", c.path)
	return saveImage(c.path, c.kind, c.pxl)
}

// save the canvas to a specific color model
func (c *canvas) SaveTo(cm string) error {
	nm := stringToColorModel(cm).toColorModel()
	nc := cloneTo(c, nm)
	c.Printf("canvas switched to color model %s", cm)
	return nc.Save()
}

// An interface for cloning a Canvas.
type Cloner interface {
	Clone() Canvas
	CloneTo(color.Model) Canvas
}

// clone the canvas
func (c *canvas) Clone() Canvas {
	return c.CloneTo(c.pxl.ColorModel())
}

func cloneTo(c *canvas, m color.Model) *canvas {
	nc := &canvas{
		Logger:        c.Logger,
		Configuration: c.Configuration,
		identity:      c.identity.clone(),
		pxl:           c.pxl.clone(m),
	}
	return nc
}

// clone the canvas to the provided color.Model
func (c *canvas) CloneTo(m color.Model) Canvas {
	return cloneTo(c, m)
}
