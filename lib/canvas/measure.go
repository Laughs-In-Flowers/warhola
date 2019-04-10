package canvas

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"

	"github.com/Laughs-In-Flowers/warhola/lib/util/mth"
	"golang.org/x/image/math/fixed"
)

// An interface of tools for examining, manipulating, and measuring a rectangle.
type Measure interface {
	PP(string) float64
	SetPP(string, float64)
	Height() float64
	Width() float64
	Radius() float64
	Distance(string, Moint, Moint) float64
	Anchor(string) Moint
	SetAnchor(string, MointFunc)
}

type measure struct {
	*anchored
	r                 *image.Rectangle
	ppu               string
	pp, ppi, ppc, ppm float64
}

var (
	defaultPPU   string   = "cm"                               // points per unit
	defaultPP    float64  = 118                                // points per
	AvailablePPU []string = []string{defaultPPU, "mm", "inch"} //
)

func newMeasure(r *image.Rectangle, ppIn float64, ppu string) *measure {
	m := &measure{
		r:   r,
		ppu: ppu,
		pp:  ppIn,
	}
	m.anchored = newAnchored(m)
	m.SetPP(m.ppu, m.pp)
	return m
}

// Returns points per unit of measurement (inch,cm,mm,or nonspecific default)
func (m *measure) PP(pp string) float64 {
	switch pp {
	case "inch", "in":
		return m.ppi
	case "centimeter", "cm":
		return m.ppc
	case "millimeter", "mm":
		return m.ppm
	}
	return float64(m.pp)
}

func (m *measure) SetPP(ppu string, pp float64) {
	m.pp = pp
	switch strings.ToLower(ppu) {
	case "cm", "centimeter":
		m.ppc = m.pp
		m.ppm = m.pp / 10
		m.ppi = m.ppc * 2.54
	case "mm", "millimeter":
		m.ppm = pp
		m.ppc = m.ppm * 10
		m.ppi = m.ppc * 2.54
	case "in", "inch":
		m.ppi = m.pp
		m.ppc = m.pp / 2.54
		m.ppm = m.ppc / 10
	}
}

// Returns the height as a float64.
func (m *measure) Height() float64 {
	b := m.r.Bounds()
	return float64(b.Dy())
}

// Returns the width as a float64.
func (m *measure) Width() float64 {
	b := m.r.Bounds()
	return float64(b.Dx())
}

// Returns the radius as a float64.
func (m *measure) Radius() float64 {
	return m.Height() / 2
}

func distance(p1, p2 Moint) float64 {
	return math.Sqrt(math.Pow((p2.X-p1.X), 2) + math.Pow((p2.Y-p1.Y), 2))
}

// Returns the float64 distance between two provided Points for the provided unit.
func (m *measure) Distance(u string, p1, p2 Moint) float64 {
	switch u {
	case "inch":
		return dINCH(m, p1, p2)
	case "cm":
		return dCM(m, p1, p2)
	case "mm":
		return dMM(m, p1, p2)
	}
	return distance(p1, p2)
}

func dINCH(m *measure, p1, p2 Moint) float64 {
	d := distance(p1, p2)
	return d / m.ppi
}

func dCM(m *measure, p1, p2 Moint) float64 {
	d := distance(p1, p2)
	return d / m.ppc
}

func dMM(m *measure, p1, p2 Moint) float64 {
	d := distance(p1, p2)
	return d / m.ppm
}

// A function type that takes a Measure and returns a Point.
type MointFunc func(f Measure) Moint

// A struct encapsulating a specific point derived by relationship function.
type anchor struct {
	fn  MointFunc
	tag string
}

// Produces a new Anchor from the provided tag and point function.
func newAnchor(tag string, fn MointFunc) *anchor {
	return &anchor{fn, tag}
}

type anchored struct {
	m   *measure
	has []*anchor
}

