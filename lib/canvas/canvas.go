package canvas

import (
	"image/color"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/xrr"
)

// Provided Config functions, returns a new Canvas and any configuration errors.
func New(cnf ...Config) (Canvas, error) {
	c := &canvas{
		identity: newIdentity(),
		pxl:      newPxl(),
	}
	cc := newConfiguration(c, cnf...)
	err := cc.Configure()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// The primary interface for all image manipulation needs.
type Canvas interface {
	log.Logger
	Identity
	Pxl
	Nooper
	Saver
	Cloner
	Operator
}

type canvas struct {
	log.Logger
	Configuration
	*identity
	*pxl
}

// An interface for denoting a non operational Canvas.
type Nooper interface {
	Noop() bool
}

// Operational or Nonoperational status of a Canvas. Returns a boolean where
// true is nonoperational and false is operational.
func (c *canvas) Noop() bool {
	switch {
	case c.pxl.m == COLORNOOP,
		c.path == PATHNOOP,
		c.fileType == FILETYPENOOP,
		c.action == ACTIONNOOP:
		return true
	default:
		return false
	}
	return true
}

// An interface for saving a canvas.
type Saver interface {
	Save() error
	SaveTo(string) error
}

var SaveNoopError = xrr.Xrror("cannot save a non operational canvas")

// Saves the canvas according to its current status.
func (c *canvas) Save() error {
	if !c.Noop() {
		c.Printf("canvas %s saving...", c.path)
		return save(c.path, c.fileType, c.pxl)
	}
	return SaveNoopError
}

// Save the canvas to its current status as the provided string color model.
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

// clone the canvas to the provided color.Model
func (c *canvas) CloneTo(m color.Model) Canvas {
	return cloneTo(c, m)
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

type canvasMutate func() (*pxl, error)

func (c *canvas) mutate(fn canvasMutate) error {
	np, err := fn()
	if err != nil {
		return err
	}
	c.pxl = np
	return nil
}
