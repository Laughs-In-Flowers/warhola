package main

import (
	"context"
	"fmt"
	"image/draw"
	"path/filepath"
	"strings"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/factory"
	"github.com/Laughs-In-Flowers/warhola/lib/star"
)

type cOptions struct {
	New                                               bool
	Directory, Tag, Color, Extension, Stars, StarArgs string
	X, Y                                              int
}

func (o *cOptions) File() string {
	return fmt.Sprintf("%s.%s", o.Tag, o.Extension)
}

func (o *cOptions) Path() string {
	return filepath.Join(o.Directory, o.File())
}

func loadStars(o *cOptions) ([]star.Star, error) {
	var ret []star.Star
	ss := strings.Split(o.Stars, ",")
	for _, s := range ss {
		st, err := star.Load(s)
		if err != nil {
			return nil, err
		}
		ret = append(ret, st)
	}
	return ret, nil
}

var defaultCanvasOptions = cOptions{false, ".", "TAG", "RGBA", "png", "", "", 100, 100}

const canvasUse = `Create and manipulate individual canvases.`

func canvasFlags(o *cOptions) *flip.FlagSet {
	fs := flip.NewFlagSet("canvas", flip.ContinueOnError)
	fs.BoolVar(&o.New, "new", o.New, "Create a new canvas.")
	fs.StringVar(&o.Directory, "directory", o.Directory, "The directory of the canvas.")
	fs.StringVar(&o.Tag, "tag", o.Tag, "The name tag of the canvas.")
	fs.StringVar(&o.Color, "color", o.Color, "The color model of the canvas. [GRAY|ALPHA|RGBA|RGBA64|NRGBA|NRGBA64|CMYK]")
	fs.StringVar(&o.Extension, "extension", o.Extension, "The file extension for the canvas. [png|jpeg]")
	fs.StringVar(&o.Stars, "stars", o.Stars, "A comma delimited list of stars to apply to the canvas.")
	fs.StringVar(&o.StarArgs, "starArgs", o.StarArgs, "A comma delimited list of key:value args used by the stars.")
	fs.IntVar(&o.X, "X", o.X, "X dimension of the canvas.")
	fs.IntVar(&o.Y, "Y", o.Y, "Y dimension of the canvas.")
	return fs
}

func starArgs(o *cOptions) *star.Args {
	path := o.Path()
	args := strings.Split(o.StarArgs, ",")
	return star.NewArgs(path, false, args...)
}

func fResult(cause, path string, err error, f *factory.Factory) flip.ExitStatus {
	f.Printf("%s at %s error: %s", cause, path, err)
	return flip.ExitFailure
}

func sResult(cause, path string, f *factory.Factory) flip.ExitStatus {
	f.Printf("%s at %s successful.", cause, path)
	return flip.ExitSuccess
}

func canvasCommand(o *cOptions) flip.ExecutionFunc {
	return func(c context.Context, a []string) flip.ExitStatus {
		path := o.Path()
		cm := canvas.StringToColorModel(o.Color)
		var action string
		var im draw.Image
		switch {
		case o.New:
			action = "create new canvas"
			im = canvas.NewFrom(cm, o.X, o.Y)
		default:
			action = "open existing canvas"
			var oe error
			im, oe = canvas.OpenImage(path, cm)
			if oe != nil {
				return fResult(action, path, oe, f)
			}
		}
		cv := canvas.New(path, im, cm)
		if o.Stars != "" {
			args := starArgs(o)
			stars, err := loadStars(o)
			if err != nil {
				return fResult("load stars", path, err, f)
			}
			err = cv.Apply(args, stars...)
			if err != nil {
				return fResult("apply stars", path, err, f)
			}
		}
		if err := cv.Save(); err != nil {
			return fResult("save canvas", path, err, f)
		}
		return sResult(action, path, f)
	}
}

func CanvasCommand() flip.Command {
	so := &defaultCanvasOptions
	fs := canvasFlags(so)
	cmd := canvasCommand(so)
	return flip.NewCommand(
		"",
		"canvas",
		canvasUse,
		1,
		cmd,
		fs,
	)
}

func init() {
	flip.RegisterGroup("canvas", 10, CanvasCommand())
}
