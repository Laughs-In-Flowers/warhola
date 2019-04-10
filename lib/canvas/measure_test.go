package canvas

import (
	"image"
	"math"
	"testing"

	"golang.org/x/image/math/fixed"
)

func TestMoint(t *testing.T) {
	id := "Moint"
	a := Moint{0, 0}

	// Fixed
	f := fixed.Point26_6{0, 0}
	if a.Fixed() != f {
		failProbe(
			t,
			id,
			"Fixed",
			commonExpect,
			a, f,
		)
	}

	b := Moint{50, 100}

	// Distance
	expD := math.Hypot(a.X-b.X, a.Y-b.Y)
	if dis := a.Distance(b); dis != expD {
		failProbe(
			t,
			id,
			"Distance",
			commonExpect,
			dis, expD,
		)
	}

	// Interpolate
	intp := float64(rr.Intn(100))
	interpExp := Moint{
		a.X + (b.X-a.X)*intp,
		a.Y + (b.Y-a.Y)*intp,
	}
	if interpHave := a.Interpolate(b, intp); interpHave != interpExp {
		failProbe(
			t,
			id,
			"Interpolate",
			commonExpect,
			interpHave, interpExp,
		)
	}

	iz := image.Point{0, 0}

	// IPoint() image.Point {
	pointExp := iz
	if pointHave := a.IPoint(); pointHave != pointExp {
		failProbe(
			t,
			id,
			"IPoint",
			commonExpect,
			pointHave, pointExp,
		)
	}

	// ToMoint(ip image.Point) Moint {
	if toMointHave := ToMoint(iz); toMointHave != a {
		failProbe(
			t,
			"-",
			"ToMoint",
			commonExpect,
			toMointHave, a,
		)
	}

	// Add
	am := Moint{float64(rr.Intn(500)), float64(rr.Intn(500))}
	addExp := Moint{a.X + am.X, a.Y + am.Y}
	if addHave := a.Add(am); addHave != addExp {
		failProbe(
			t,
			id,
			"Add",
			commonExpect,
			addHave, addExp,
		)
	}

	// Div
	divV := float64(rr.Intn(10))
	divExp := Moint{b.X / divV, b.Y / divV}
	if divHave := b.Div(divV); divHave != divExp {
		failProbe(
			t,
			id,
			"Div",
			commonExpect,
			divHave, divExp,
		)
	}

	// Eq(o Moint) bool {
	eqM := Moint{0, 0}
	if !a.Eq(eqM) {
		failProbe(
			t,
			id,
			"Eq",
			commonExpect,
			eqM, a,
		)
	}

	irect := image.Rect(0, 0, 51, 101)

	// In
	rmX, rmY := float64(irect.Min.X), float64(irect.Min.Y)
	rxX, rxY := float64(irect.Max.X), float64(irect.Max.Y)
	isIn := rmX <= b.X && b.X < rxX &&
		rmY <= b.Y && b.Y < rxY
	if isIn != b.In(irect) {
		failProbe(
			t,
			id,
			"In",
			"%s is not in %s",
			b, irect,
		)
	}

	// Mod
	// warning label on this section; still a ?
	rMin := ToMoint(irect.Min)
	w, h := float64(irect.Dx()), float64(irect.Dy())
	p := b.Sub(rMin)
	p.X = math.Mod(p.X, w)
	if p.X < 0 {
		p.X += w
	}
	p.Y = math.Mod(p.Y, h)
	if p.Y < 0 {
		p.Y += h
	}
	modExp := p.Add(rMin)
	if modHave := b.Mod(irect); modHave != modExp {
		failProbe(
			t,
			id,
			"Mod",
			commonExpect,
			modHave, modExp,
		)
	}

	// Mul
	mulV := float64(rr.Intn(100) + 1)
	mulExp := Moint{b.X * mulV, b.Y * mulV}
	if mulHave := b.Mul(mulV); mulHave != mulExp {
		failProbe(
			t,
			id,
			"Mul",
			commonExpect,
			mulHave, mulExp,
		)
	}

	// String()
	stringExp := "(0.0000,0.0000)"
	if stringHave := a.String(); stringHave != stringExp {
		failProbe(
			t,
			id,
			"String",
			commonExpect,
			stringHave, stringExp,
		)
	}

	// Sub(o Moint) Moint {
	subMoint := Moint{float64(rr.Intn(10)), float64(rr.Intn(10))}
	subExp := Moint{b.X - subMoint.X, b.Y - subMoint.Y}
	if subHave := b.Sub(subMoint); subHave != subExp {
		failProbe(
			t,
			id,
			"Sub",
			commonExpect,
			subHave, subExp,
		)
	}
}
