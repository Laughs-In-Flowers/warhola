package core

import (
	"image"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
)

var (
	fliip = NewCommand(
		"", "flip", "Flip an image in an opposite direction", 1,
		func(o *Options) *flip.FlagSet {
			v := o.Vector
			fs := flip.NewFlagSet("flip", flip.ContinueOnError)
			fs.BoolVector(v, "vertical", "flip.vertical", "flip the image vertically")
			fs.BoolVector(v, "horizontal", "flip.horizontal", "flip the image horizontally")
			return fs
		},
		defaultCommandFunc,
		coreExec(func(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
			v, h := o.ToBool("flip.vertical"), o.ToBool("flip.horizontal")
			var err error
			if v {
				err = cv.Flip(canvas.TVertical)
			}
			if h {
				err = cv.Flip(canvas.THorizontal)
			}
			return cv, coreErrorHandler(o, err)
		})...,
	).Command

	rotate = NewCommand(
		"", "rotate", "Rotate an image", 1,
		func(o *Options) *flip.FlagSet {
			v := o.Vector
			fs := flip.NewFlagSet("rotate", flip.ContinueOnError)
			fs.Float64Vector(v, "angle", "rotate.angle", "the angle of rotation to apply")
			fs.BoolVector(v, "preserveSize", "rotate.preserve", "preserve the size the image")
			fs.IntVector(v, "pivotX", "rotate.pivot.x", "the x value of the rotation pivot point")
			fs.IntVector(v, "pivotY", "rotate.pivot.y", "the y value of the rotation pivot point")
			return fs
		},
		defaultCommandFunc,
		coreExec(func(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
			angle := o.ToFloat64("rotate.angle")
			preserve := o.ToBool("rotate.preserve")
			x, y := o.ToInt("rotate.pivot.x"), o.ToInt("rotate.pivot.y")
			err := cv.Rotate(angle, preserve, image.Point{x, y})
			return cv, coreErrorHandler(o, err)
		})...,
	).Command

	shear = NewCommand(
		"", "shear", "Linear transformation along an axis", 1,
		func(o *Options) *flip.FlagSet {
			v := o.Vector
			fs := flip.NewFlagSet("shear", flip.ContinueOnError)
			fs.Float64Vector(v, "vertical", "shear.vertical.angle", "the angle of vertical shear to apply")
			fs.Float64Vector(v, "horizontal", "shear.horizontal.angle", "the angle of horizontal  shear to apply")
			return fs
		},
		defaultCommandFunc,
		coreExec(func(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
			v, h := o.ToFloat64("shear.vertical.angle"), o.ToFloat64("shear.horizontal.angle")
			var err error
			if v != 0 {
				err = cv.Shear(canvas.TVertical, v)
			}
			if h != 0 {
				err = cv.Shear(canvas.THorizontal, h)
			}
			return cv, coreErrorHandler(o, err)
		})...,
	).Command

	translate = NewCommand(
		"", "translate", "repositions a copy of an image by dx on the x-axis and dy on the y-axis", 1,
		func(o *Options) *flip.FlagSet {
			v := o.Vector
			fs := flip.NewFlagSet("translate", flip.ContinueOnError)
			fs.IntVector(v, "dx", "translate.dx", "reposition by dx on the x-axis")
			fs.IntVector(v, "dy", "translate.dy", "reposition by dy on the y-axis")
			return fs
		},
		defaultCommandFunc,
		coreExec(func(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
			dx, dy := o.ToInt("translate.dx"), o.ToInt("translate.dy")
			err := cv.Translate(dx, dy)
			return cv, coreErrorHandler(o, err)
		})...,
	).Command
)

func registerTranslateCmds(cm cmdMap) {
	cm.Register("flip", fliip)
	cm.Register("rotate", rotate)
	cm.Register("shear", shear)
	cm.Register("translate", translate)
}
