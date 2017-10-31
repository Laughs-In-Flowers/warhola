package canvas

import (
	"math"
	"strings"

	"golang.org/x/image/math/fixed"
)

// A suite of tools for examining, manipulating, and measuring a canvas.
type Factory interface {
	Distance(Point, Point) float64
	DistanceInch(Point, Point) float64
	DistanceCM(Point, Point) float64
	DistanceMM(Point, Point) float64
	Anchor(string, Point)
	GetAnchor(string) Point
}

type factory struct {
	height, width, radius, ppi, ppc float64
	anchors                         []*Anchor
}

func newFactory(c *canvas, ppIn int, ppu string) *factory {
	f := &factory{}
	initializeFactory(f, c, ppIn, ppu)
	return f
}

func initializeFactory(f *factory, c *canvas, ppIn int, ppu string) {
	pp := float64(ppIn)
	switch strings.ToLower(ppu) {
	case "in", "inch":
		f.ppi = pp
		f.ppc = pp / 2.54
	case "cm", "centimeter":
		f.ppc = pp
		f.ppi = pp * 2.54
	}
	updateFactory(c, f)
}

func updateFactory(c *canvas, f *factory) {
	i := c.Image
	b := i.Bounds()
	f.width, f.height = float64(b.Dx()), float64(b.Dy())
	f.radius = f.height / 2
	if f.anchors == nil {
		f.anchors = make([]*Anchor, 0)
	}
	f.Anchor("origin", Point{0, 0})
	f.Anchor("center", Point{f.width / 2, f.height / 2})
	f.Anchor("bound", Point{f.width, f.height})
}

func distance(p1, p2 Point) float64 {
	return math.Sqrt(math.Pow((p2.X-p1.X), 2) + math.Pow((p2.Y-p1.Y), 2))
}

// Returns the float64 distance between 2 provided Points.
func (f *factory) Distance(p1, p2 Point) float64 {
	return distance(p1, p2)
}

// Returns the float64 distance between 2 provided Points in inches.
func (f *factory) DistanceInch(p1, p2 Point) float64 {
	d := distance(p1, p2)
	return d / f.ppi
}

// Returns the float64 distance between 2 provided Points in centimeters.
func (f *factory) DistanceCM(p1, p2 Point) float64 {
	d := distance(p1, p2)
	return d / f.ppc
}

// Returns the float64 distance between 2 provided Points in millimeters.
func (f *factory) DistanceMM(p1, p2 Point) float64 {
	d := distance(p1, p2)
	mm := f.ppc / 10
	return d / mm
}

// A struct encompassing a 2D float64 X - Y coordinate
type Point struct {
	X, Y float64
}

// fixed.Point26_6 representation of point
func (p Point) Fixed() fixed.Point26_6 {
	return fixp(p.X, p.Y)
}

// Distance from point to other point.
func (p Point) Distance(o Point) float64 {
	return math.Hypot(p.X-o.X, p.Y-o.Y)
}

// Interpolate other point to provided float64.
func (p Point) Interpolate(o Point, t float64) Point {
	x := p.X + (o.X-p.X)*t
	y := p.Y + (o.Y-p.Y)*t
	return Point{x, y}
}

// A struct embedding a Point and a string tag as formal name.
type Anchor struct {
	Tag string
	*Point
}

// Produces a new Anchor from the provided tag nad point.
func NewAnchor(tag string, point Point) *Anchor {
	return &Anchor{tag, &point}
}

// Sets an Anchor for use in a Tool instance.
func (f *factory) Anchor(tag string, point Point) {
	a := NewAnchor(tag, point)
	f.anchors = append(f.anchors, a)
}

// Gets the provided Anchor by tag from the tool instance, defaults to Point{0,0}
func (f *factory) GetAnchor(tag string) Point {
	for _, a := range f.anchors {
		if a.Tag == tag {
			return *a.Point
		}
	}
	return Point{0, 0}
}
