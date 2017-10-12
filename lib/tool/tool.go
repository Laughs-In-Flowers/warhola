package tool

import (
	"image/draw"
	"math"

	"github.com/fogleman/gg"
)

// A struct that embeds a *gg.Context as well as facilities for manipulating
// images in a point by point or drawing capacity.
type Tool struct {
	i                               draw.Image
	Height, Width, Radius, PPI, PPC float64
	Center, Origin, Bound           *Point
	Anchors                         []*Anchor
	*gg.Context
}

func initialize(t *Tool, ppi float64) {
	t.PPI = ppi
	t.PPC = ppi / 2.54
	d := t.i
	b := d.Bounds()
	c := gg.NewContextForImage(d)
	t.Context = c
	t.Width, t.Height = float64(b.Dx()), float64(b.Dy())
	t.Radius = t.Height / 2
	t.Origin = &Point{0, 0}
	t.Center = &Point{t.Width / 2, t.Height / 2}
	t.Bound = &Point{t.Width, t.Height}
	t.Anchors = make([]*Anchor, 0)
}

// Produces a new Tool instance with a specified PPI(points per inch) bound to
// the provided draw.Image.
func New(ppi float64, i draw.Image) *Tool {
	t := &Tool{i: i}
	initialize(t, ppi)
	return t
}

func distance(p1, p2 *Point) float64 {
	return math.Sqrt(math.Pow((p2.X-p1.X), 2) + math.Pow((p2.Y-p1.Y), 2))
}

// Returns the float64 distance between 2 provided Points.
func (t *Tool) Distance(p1, p2 *Point) float64 {
	return distance(p1, p2)
}

// Returns the float64 distance between 2 provided Points in inches.
func (t *Tool) DistanceInch(p1, p2 *Point) float64 {
	d := distance(p1, p2)
	return d / t.PPI
}

// Returns the float64 distance between 2 provided Points in centimeters.
func (t *Tool) DistanceCM(p1, p2 *Point) float64 {
	d := distance(p1, p2)
	return d / t.PPC
}

// Returns the float64 distance between 2 provided Points in millimeters.
func (t *Tool) DistanceMM(p1, p2 *Point) float64 {
	d := distance(p1, p2)
	mm := t.PPC / 10
	return d / mm
}
