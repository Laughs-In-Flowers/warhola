package core

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/mth"
	"github.com/Laughs-In-Flowers/xrr"
)

type blendAction int

const (
	noblend blendAction = iota
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

var blends = []blendAction{
	normal,
	add,
	multiply,
	overlay,
	softlight,
	screen,
	difference,
	divide,
	colorburn,
	exclusion,
	colordodge,
	linearburn,
	linearlight,
	subtract,
	opacity,
	darken,
	lighten,
}

func (b blendAction) String() string {
	switch b {
	case normal:
		return "normal"
	case add:
		return "add"
	case multiply:
		return "multiply"
	case overlay:
		return "overlay"
	case softlight:
		return "softLight"
	case screen:
		return "screen"
	case difference:
		return "difference"
	case divide:
		return "divide"
	case colorburn:
		return "colorBurn"
	case exclusion:
		return "exclusion"
	case colordodge:
		return "colorDodge"
	case linearburn:
		return "linearBurn"
	case linearlight:
		return "linearLight"
	case subtract:
		return "subtract"
	case opacity:
		return "opacity"
	case darken:
		return "darken"
	case lighten:
		return "lighten"
	}
	return "noBlend"
}

func stringToBlend(s string) blendAction {
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
	return noblend
}

type (
	BlendFunc = canvas.BlendFunc
	RGBA164   = canvas.RGBA164
	Position  = canvas.BlendPosition
)

func (b blendAction) fn(opt float64) BlendFunc {
	switch b {
	case normal:
		return func(c0, c1 RGBA164) RGBA164 {
			return canvas.AlphaComp(c0, c1)
		}
	case add:
		return func(c0, c1 RGBA164) RGBA164 {
			r := c0.R + c1.R
			g := c0.G + c1.G
			b := c0.B + c1.B

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case multiply:
		return func(c0, c1 RGBA164) RGBA164 {
			r := c0.R * c1.R
			g := c0.G * c1.G
			b := c0.B * c1.B

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case overlay:
		return func(c0, c1 RGBA164) RGBA164 {
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

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case softlight:
		return func(c0, c1 RGBA164) RGBA164 {
			r := (1-2*c1.R)*c0.R*c0.R + 2*c0.R*c1.R
			g := (1-2*c1.G)*c0.G*c0.G + 2*c0.G*c1.G
			b := (1-2*c1.B)*c0.B*c0.B + 2*c0.B*c1.B

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case screen:
		return func(c0, c1 RGBA164) RGBA164 {
			r := 1 - (1-c0.R)*(1-c1.R)
			g := 1 - (1-c0.G)*(1-c1.G)
			b := 1 - (1-c0.B)*(1-c1.B)

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case difference:
		return func(c0, c1 RGBA164) RGBA164 {
			r := math.Abs(c0.R - c1.R)
			g := math.Abs(c0.G - c1.G)
			b := math.Abs(c0.B - c1.B)

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case divide:
		return func(c0, c1 RGBA164) RGBA164 {
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

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case colorburn:
		return func(c0, c1 RGBA164) RGBA164 {
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

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case exclusion:
		return func(c0, c1 RGBA164) RGBA164 {
			r := 0.5 - 2*(c0.R-0.5)*(c1.R-0.5)
			g := 0.5 - 2*(c0.G-0.5)*(c1.G-0.5)
			b := 0.5 - 2*(c0.B-0.5)*(c1.B-0.5)

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case colordodge:
		return func(c0, c1 RGBA164) RGBA164 {
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

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case linearburn:
		return func(c0, c1 RGBA164) RGBA164 {
			r := c0.R + c1.R - 1
			g := c0.G + c1.G - 1
			b := c0.B + c1.B - 1

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case linearlight:
		return func(c0, c1 RGBA164) RGBA164 {
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

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case subtract:
		return func(c0, c1 RGBA164) RGBA164 {
			r := c1.R - c0.R
			g := c1.G - c0.G
			b := c1.B - c0.B

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case opacity:
		percent := mth.Clamp(opt, 0, 1.0)
		return func(c0, c1 RGBA164) RGBA164 {
			r := c1.R*percent + (1-percent)*c0.R
			g := c1.G*percent + (1-percent)*c0.G
			b := c1.B*percent + (1-percent)*c0.B

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case darken:
		return func(c0, c1 RGBA164) RGBA164 {
			r := math.Min(c0.R, c1.R)
			g := math.Min(c0.G, c1.G)
			b := math.Min(c0.B, c1.B)

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	case lighten:
		return func(c0, c1 RGBA164) RGBA164 {
			r := math.Max(c0.R, c1.R)
			g := math.Max(c0.G, c1.G)
			b := math.Max(c0.B, c1.B)

			c2 := RGBA164{R: r, G: g, B: b, A: c1.A}
			return canvas.AlphaComp(c0, c2)
		}
	}
	return func(bgc RGBA164, fbc RGBA164) RGBA164 { return bgc }
}

func fgbgFlag(o *Options, fs *flip.FlagSet) {
	fs.StringVector(o.Vector, "fg", "blend.fg", "a foreground image to the canvas blend operation")
	fs.StringVector(o.Vector, "bg", "blend.bg", "a background image to the canvas blend operation")
}

func extractImage(k string, o *Options) image.Image {
	path := o.ToString(k)
	if path != "" {
		img, err := canvas.OpenTo(path)
		if err != nil {
			return nil
		}
		return img
	}
	return nil
}

var extractGroundError = xrr.Xrror("Could not extract one of 'foreground' or 'background' for blending.\nOne and only one of foreground or background must be specified.")

func hasFgbgFlag(o *Options) (Position, image.Image, error) {
	fg := extractImage("blend.fg", o)
	bg := extractImage("blend.bg", o)
	switch {
	case fg == nil && bg != nil:
		return canvas.BG, bg, nil
	case bg == nil && fg != nil:
		return canvas.FG, fg, nil
	}
	return canvas.NoBlendPosition, nil, extractGroundError
}

func blendFlag(o *Options, fs *flip.FlagSet, b blendAction) {
	s := b.String()
	fs.BoolVector(o.Vector,
		s,
		fmt.Sprintf("blend.%s", s),
		fmt.Sprintf("blend: %s", s),
	)
}

func hasBlendFlag(b blendAction, o *Options) bool {
	return o.ToBool(fmt.Sprintf("blend.%s", b.String()))
}

func optionFlag(o *Options, fs *flip.FlagSet) {
	fs.Float64Vector(o.Vector, "opt", "blend.option", "a float64 option value passed to the blend operation")
}

func hasOptionFlag(o *Options) float64 {
	return o.ToFloat64("blend.option")
}

var blend = NewCommand(
	"", "blend", "Blend a canvas with any provided image as foreground or background", 1,
	func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("blend", flip.ContinueOnError)
		fgbgFlag(o, fs)
		for _, b := range blends {
			blendFlag(o, fs, b)
		}
		optionFlag(o, fs)
		return fs
	},
	defaultCommandFunc,
	coreExec(blendStep)...,
).Command

func blendStep(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	cv.Print("execute blend")
	pos, img, pErr := hasFgbgFlag(o)
	if pErr != nil {
		return cv, coreErrorHandler(o, pErr)
	}
	opt := hasOptionFlag(o)
	for _, b := range blends {
		if hasBlendFlag(b, o) {
			fn := b.fn(opt)
			bErr := cv.Blend(img, pos, fn)
			if bErr != nil {
				return cv, coreErrorHandler(o, bErr)
			}
		}
	}
	return cv, flip.ExitNo
}
