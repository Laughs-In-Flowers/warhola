package mth

import (
	"golang.org/x/image/math/fixed"
)

func Fixp(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{Fix(x), Fix(y)}
}

func Fix(x float64) fixed.Int26_6 {
	return fixed.Int26_6(x * 64)
}

func Clamp(value, min, max float64) float64 {
	if value > max {
		return max
	}
	if value < min {
		return min
	}
	return value
}
