package canvas

import (
	"image/color"
	"math"
	"strings"
)

type ColorModel int

const (
	COLORNOOP ColorModel = iota
	ALPHA
	ALPHA16
	CMYK
	GRAY
	GRAY16
	NRGBA
	NRGBA64
	RGBA
	RGBA64
)

func (c ColorModel) String() string {
	switch c {
	case ALPHA:
		return "ALPHA"
	case ALPHA16:
		return "ALPHA16"
	case CMYK:
		return "CMYK"
	case GRAY:
		return "GRAY"
	case GRAY16:
		return "GRAY16"
	case NRGBA:
		return "NRGBA"
	case NRGBA64:
		return "NRGBA64"
	case RGBA:
		return "RGBA"
	case RGBA64:
		return "RGBA64"
	}
	return "COLORNOOP"
}

func stringToColorModel(s string) ColorModel {
	switch strings.ToUpper(s) {
	case "ALPHA":
		return ALPHA
	case "ALPHA16":
		return ALPHA16
	case "CMYK":
		return CMYK
	case "GRAY":
		return GRAY
	case "GRAY16":
		return GRAY16
	case "NRGBA":
		return NRGBA
	case "NRGBA64":
		return NRGBA64
	case "RGBA":
		return RGBA
	case "RGBA64":
		return RGBA64
	}
	return COLORNOOP
}

func (c ColorModel) toColorModel() color.Model {
	switch c {
	case ALPHA:
		return color.AlphaModel
	case ALPHA16:
		return color.Alpha16Model
	case CMYK:
		return color.CMYKModel
	case GRAY:
		return color.GrayModel
	case GRAY16:
		return color.Gray16Model
	case NRGBA:
		return color.NRGBAModel
	case NRGBA64:
		return color.NRGBA64Model
	case RGBA:
		return color.RGBAModel
	case RGBA64:
		return color.RGBA64Model
	}
	return nil
}

func icmTocm(m color.Model, p *pxl) {
	switch m {
	case color.AlphaModel:
		p.m = ALPHA
		return
	case color.Alpha16Model:
		p.m = ALPHA16
		return
	case color.CMYKModel:
		p.m = CMYK
		return
	case color.GrayModel:
		p.m = GRAY
		return
	case color.Gray16Model:
		p.m = GRAY16
		return
	case color.NRGBAModel:
		p.m = NRGBA
		return
	case color.NRGBA64Model:
		p.m = NRGBA64
		return
	case color.RGBAModel:
		p.m = RGBA
		return
	case color.RGBA64Model:
		p.m = RGBA64
		return
	}

	p.m = COLORNOOP
}

var WorkingColorModel = "RGBA"

var AvailableColorModels = []ColorModel{
	ALPHA,
	ALPHA16,
	CMYK,
	GRAY,
	GRAY16,
	NRGBA,
	NRGBA64,
	RGBA,
	RGBA64,
}

type Paletter interface {
	Palettize(color.Color) color.Color
}

type PaletteFunc func(color.Color) color.Color

func mkUniformPaletteFunc(l color.Color) PaletteFunc {
	return func(color.Color) color.Color {
		return l
	}
}

func mkPaletteFunc(p color.Palette) PaletteFunc {
	return func(l color.Color) color.Color {
		return p.Convert(l)
	}
}

type rgba164 struct {
	R, G, B, A float64
}

func newRGBA164(r, g, b, a uint8) rgba164 {
	return rgba164{float64(r) / 255, float64(g) / 255, float64(b) / 255, float64(a) / 255}
}

func (c *rgba164) clamp() {
	c.R = clamp(c.R, 0, 1)
	c.G = clamp(c.G, 0, 1)
	c.B = clamp(c.B, 0, 1)
	c.A = clamp(c.A, 0, 1)
}

func alphaComp(bg, fg rgba164) rgba164 {
	fg.clamp()
	fga := fg.A

	r := (fg.R * fga / 1) + ((1 - fga) * bg.R / 1)
	g := (fg.G * fga / 1) + ((1 - fga) * bg.G / 1)
	b := (fg.B * fga / 1) + ((1 - fga) * bg.B / 1)
	a := bg.A + fga

	return rgba164{R: r, G: g, B: b, A: a}
}

