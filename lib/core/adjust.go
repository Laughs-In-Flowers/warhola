package core

import (
	"fmt"
	"image/color"
	"math"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/mth"
)

var adjust = NewCommand(
	"", "adjust", "Adjust the brightness,gamma,contrast,hue or saturation of an image", 1,
	func(o *Options) *flip.FlagSet {
		v := o.Vector
		fs := flip.NewFlagSet("adjust", flip.ContinueOnError)
		for _, a := range adjustments {
			fs.Float64Vector(v, a.String(), a.key(), a.instruction())
		}
		return fs
	},
	defaultCommandFunc,
	coreExec(adjustStep)...,
).Command

func adjustStep(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	for _, a := range adjustments {
		change := o.ToFloat64(a.key())
		afn := a.fn(change)
		if afn != nil {
			t := a.String()
			cv.Printf("executing %s: %f", t, change)
			err := cv.Adjust(afn)
			cv.Printf("adjusted %s...", t)
			if err != nil {
				return cv, coreErrorHandler(o, err)
			}
		}
	}
	return cv, flip.ExitNo
}

type adjustAction int

const (
	noAdjust adjustAction = iota
	brightness
	contrast
	gamma
	hue
	saturation
)

var adjustments = []adjustAction{
	brightness,
	contrast,
	gamma,
	hue,
	saturation,
}

func (a adjustAction) String() string {
	switch a {
	case brightness:
		return "brightness"
	case contrast:
		return "contrast"
	case gamma:
		return "gamma"
	case hue:
		return "hue"
	case saturation:
		return "saturation"
	}
	return "no.adjust"
}

func (a adjustAction) key() string {
	return fmt.Sprintf("adjust.%s", a.String())
}

func (a adjustAction) scope() (float64, float64) {
	switch a {
	case brightness:
		return -1, 1
	case contrast:
		return -1, 1
	case gamma:
		return 0, math.Inf(+1)
	case hue:
		return -360, 360
	case saturation:
		return -1, 1
	}
	return 0, 0
}

func (a adjustAction) instruction() string {
	switch a {
	case gamma:
		return "Amount of gamma adjustment to apply, greater than 0"
	default:
		lr, hr := a.scope()
		return fmt.Sprintf("Amount of %s adjustment to apply, %F to %F", a.String(), lr, hr)
	}
	return ""
}

func (a adjustAction) fn(change float64) canvas.AdjustmentFunc {
	if change == 0 {
		return nil
	}
	switch a {
	case brightness:
		lookup := make([]uint8, 256)
		for i := 0; i < 256; i++ {
			lookup[i] = uint8(mth.Clamp(float64(i)*(1+change), 0, 255))
		}
		return func(c color.RGBA) color.RGBA {
			return color.RGBA{lookup[c.R], lookup[c.G], lookup[c.B], c.A}
		}
	case contrast:
		lookup := make([]uint8, 256)
		for i := 0; i < 256; i++ {
			lookup[i] = uint8(mth.Clamp(((((float64(i)/255)-0.5)*(1+change))+0.5)*255, 0, 255))
		}
		return func(c color.RGBA) color.RGBA {
			return color.RGBA{lookup[c.R], lookup[c.G], lookup[c.B], c.A}
		}
	case gamma:
		gamma := math.Max(0.00001, change)
		lookup := make([]uint8, 256)
		for i := 0; i < 256; i++ {
			lookup[i] = uint8(mth.Clamp(math.Pow(float64(i)/255, 1.0/gamma)*255, 0, 255))
		}
		return func(c color.RGBA) color.RGBA {
			return color.RGBA{lookup[c.R], lookup[c.G], lookup[c.B], c.A}
		}
	case hue:
		return func(c color.RGBA) color.RGBA {
			h, s, l := canvas.RGBToHSL(c)
			h = float64((int(h) + int(change)) % 360)
			out := canvas.HSLToRGB(h, s, l)
			out.A = c.A
			return out
		}
	case saturation:
		return func(c color.RGBA) color.RGBA {
			h, s, l := canvas.RGBToHSL(c)
			s = mth.Clamp(s*(1+change), 0.0, 1.0)
			out := canvas.HSLToRGB(h, s, l)
			out.A = c.A
			return out
		}
	}
	return nil
}
