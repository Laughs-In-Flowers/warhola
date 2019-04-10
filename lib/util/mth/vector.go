package mth

import "math"

type V2 struct {
	X, Y float64
}

func (v V2) Abs() V2 {
	return V2{math.Abs(v.X), math.Abs(v.Y)}
}

func (v V2) Len() float64 {
	return math.Sqrt(v.Len2())
}

func (v V2) Len2() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v V2) LenV() V2 {
	d := v.Len()
	return V2{d, d}
}

func (v V2) Scale(s float64) V2 {
	return V2{v.X * s, v.Y * s}
}

func (a V2) Add(b V2) V2 {
	return V2{a.X + b.X, a.Y + b.Y}
}

func (a V2) Offset(s float64) V2 {
	return V2{a.X + s, a.Y + s}
}

type V3 struct {
	X, Y, Z float64
}

func (v V3) Abs() V3 {
	return V3{math.Abs(v.X), math.Abs(v.Y), math.Abs(v.Z)}
}

func (v V3) Len2() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v V3) Len() float64 {
	return math.Sqrt(v.Len2())
}

func (v V3) LenV() V3 {
	d := v.Len()
	return V3{d, d, d}
}

func (v V3) Scale(s float64) V3 {
	return V3{v.X * s, v.Y * s, v.Z * s}
}

func (a V3) Add(b V3) V3 {
	return V3{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func (a V3) Offset(s float64) V3 {
	return V3{a.X + s, a.Y + s, a.Z + s}
}
