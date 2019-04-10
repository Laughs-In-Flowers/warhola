package core

import (
	"image"
	"math"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/geo"
)

var (
	crop = NewCommand(
		"", "crop", "Crop an image", 1,
		func(o *Options) *flip.FlagSet {
			v := o.Vector
			fs := flip.NewFlagSet("crop", flip.ContinueOnError)
			geo.GeometryVectorFlag(fs, v, "crop.geometry")
			return fs
		},
		defaultCommandFunc,
		coreExec(
			func(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
				g := o.pullGeometry("crop.geometry")
				b := cv.Bounds()
				var x, y int = b.Max.X, b.Max.Y
				if g.X != 0 {
					x = g.X
				}
				if g.Y != 0 {
					y = g.Y
				}
				rect := image.Rect(g.OffsetX, g.OffsetY, x, y)
				cv.Printf("execute crop of %v to %v", b, rect)
				cv.Crop(rect)
				cv.Print("cropped...")
				return cv, flip.ExitNo
			})...,
	).Command

	resize = NewCommand(
		"", "resize", "Resize an image", 1,
		func(o *Options) *flip.FlagSet {
			v := o.Vector
			fs := flip.NewFlagSet("resize", flip.ContinueOnError)
			geo.GeometryVectorFlag(fs, v, "resize.geometry")
			fs.StringVectorVar(v, "filter", "resize.filter", "nearest", "the resample filter to use in resizing\n\t\t[nearest|\n\t\tbox|\n\t\tlinear|\n\t\tgaussian|\n\t\tmitchellnetravali|\n\t\tcatmullrom|\n\t\tlanczos|\n\t\tbartlett|\n\t\thermite|\n\t\tbspline|\n\t\thann|\n\t\thamming|\n\t\tblackman|\n\t\twelch|\n\t\tcosine]")
			return fs
		},
		defaultCommandFunc,
		coreExec(
			func(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
				b := cv.Bounds()
				cv.Printf("execute resize of (w:%d, h: %d)", b.Max.X, b.Max.Y)
				var aw, ah int
				w, h := b.Max.X, b.Max.Y
				g := o.pullGeometry("resize.geometry")
				if g.X != 0 {
					aw = g.X
				}
				if g.Y != 0 {
					ah = g.Y
				}
				if g.ScaleX > 0 {
					aw = int(float64(w) * ((g.ScaleX / 100) + 1))
				}
				if g.ScaleX < 0 {
					aw = int(float64(w) * (1 - (math.Abs(g.ScaleX) / 100)))
				}
				if g.ScaleY > 0 {
					ah = int(float64(h) * ((g.ScaleY / 100) + 1))
				}
				if g.ScaleY < 0 {
					ah = int(float64(h) * (1 - (math.Abs(g.ScaleY) / 100)))
				}
				filter := stringToFilter(o.ToString("resize.filter"))
				cv.Printf("resizing to (w: %d, h: %d)", aw, ah)
				err := cv.Resize(aw, ah, filter)
				cv.Printf("resized....")
				if err != nil {
					return cv, coreErrorHandler(o, err)
				}
				return cv, flip.ExitNo
			})...,
	).Command
)

func registerTransformCmds(cm cmdMap) {
	cm.Register("crop", crop)
	cm.Register("resize", resize)
}
