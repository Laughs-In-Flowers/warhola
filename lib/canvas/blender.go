package canvas

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/Laughs-In-Flowers/warhola/lib/util/xrr"
)

type Blender interface {
	Blend(string, string, image.Image, float64) error
	BlendTo(string, string, image.Image, float64) (Canvas, error)
}

type blendAct int

func stringToBlendAct(s string) blendAct {
	switch strings.ToLower(s) {
	case "normal":
		return normal
	case "add":
		return add
	case "multiply":
		return multiply
	case "overlay":
		return overlay
	case "softlight":
		return softlight
	case "screen":
		return screen
	case "difference":
		return difference
	case "divide":
		return divide
	case "colorburn":
		return colorburn
	case "exclusion":
		return exclusion
	case "colordodge":
		return colordodge
	case "linearburn":
		return linearburn
	case "linearlight":
		return linearlight
	case "subtract":
		return subtract
	case "opacity":
		return opacity
	case "darken":
		return darken
	case "lighten":
		return lighten
	}
	return noBlendAct
}

func (b blendAct) ToFunc(opt float64) blendColorFunc {
	switch b {
	case normal:
		return func(c0, c1 rgba164) rgba164 {
			return alphaComp(c0, c1)
		}
	case add:
		return func(c0, c1 rgba164) rgba164 {
			r := c0.R + c1.R
			g := c0.G + c1.G
			b := c0.B + c1.B

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case multiply:
		return func(c0, c1 rgba164) rgba164 {
			r := c0.R * c1.R
			g := c0.G * c1.G
			b := c0.B * c1.B

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case overlay:
		return func(c0, c1 rgba164) rgba164 {
			var r, g, b float64
			if c0.R > 0.5 {
				r = 1 - (1-2*(c0.R-0.5))*(1-c1.R)
			} else {
				r = 2 * c0.R * c1.R
			}
			if c0.G > 0.5 {
				g = 1 - (1-2*(c0.G-0.5))*(1-c1.G)
			} else {
				g = 2 * c0.G * c1.G
			}
			if c0.B > 0.5 {
				b = 1 - (1-2*(c0.B-0.5))*(1-c1.B)
			} else {
				b = 2 * c0.B * c1.B
			}

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case softlight:
		return func(c0, c1 rgba164) rgba164 {
			r := (1-2*c1.R)*c0.R*c0.R + 2*c0.R*c1.R
			g := (1-2*c1.G)*c0.G*c0.G + 2*c0.G*c1.G
			b := (1-2*c1.B)*c0.B*c0.B + 2*c0.B*c1.B

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case screen:
		return func(c0, c1 rgba164) rgba164 {
			r := 1 - (1-c0.R)*(1-c1.R)
			g := 1 - (1-c0.G)*(1-c1.G)
			b := 1 - (1-c0.B)*(1-c1.B)

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case difference:
		return func(c0, c1 rgba164) rgba164 {
			r := math.Abs(c0.R - c1.R)
			g := math.Abs(c0.G - c1.G)
			b := math.Abs(c0.B - c1.B)

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case divide:
		return func(c0, c1 rgba164) rgba164 {
			var r, g, b float64
			if c1.R == 0 {
				r = 1
			} else {
				r = c0.R / c1.R
			}
			if c1.G == 0 {
				g = 1
			} else {
				g = c0.G / c1.G
			}
			if c1.B == 0 {
				b = 1
			} else {
				b = c0.B / c1.B
			}

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case colorburn:
		return func(c0, c1 rgba164) rgba164 {
			var r, g, b float64
			if c1.R == 0 {
				r = 0
			} else {
				r = 1 - (1-c0.R)/c1.R
			}
			if c1.G == 0 {
				g = 0
			} else {
				g = 1 - (1-c0.G)/c1.G
			}
			if c1.B == 0 {
				b = 0
			} else {
				b = 1 - (1-c0.B)/c1.B
			}

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case exclusion:
		return func(c0, c1 rgba164) rgba164 {
			r := 0.5 - 2*(c0.R-0.5)*(c1.R-0.5)
			g := 0.5 - 2*(c0.G-0.5)*(c1.G-0.5)
			b := 0.5 - 2*(c0.B-0.5)*(c1.B-0.5)

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case colordodge:
		return func(c0, c1 rgba164) rgba164 {
			var r, g, b float64
			if c1.R == 1 {
				r = 1
			} else {
				r = c0.R / (1 - c1.R)
			}
			if c1.G == 1 {
				g = 1
			} else {
				g = c0.G / (1 - c1.G)
			}
			if c1.B == 1 {
				b = 1
			} else {
				b = c0.B / (1 - c1.B)
			}

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case linearburn:
		return func(c0, c1 rgba164) rgba164 {
			r := c0.R + c1.R - 1
			g := c0.G + c1.G - 1
			b := c0.B + c1.B - 1

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case linearlight:
		return func(c0, c1 rgba164) rgba164 {
			var r, g, b float64
			if c1.R > 0.5 {
				r = c0.R + 2*c1.R - 0.5
			} else {
				r = c0.R + 2*c1.R - 1
			}
			if c1.G > 0.5 {
				g = c0.G + 2*c1.G - 0.5
			} else {
				g = c0.G + 2*c1.G - 1
			}
			if c1.B > 0.5 {
				b = c0.B + 2*c1.B - 0.5
			} else {
				b = c0.B + 2*c1.B - 1
			}

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case subtract:
		return func(c0, c1 rgba164) rgba164 {
			r := c1.R - c0.R
			g := c1.G - c0.G
			b := c1.B - c0.B

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case opacity:
		percent := clamp(opt, 0, 1.0)
		return func(c0, c1 rgba164) rgba164 {
			r := c1.R*percent + (1-percent)*c0.R
			g := c1.G*percent + (1-percent)*c0.G
			b := c1.B*percent + (1-percent)*c0.B

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case darken:
		return func(c0, c1 rgba164) rgba164 {
			r := math.Min(c0.R, c1.R)
			g := math.Min(c0.G, c1.G)
			b := math.Min(c0.B, c1.B)

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	case lighten:
		return func(c0, c1 rgba164) rgba164 {
			r := math.Max(c0.R, c1.R)
			g := math.Max(c0.G, c1.G)
			b := math.Max(c0.B, c1.B)

			c2 := rgba164{R: r, G: g, B: b, A: c1.A}
			return alphaComp(c0, c2)
		}
	}
	return func(bgc rgba164, fbc rgba164) rgba164 { return bgc }
}

const (
	noBlendAct blendAct = iota
	normal
	add
	multiply
	overlay
	softlight
	screen
	difference
	divide
	colorburn
	exclusion
	colordodge
	linearburn
	linearlight
	subtract
	opacity
	darken
	lighten
)

type blendPos int

func stringToBlendPos(s string) blendPos {
	switch strings.ToLower(s) {
	case "fg", "foreground":
		return fg
	case "bg", "background":
		return bg
	}
	return noBlendPos
}

const (
	noBlendPos blendPos = iota
	fg
	bg
)

var (
	NoBlendPositionError = xrr.Xrror("not a valid blend position: %s").Out
	NoBlendActionError   = xrr.Xrror("not a valid blend action: %s").Out
)

func (c *canvas) Blend(pos, act string, o image.Image, opt float64) error {
	position := stringToBlendPos(pos)
	if position == noBlendPos {
		return NoBlendPositionError(pos)
	}
	action := stringToBlendAct(act)
	if action == noBlendAct {
		return NoBlendActionError(act)
	}

	np := scratch(c.pxl, o.ColorModel(), 0, 0)
	existingTo(o, np)

	var bgp, fgp *pxl
	switch position {
	case bg:
		bgp = c.pxl
		fgp = np
	case fg:
		bgp = np
		fgp = c.pxl
	}

	dstP, err := blend(bgp, fgp, action.ToFunc(opt))
	if err != nil {
		return err
	}

	c.pxl = dstP.clone(c.ColorModel())
	return nil
}

func (c *canvas) BlendTo(pos, act string, o image.Image, opt float64) (Canvas, error) {
	nc := c.Clone()
	err := nc.Blend(pos, act, o, opt)
	return nc, err
}

type blendColorFunc func(rgba164, rgba164) rgba164

func blend(bg *pxl, fg *pxl, fn blendColorFunc) (*pxl, error) {
	bgBounds := bg.Bounds()
	fgBounds := fg.Bounds()

	var w, h int
	if bgBounds.Dx() < fgBounds.Dx() {
		w = bgBounds.Dx()
	} else {
		w = fgBounds.Dx()
	}
	if bgBounds.Dy() < fgBounds.Dy() {
		h = bgBounds.Dy()
	} else {
		h = fgBounds.Dy()
	}

	bgSrc := bg.clone(color.RGBAModel)
	fgSrc := fg.clone(color.RGBAModel)
	dstP := scratch(bg, color.RGBAModel, w, h)

	parallel(h, func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < w; x++ {
				bgPos := y*bgSrc.Stride() + x*4
				fgPos := y*fgSrc.Stride() + x*4
				result := fn(
					newRGBA164(bgSrc.pix[bgPos+0], bgSrc.pix[bgPos+1], bgSrc.pix[bgPos+2], bgSrc.pix[bgPos+3]),
					newRGBA164(fgSrc.pix[fgPos+0], fgSrc.pix[fgPos+1], fgSrc.pix[fgPos+2], fgSrc.pix[fgPos+3]),
				)
				result.clamp()
				dstPos := y*dstP.Stride() + x*4
				dstP.pix[dstPos+0] = uint8(result.R * 255)
				dstP.pix[dstPos+1] = uint8(result.G * 255)
				dstP.pix[dstPos+2] = uint8(result.B * 255)
				dstP.pix[dstPos+3] = uint8(result.A * 255)
			}

		}
	})

	return dstP, nil
}
