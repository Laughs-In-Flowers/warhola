package canvas

import (
	"image"
	"math"
	"strings"

	"github.com/Laughs-In-Flowers/warhola/lib/util"
	"golang.org/x/image/math/fixed"
)

// A suite of tools for examining, manipulating, and measuring a canvas.
type Factory interface {
	Height() float64
	Width() float64
	Radius() float64
	Distance(Point, Point) float64
	DistanceInch(Point, Point) float64
	DistanceCM(Point, Point) float64
	DistanceMM(Point, Point) float64
	Anchor(string, PointFunc)
	GetAnchor(string) Point
}

type factory struct {
	r        image.Rectangle
	ppi, ppc float64
	anchors  []*Anchor
}

func newFactory(r image.Rectangle, ppIn int, ppu string) *factory {
	f := &factory{}
	initializeFactory(f, r, ppIn, ppu)
	return f
}

func initializeFactory(f *factory, r image.Rectangle, ppIn int, ppu string) {
	pp := float64(ppIn)
	switch strings.ToLower(ppu) {
	case "in", "inch":
		f.ppi = pp
		f.ppc = pp / 2.54
	case "cm", "centimeter":
		f.ppc = pp
		f.ppi = pp * 2.54
	}
	updateFactory(f, r)
}

func updateFactory(f *factory, r image.Rectangle) {
	f.r = r
	if f.anchors == nil {
		f.anchors = make([]*Anchor, 0)
		f.Anchor(
			"origin",
			func(Factory) Point { return Point{0, 0} },
		)
		f.Anchor(
			"center",
			func(f Factory) Point { return Point{f.Width() / 2, f.Height() / 2} },
		)
		f.Anchor(
			"bound",
			func(f Factory) Point { return Point{f.Width(), f.Height()} },
		)
	}
}

func cloneFactory(f *factory, r image.Rectangle) *factory {
	of := *f
	nf := &of
	updateFactory(nf, r)
	return nf
}

// Returns the height of the underlying image.Rectangle as a float64.
func (f *factory) Height() float64 {
	b := f.r.Bounds()
	return float64(b.Dy())
}

// Returns the width of the underlying image.Rectangle as a float64.
func (f *factory) Width() float64 {
	b := f.r.Bounds()
	return float64(b.Dx())
}

// Returns the radius of the underlying image.Rectangle as a float64.
func (f *factory) Radius() float64 {
	return f.Height() / 2
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

// A struct encompassing a 2D float64 X - Y coordinate.
type Point struct {
	X, Y float64
}

// A Point from the provided x,y float64 numbers.
func Pt(x, y float64) Point {
	return Point{float64(x), float64(y)}
}

// Returns a fixed.Point26_6 representation of Point.
func (p Point) Fixed() fixed.Point26_6 {
	return util.Fixp(p.X, p.Y)
}

// Distance from point to other point.
func (p Point) Distance(o Point) float64 {
	return math.Hypot(p.X-o.X, p.Y-o.Y)
}

// Interpolate other Point to provided float64.
func (p Point) Interpolate(o Point, t float64) Point {
	x := p.X + (o.X-p.X)*t
	y := p.Y + (o.Y-p.Y)*t
	return Point{x, y}
}

// Returns an image.Point.
func (p Point) IPoint() image.Point {
	return image.Point{int(p.X), int(p.Y)}
}

// Add other Point to this Point.
func (p Point) Add(o Point) Point {
	return Point{p.X + o.X, p.Y + o.Y}
}

// Subtract other Point from this Point
func (p Point) Sub(o Point) Point {
	return Point{p.X - o.X, p.Y - o.Y}
}

// A function that takes a Factory and returns a Point.
type PointFunc func(f Factory) Point

// A struct encapsulating a specific point derived by relationship function.
type Anchor struct {
	fn  PointFunc
	Tag string
}

// Provided a Factory, return a Point
func (a *Anchor) Point(f Factory) Point {
	return a.fn(f)
}

// Produces a new Anchor from the provided tag and point function.
func NewAnchor(tag string, pf PointFunc) *Anchor {
	return &Anchor{pf, tag}
}

// Sets an Anchor.
func (f *factory) Anchor(tag string, pf PointFunc) {
	a := NewAnchor(tag, pf)
	f.anchors = append(f.anchors, a)
}

// Gets the provided Anchor by tag from the tool instance, defaults to Point{0,0}
func (f *factory) GetAnchor(tag string) Point {
	for _, a := range f.anchors {
		if a.Tag == tag {
			return a.fn(f)
		}
	}
	return Point{0, 0}
}

// image.Rect from two Points
func Rect(p1, p2 Point) image.Rectangle {
	return image.Rect(int(p1.X), int(p1.Y), int(p2.X), int(p2.Y))
}
