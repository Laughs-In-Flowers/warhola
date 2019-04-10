package mth

import (
	"fmt"
	"math"
)

type Matrix interface {
	At(x, y int) float64
	Normalized() Matrix
	MaxX() int
	MaxY() int
}

//func StringToMatrix() *M {
//	return NewMatrix(0, 0)
//}

type M struct {
	MX   []float64
	W, H int
}

func NewMatrix(width, height int) *M {
	return &M{make([]float64, width*height), width, height}
}

func (m *M) Normalized() Matrix {
	sum := m.Absum()
	w := m.W
	h := m.H
	nm := NewMatrix(w, h)

	// avoid division by 0
	if sum == 0 {
		sum = 1
	}

	for i := 0; i < w*h; i++ {
		nm.MX[i] = m.MX[i] / sum
	}

	return nm
}

func (m *M) MaxX() int {
	return m.W
}

func (m *M) MaxY() int {
	return m.H
}

func (m *M) At(x, y int) float64 {
	return m.MX[y*m.W+x]
}

func (m *M) String() string {
	result := ""
	stride := m.MaxX()
	height := m.MaxY()
	for y := 0; y < height; y++ {
		result += fmt.Sprintf("\n")
		for x := 0; x < stride; x++ {
			result += fmt.Sprintf("%-8.4f", m.At(x, y))
		}
	}
	return result
}

func (m *M) Absum() float64 {
	var sum float64
	for _, v := range m.MX {
		sum += math.Abs(v)
	}
	return sum
}