// RGBToHSL converts from  RGB to HSL color model.
// Parameter c is the RGBA color and must implement the color.RGBA interface.
// Returned values h, s and l correspond to the hue, saturation and lightness.
// The hue is of range 0 to 360 and the saturation and lightness are of range 0.0 to 1.0.
func RGBToHSL(c color.RGBA) (float64, float64, float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	var h, s, l float64
	l = (max + min) / 2

	// Achromatic
	if delta <= 0 {
		return h, s, l
	}

	// Should it be smaller than or equals instead?
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2 - max - min)
	}

	if r >= max {
		h = (g - b) / delta
	} else if g >= max {
		h = (b-r)/delta + 2
	} else {
		h = (r-g)/delta + 4
	}

	h *= 60
	if h < 0 {
		h += 360
	}

	return h, s, l
}

// HSLToRGB converts from HSL to RGB color model.
// Parameter h is the hue and its range is from 0 to 360 degrees.
// Parameter s is the saturation and its range is from 0.0 to 1.0.
// Parameter l is the lightness and its range is from 0.0 to 1.0.
func HSLToRGB(h, s, l float64) color.RGBA {
	var r, g, b float64
	if s == 0 {
		r = l
		g = l
		b = l
	} else {
		var temp0, temp1 float64
		if l < 0.5 {
			temp0 = l * (1 + s)
		} else {
			temp0 = (l + s) - (s * l)
		}
		temp1 = 2*l - temp0

		h /= 360

		hueFn := func(v float64) float64 {
			if v < 0 {
				v++
			} else if v > 1 {
				v--
			}

			if v < 1.0/6.0 {
				return temp1 + (temp0-temp1)*6*v
			}
			if v < 1.0/2.0 {
				return temp0
			}
			if v < 2.0/3.0 {
				return temp1 + (temp0-temp1)*(2.0/3.0-v)*6
			}
			return temp1
		}

		r = hueFn(h + 1.0/3.0)
		g = hueFn(h)
		b = hueFn(h - 1.0/3.0)

	}

	outR := uint8(clamp(r*255+0.5, 0, 255))
	outG := uint8(clamp(g*255+0.5, 0, 255))
	outB := uint8(clamp(b*255+0.5, 0, 255))

	return color.RGBA{outR, outG, outB, 0xFF}
}

// RGBToHSV converts from  RGB to HSV color model.
// Parameter c is the RGBA color and must implement the color.RGBA interface.
// Returned values h, s and v correspond to the hue, saturation and value.
// The hue is of range 0 to 360 and the saturation and value are of range 0.0 to 1.0.
func RGBToHSV(c color.RGBA) (h, s, v float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	v = max
	delta := max - min

	// Avoid division by zero
	if max > 0 {
		s = delta / max
	} else {
		h = 0
		s = 0
		return
	}

	// Achromatic
	if max == min {
		h = 0
		return
	}

	if r >= max {
		h = (g - b) / delta
	} else if g >= max {
		h = (b-r)/delta + 2
	} else {
		h = (r-g)/delta + 4
	}

	h *= 60
	if h < 0 {
		h += 360
	}

	return
}

// HSVToRGB converts from HSV to RGB color model.
// Parameter h is the hue and its range is from 0 to 360 degrees.
// Parameter s is the saturation and its range is from 0.0 to 1.0.
// Parameter v is the value and its range is from 0.0 to 1.0.
func HSVToRGB(h, s, v float64) color.RGBA {
	var i, f, p, q, t float64

	// Achromatic
	if s == 0 {
		outV := uint8(clamp(v*255+0.5, 0, 255))
		return color.RGBA{outV, outV, outV, 0xFF}
	}

	h /= 60
	i = math.Floor(h)
	f = h - i
	p = v * (1 - s)
	q = v * (1 - s*f)
	t = v * (1 - s*(1-f))

	var r, g, b float64
	switch i {
	case 0:
		r = v
		g = t
		b = p
	case 1:
		r = q
		g = v
		b = p
	case 2:
		r = p
		g = v
		b = t
	case 3:
		r = p
		g = q
		b = v
	case 4:
		r = t
		g = p
		b = v
	default:
		r = v
		g = p
		b = q
	}

	outR := uint8(clamp(r*255+0.5, 0, 255))
	outG := uint8(clamp(g*255+0.5, 0, 255))
	outB := uint8(clamp(b*255+0.5, 0, 255))
	return color.RGBA{outR, outG, outB, 0xFF}
}
