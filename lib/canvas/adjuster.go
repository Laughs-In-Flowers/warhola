package canvas

import (
	"image/color"
	"math"
	"strings"

	"github.com/Laughs-In-Flowers/warhola/lib/util/xrr"
)

type Adjuster interface {
	Adjust(string, float64) error
	AdjustTo(string, float64) (Canvas, error)
}

type adjust int

func stringToAdjust(s string) adjust {
	switch strings.ToLower(s) {
	case "brightness":
		return brightness
	case "gamma":
		return gamma
	case "contrast":
		return contrast
	case "saturation":
		return saturation
	case "hue":
		return hue
	}
	return noAdjust
}

const (
	noAdjust adjust = iota
	brightness
	gamma
	contrast
	saturation
	hue
)

var NoAdjustError = xrr.Xrror("No adjustment corresponding to %s, %F").Out

func (c *canvas) Adjust(s string, v float64) error {
	switch stringToAdjust(s) {
	case brightness:
		return brightAdj(c, v)
	case gamma:
		return gammaAdj(c, v)
	case contrast:
		return contrastAdj(c, v)
	case saturation:
		return saturationAdj(c, v)
	case hue:
		return hueAdj(c, v)
	}
	return NoAdjustError(s, v)
}

func (c *canvas) AdjustTo(s string, v float64) (Canvas, error) {
	nc := c.Clone()
	err := nc.Adjust(s, v)
	return nc, err
}

type adjustmentFunc func(color.RGBA) color.RGBA

func adjustment(src *canvas, fn adjustmentFunc) error {
	srcP := src.pxl
	bounds := srcP.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dstP := scratch(srcP, color.RGBAModel, w, h)
	parallel(h, func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < w; x++ {
				srcPos := y*srcP.Stride() + x*4

				c := color.RGBA{}

				c.R = srcP.pix[srcPos+0]
				c.G = srcP.pix[srcPos+1]
				c.B = srcP.pix[srcPos+2]
				c.A = srcP.pix[srcPos+3]

				c = fn(c)

				dstP.pix[srcPos+0] = c.R
				dstP.pix[srcPos+1] = c.G
				dstP.pix[srcPos+2] = c.B
				dstP.pix[srcPos+3] = c.A
			}
		}
	})
	src.pxl = dstP.clone(src.ColorModel())
	return nil
}

func brightAdj(c *canvas, change float64) error {
	lookup := make([]uint8, 256)
	for i := 0; i < 256; i++ {
		lookup[i] = uint8(clamp(float64(i)*(1+change), 0, 255))
	}

	fn := func(c color.RGBA) color.RGBA {
		return color.RGBA{lookup[c.R], lookup[c.G], lookup[c.B], c.A}
	}

	return adjustment(c, fn)
}

func gammaAdj(c *canvas, change float64) error {
	gamma := math.Max(0.00001, change)

	lookup := make([]uint8, 256)
	for i := 0; i < 256; i++ {
		lookup[i] = uint8(clamp(math.Pow(float64(i)/255, 1.0/gamma)*255, 0, 255))
	}

	fn := func(c color.RGBA) color.RGBA {
		return color.RGBA{lookup[c.R], lookup[c.G], lookup[c.B], c.A}
	}

	return adjustment(c, fn)
}

func contrastAdj(c *canvas, change float64) error {
	lookup := make([]uint8, 256)
	for i := 0; i < 256; i++ {
		lookup[i] = uint8(clamp(((((float64(i)/255)-0.5)*(1+change))+0.5)*255, 0, 255))
	}

	fn := func(c color.RGBA) color.RGBA {
		return color.RGBA{lookup[c.R], lookup[c.G], lookup[c.B], c.A}
	}

	return adjustment(c, fn)
}

func saturationAdj(c *canvas, change float64) error {
	fn := func(in color.RGBA) color.RGBA {
		h, s, l := RGBToHSL(in)
		s = clamp(s*(1+change), 0.0, 1.0)
		out := HSLToRGB(h, s, l)
		out.A = in.A
		return out
	}

	return adjustment(c, fn)
}

func hueAdj(c *canvas, change float64) error {
	fn := func(in color.RGBA) color.RGBA {
		h, s, l := RGBToHSL(in)
		h = float64((int(h) + int(change)) % 360)
		out := HSLToRGB(h, s, l)
		out.A = in.A
		return out
	}

	return adjustment(c, fn)
}
