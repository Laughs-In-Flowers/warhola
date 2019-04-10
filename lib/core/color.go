package core

import (
	"fmt"
	"image/color"
)

//
type ColorRGB struct {
	R, G, B float64
}

//
func (c ColorRGB) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R*65535.0 + 0.5)
	g = uint32(c.G*65535.0 + 0.5)
	b = uint32(c.B*65535.0 + 0.5)
	a = 0xFFFF
	return
}

//type ColorRGBAF64 struct {
//	R, G, B, A float64
//}

//func (c ColorRGBAF64) Clamp() {
//	c.R = clamp(c.R, 0, 1)
//	c.G = clamp(c.G, 0, 1)
//	c.B = clamp(c.B, 0, 1)
//	c.A = clamp(c.A, 0, 1)
//}

//func (c ColorRGBAF64) RGBA() (r, g, b, a uint32) {
//	r = uint32(c.R*65535.0 + 0.5)
//	g = uint32(c.G*65535.0 + 0.5)
//	b = uint32(c.B*65535.0 + 0.5)
//	a = uint32(c.A*65535.0 + 0.5)
//	return
//}

// Hex parses a hex color-string, either in the 3 "#f0c", 3 "f0c",
// 6 "ff1034", or the default 6 "#ff1034" digits form.
func hex(scol string) (color.Color, error) {
	var format string
	var factor float64
	var f3 string = "#%1x%1x%1x"
	var fa3 = 1.0 / 15.0
	var f6 string = "#%02x%02x%02x"
	var fa6 = 1.0 / 255.0
	switch len(scol) {
	case 3:
		scol = fmt.Sprintf("#%s", scol)
		format = f3
		factor = fa3
	case 4:
		format = f3
		factor = fa3
	case 6:
		scol = fmt.Sprintf("#%s", scol)
		format = f6
		factor = fa6
	default:
		format = f6
		factor = fa6
	}

	var r, g, b uint8
	n, err := fmt.Sscanf(scol, format, &r, &g, &b)
	if err != nil {
		return ColorRGB{}, err
	}
	if n != 3 {
		return ColorRGB{}, fmt.Errorf("color: %v is not a hex-color", scol)
	}

	return ColorRGB{float64(r) * factor, float64(g) * factor, float64(b) * factor}, nil
}

func ToColor(model, value string) color.Color {
	var col color.Color
	var err error
	switch model {
	case "hex":
		col, err = hex(value)
	default:
		col = color.White
	}
	if err != nil {
		return color.White
	}
	return col
}
