package main

import (
	"image/color"
	"math"
)

type V2 struct{ X, Y float64 }

type V3 struct{ X, Y, Z float64 }

func (v V2) Abs() V2 { return V2{math.Abs(v.X), math.Abs(v.Y)} }
func (v V3) Abs() V3 { return V3{math.Abs(v.X), math.Abs(v.Y), math.Abs(v.Z)} }

func (v V2) Len2() float64 { return v.X*v.X + v.Y*v.Y }
func (v V3) Len2() float64 { return v.X*v.X + v.Y*v.Y + v.Z*v.Z }

func (v V2) Len() float64 { return math.Sqrt(v.Len2()) }
func (v V3) Len() float64 { return math.Sqrt(v.Len2()) }

func (v V2) LenV() V2 {
	d := v.Len()
	return V2{d, d}
}
func (v V3) LenV() V3 {
	d := v.Len()
	return V3{d, d, d}
}

func (v V2) Scale(s float64) V2 { return V2{v.X * s, v.Y * s} }
func (v V3) Scale(s float64) V3 { return V3{v.X * s, v.Y * s, v.Z * s} }

func (a V2) Add(b V2) V2 { return V2{a.X + b.X, a.Y + b.Y} }
func (a V3) Add(b V3) V3 { return V3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }

func (a V2) Offset(s float64) V2 { return V2{a.X + s, a.Y + s} }
func (a V3) Offset(s float64) V3 { return V3{a.X + s, a.Y + s, a.Z + s} }

type M2 [4]float64

func (m M2) Mul2x1(v V2) V2 {
	return V2{
		m[0]*v.X + m[2]*v.Y,
		m[1]*v.X + m[3]*v.Y,
	}
}

func C3(R, G, B float64) Color { return Color{R, G, B, 1} }

type Color struct{ R, G, B, A float64 }

func (c Color) RGBA64() color.RGBA64 {
	return color.RGBA64{u16(c.R), u16(c.G), u16(c.B), u16(c.A)}
}

func u16(v float64) uint16 {
	x := int(v * 65536)
	if x >= 65536 {
		return 65535
	} else if x < 0 {
		return 0
	}
	return uint16(x)
}