func newAnchored(m *measure) *anchored {
	a := &anchored{
		m:   m,
		has: make([]*anchor, 0),
	}
	a.SetAnchor(
		"origin",
		func(Measure) Moint { return Moint{0, 0} },
	)
	a.SetAnchor(
		"center",
		func(mm Measure) Moint { return Moint{mm.Width() / 2, mm.Height() / 2} },
	)
	a.SetAnchor(
		"bound",
		func(mm Measure) Moint { return Moint{mm.Width(), mm.Height()} },
	)
	return a
}

func (a *anchored) getAnchor(tag string) *anchor {
	for _, aa := range a.has {
		if aa.tag == tag {
			return aa
		}
	}
	return nil
}

// Gets an anchor point corresponding to provided tag. No available anchor
// returns a zeroed Moint.
func (a *anchored) Anchor(tag string) Moint {
	if aa := a.getAnchor(tag); aa != nil {
		return aa.fn(a.m)
	}
	return ZM
}

// Sets an Anchor.
func (a *anchored) SetAnchor(tag string, fn MointFunc) {
	aa := newAnchor(tag, fn)
	a.has = append(a.has, aa)
}

// A struct encompassing a 2D float64 X - Y coordinate for measurement.
// Functionality is paralell to image.Point, but not an interface.
type Moint struct {
	X, Y float64
}

var ZM Moint = Moint{0, 0}

// Returns a fixed.Point26_6 representation of Point.
func (p Moint) Fixed() fixed.Point26_6 {
	return mth.Fixp(p.X, p.Y)
}

// Returns the distance from point to other point.
func (p Moint) Distance(o Moint) float64 {
	return math.Hypot(p.X-o.X, p.Y-o.Y)
}

// Interpolate other Point to provided float64.
func (p Moint) Interpolate(o Moint, t float64) Moint {
	x := p.X + (o.X-p.X)*t
	y := p.Y + (o.Y-p.Y)*t
	return Moint{x, y}
}

//
func (p Moint) IPoint() image.Point {
	return image.Point{
		int(math.Round(p.X)),
		int(math.Round(p.Y)),
	}
}

//
func ToMoint(ip image.Point) Moint {
	return Moint{float64(ip.X), float64(ip.Y)}
}

// Add other Point to this Point.
func (p Moint) Add(o Moint) Moint {
	return Moint{p.X + o.X, p.Y + o.Y}
}

//
func (p Moint) Div(k float64) Moint {
	return Moint{p.X / k, p.Y / k}
}

//
func (p Moint) Eq(o Moint) bool {
	return p.X == o.X && p.Y == o.Y
}

//
func (p Moint) In(r image.Rectangle) bool {
	rmX, rmY := float64(r.Min.X), float64(r.Min.Y)
	rxX, rxY := float64(r.Max.X), float64(r.Max.Y)
	return rmX <= p.X && p.X < rxX &&
		rmY <= p.Y && p.Y < rxY
}

//
func (p Moint) Mod(r image.Rectangle) Moint {
	rMin := ToMoint(r.Min)
	w, h := float64(r.Dx()), float64(r.Dy())
	mp := p.Sub(rMin)
	mp.X = math.Mod(mp.X, w)
	if mp.X < 0 {
		mp.X += w
	}
	mp.Y = math.Mod(mp.Y, h)
	if mp.Y < 0 {
		mp.Y += h
	}
	return mp.Add(rMin)
}

//
func (p Moint) Mul(k float64) Moint {
	return Moint{p.X * k, p.Y * k}
}

//
func (p Moint) String() string {
	x := strconv.FormatFloat(p.X, 'f', 4, 64)
	y := strconv.FormatFloat(p.Y, 'f', 4, 64)
	return fmt.Sprintf("(%s,%s)", x, y)
}

// Subtract other Moint from this Moint
func (p Moint) Sub(o Moint) Moint {
	return Moint{p.X - o.X, p.Y - o.Y}
}
