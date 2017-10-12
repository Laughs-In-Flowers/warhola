package tool

import (
	"math"

	"golang.org/x/image/math/fixed"
)

// A struct encompassing a 2D float64 X - Y coordinate
type Point struct {
	X, Y float64
}

func (p *Point) Fixed() fixed.Point26_6 {
	return fixp(p.X, p.Y)
}

func (p *Point) Distance(o *Point) float64 {
	return math.Hypot(p.X-o.X, p.Y-o.Y)
}

func (p *Point) Interpolate(o *Point, t float64) *Point {
	x := p.X + (o.X-p.X)*t
	y := p.Y + (o.Y-p.Y)*t
	return &Point{x, y}
}

// A struct embedding a Point and a string tag as formal name.
type Anchor struct {
	Tag string
	*Point
}

// Produces a new Anchor from the provided tag nad point.
func NewAnchor(tag string, point *Point) *Anchor {
	return &Anchor{tag, point}
}

// Sets an Anchor for use in a Tool instance.
func (t *Tool) SetAnchor(tag string, point *Point) {
	a := NewAnchor(tag, point)
	t.Anchors = append(t.Anchors, a)
}

// Gets the provided Anchor by tag from the tool instance.
func (t *Tool) GetAnchor(tag string) *Point {
	for _, a := range t.Anchors {
		if a.Tag == tag {
			return a.Point
		}
	}
	return nil
}
