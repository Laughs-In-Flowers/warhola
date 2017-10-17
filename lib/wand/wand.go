package wand

import (
	"image/draw"
	"math"

	"github.com/fogleman/gg"
)

// A struct that embeds a *gg.Context as well as facilities for manipulating
// images in a point by point or drawing capacity.
type Wand struct {
	i                               draw.Image
	Height, Width, Radius, PPI, PPC float64
	Center, Origin, Bound           *Point
	Anchors                         []*Anchor
	*gg.Context
}

func initialize(w *Wand, ppi float64) {
	w.PPI = ppi
	w.PPC = ppi / 2.54
	d := w.i
	b := d.Bounds()
	c := gg.NewContextForImage(d)
	w.Context = c
	w.Width, w.Height = float64(b.Dx()), float64(b.Dy())
	w.Radius = w.Height / 2
	w.Origin = &Point{0, 0}
	w.Center = &Point{w.Width / 2, w.Height / 2}
	w.Bound = &Point{w.Width, w.Height}
	w.Anchors = make([]*Anchor, 0)
}

// Produces a new Wand instance with a specified PPI(points per inch) bound to
// the provided draw.Image.
func New(ppi float64, i draw.Image) *Wand {
	w := &Wand{i: i}
	initialize(w, ppi)
	return w
}

func (w *Wand) SetImage(i draw.Image) {
	w.i = i
}

func (w *Wand) GetImage() draw.Image {
	return w.i
}

func distance(p1, p2 *Point) float64 {
	return math.Sqrt(math.Pow((p2.X-p1.X), 2) + math.Pow((p2.Y-p1.Y), 2))
}

// Returns the float64 distance between 2 provided Points.
func (w *Wand) Distance(p1, p2 *Point) float64 {
	return distance(p1, p2)
}

// Returns the float64 distance between 2 provided Points in inches.
func (w *Wand) DistanceInch(p1, p2 *Point) float64 {
	d := distance(p1, p2)
	return d / w.PPI
}

// Returns the float64 distance between 2 provided Points in centimeters.
func (w *Wand) DistanceCM(p1, p2 *Point) float64 {
	d := distance(p1, p2)
	return d / w.PPC
}

// Returns the float64 distance between 2 provided Points in millimeters.
func (w *Wand) DistanceMM(p1, p2 *Point) float64 {
	d := distance(p1, p2)
	mm := w.PPC / 10
	return d / mm
}
