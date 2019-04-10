package core

import (
	"math"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/mth"
)

type blurType int

const (
	noBlur blurType = iota
	bBox
	bGaussian
)

var blurs = []blurType{
	bBox,
	bGaussian,
}

type blurOptions struct {
	t      blurType
	radius float64
}

func optionsToBlurOptions(o *Options) *blurOptions {
	b := &blurOptions{}
	return b
}

func runBlur(cv canvas.Canvas, o *blurOptions) (canvas.Canvas, error) {
	switch o.t {
	case bBox:
		return boxBlur(cv, o)
	case bGaussian:
		return gaussianBlur(cv, o)
	}
	return cv, nil
}

func boxBlur(cv canvas.Canvas, o *blurOptions) (canvas.Canvas, error) {
	radius := o.radius
	if radius >= 0 {
		length := int(math.Ceil(2*radius + 1))
		k := mth.NewMatrix(length, length)

		for x := 0; x < length; x++ {
			for y := 0; y < length; y++ {
				k.MX[y*length+x] = 1
			}
		}

		err := cv.Convolve(k.Normalized(), 0, false, false)
		return cv, err
	}
	return cv, nil
}

func gaussianBlur(cv canvas.Canvas, o *blurOptions) (canvas.Canvas, error) {
	radius := o.radius
	if radius >= 0 {
		length := int(math.Ceil(2*radius + 1))
		k := mth.NewMatrix(length, length)

		gaussianFn := func(x, y, sigma float64) float64 {
			return math.Exp(-(x*x/sigma + y*y/sigma))
		}

		for x := 0; x < length; x++ {
			for y := 0; y < length; y++ {
				k.MX[y*length+x] = gaussianFn(float64(x)-radius, float64(y)-radius, 4*radius)
			}
		}

		err := cv.Convolve(k.Normalized(), 0, false, false)
		return cv, err
	}
	return cv, nil
}

var blur = NewCommand(
	"", "blur", "Apply blur to a canvas", 1,
	func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("blur", flip.ContinueOnError)
		// bool flag for blurs
		// radius
		return fs
	},
	defaultCommandFunc,
	coreExec(blurStep)...,
).Command

func blurStep(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	cv.Print("execute blur")
	bo := optionsToBlurOptions(o)
	cv, bErr := runBlur(cv, bo)
	if bErr != nil {
		return cv, coreErrorHandler(o, bErr)
	}
	return cv, flip.ExitNo
}
