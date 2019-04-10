package core

import (
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
)

var convolve = NewCommand(
	"", "convolve", "Apply a convolution matrice to a canvas", 1,
	func(o *Options) *flip.FlagSet {
		//v := o.Vector
		fs := flip.NewFlagSet("convolve", flip.ContinueOnError)
		return fs
	},
	defaultCommandFunc,
	coreExec(convolveStep)...,
).Command

func convolveStep(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	//cv.Print("execute convolution")
	return cv, flip.ExitNo
}
