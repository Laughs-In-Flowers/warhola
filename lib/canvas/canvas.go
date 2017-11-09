package canvas

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/util"
)

type Pather interface {
	Path() string
	SetPath(string)
}

type pather struct {
	path string
}

const DefaultPath = "DEFAULT"

func (p *pather) Path() string {
	return p.path
}

func (p *pather) SetPath(as string) {
	p.path = as
}

func (p *pather) clone() *pather {
	np := *p
	return &np
}

type Getter interface {
	IGet() Image
}

type Setter interface {
	ISet(Image)
}

type GetterSetter interface {
	Getter
	Setter
}

type Copier interface {
	Copy() Image
	CopyTo(color.Model) Image
}

type Paster interface {
	Paste(Image, Point)
	Overlay(Image, Point, float64)
}

type Cloner interface {
	Clone() Canvas
}

// An interface for manipulating a single draw.Image.
type Canvas interface {
	Image
	Pather
	GetterSetter
	Copier
	Paster
	Cloner
	Factory
	Noop() bool
	Save() error
}

type canvas struct {
	Image
	kind   Kind
	action Action
	*pather
	*factory
}

type Kind int

const (
	NoKind Kind = iota
	PNG
	JPG
)

func StringToKind(s string) Kind {
	switch strings.ToLower(s) {
	case "png":
		return PNG
	case "jpeg", "jpg":
		return JPG
	}
	return NoKind
}

func (k Kind) String() string {
	switch k {
	case PNG:
		return "png"
	case JPG:
		return "jpg"
	}
	return "NoKind"
}

type options struct {
	Debug  bool
	ColorM color.Model
	Path   string
	Kind   Kind
	X, Y   int
	PP     int
	PPUnit string
}

// Provide options by function when creating a canvas
// (color model, path, extension, x, y, points per, points per unit).
func Options(debug bool, colorm, path, kind string, x, y, pp int, ppu string) options {
	return options{
		Debug:  debug,
		ColorM: util.StringToColorModel(colorm),
		Path:   path,
		Kind:   StringToKind(kind),
		X:      x,
		Y:      y,
		PP:     pp,
		PPUnit: ppu,
	}
}

type Action int

const (
	ImageNoop Action = iota
	ImageNew
	ImageOpen
	ImageClone
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
	var i Image
	var err error
	action := newAction(o)
	switch action {
	case ImageNoop:
		i = NewFrom(o.ColorM, 1, 1)
		l.Println("image is noop")
	case ImageNew:
		i = NewFrom(o.ColorM, o.X, o.Y)
		err = SaveImage(o.Path, o.Kind, i)
		if err != nil {
			l.Fatalln(err)
		}
		l.Printf("image is new: %s", o.Path)
	case ImageOpen:
		var ext string
		i, ext, err = OpenImage(o.Path, o.ColorM)
		o.Kind = StringToKind(ext)
		if err != nil {
			l.Fatalln(err)
		}
		l.Printf("image exists: %s", o.Path)
	}
	ni := Clone(i, o.ColorM)
	if o.Debug {
		sp := o.Path
		o.Path = fmt.Sprintf("%s-debug", sp)
		l.Printf("debug path is: %s", o.Path)
	}
	c := &canvas{
		Image:  ni,
		kind:   o.Kind,
		action: action,
		pather: &pather{o.Path},
	}
	c.factory = newFactory(c.Bounds(), o.PP, o.PPUnit)
	return c
}

// A boolean indicating this canvas has no useful image to work with.
func (c *canvas) Noop() bool {
	if c.action == ImageNoop {
		return true
	}
	return false
}

// Saves the canvas image to its path.
func (c *canvas) Save() error {
	if c.action != ImageNoop {
		return SaveImage(c.path, c.kind, c.Image)
	}
	return nil
}

// Extract the draw.Image of this canvas.
func (c *canvas) IGet() Image {
	return c.Image
}

// Replace the image of this canvas with the supplied image.
func (c *canvas) ISet(i Image) {
	c.action = ImageOpen
	c.Image = i
	updateFactory(c.factory, c.Bounds())
}

// Returns a copy of the canvas image.
func (c *canvas) Copy() Image {
	return Clone(c.Image, c.ColorModel())
}

// Returns a copy of the image with the specified color model.
func (c *canvas) CopyTo(m color.Model) Image {
	return Clone(c.Image, m)
}

// Pastes the provided image over the canvas image at the specified point.
func (c *canvas) Paste(i Image, p Point) {
	c.ISet(paste(i, c.IGet(), p))
}

// Pastes the provided image over the canvas image at the specified point with
// the specified opacity as a float indicating up to 100% opacity
// (i.e. a float from 1 to 100).
func (c *canvas) Overlay(i Image, p Point, opacity float64) {
	c.ISet(overlay(i, c.IGet(), p, opacity))
}

// Clone canvas to new and separate canvas.
// Note: will retain canvas path.
func (c *canvas) Clone() Canvas {
	nc := *c
	nc.Image = nil
	nc.pather = nil
	nc.factory = nil
	ni := c.Copy()
	nf := cloneFactory(c.factory, ni.Bounds())
	n := &nc
	n.Image = ni
	n.pather = c.pather.clone()
	n.factory = nf
	n.action = ImageClone
	return n
}
