package core

import (
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
)

var noise = NewCommand(
	"", "noise", "Apply noise to a canvas", 1,
	func(o *Options) *flip.FlagSet {
		//v := o.Vector
		fs := flip.NewFlagSet("noise", flip.ContinueOnError)
		return fs
	},
	defaultCommandFunc,
	coreExec(noiseStep)...,
).Command

func noiseStep(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	//cv.Print("execute noise")
	return cv, flip.ExitNo
}
