package main

import (
	"context"
	"fmt"
	"image/draw"
	"path/filepath"
	"strings"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola"
)

type canvasOptions struct {
	New                                               bool
	Directory, Tag, Color, Extension, Stars, StarArgs string
	X, Y                                              int
}

func (o *canvasOptions) File() string {
	return fmt.Sprintf("%s.%s", o.Tag, o.Extension)
}

func (o *canvasOptions) Path() string {
	return filepath.Join(o.Directory, o.File())
}

func loadStars(o *canvasOptions) ([]warhola.Star, error) {
	var ret []warhola.Star
	ss := strings.Split(o.Stars, ",")
	for _, s := range ss {
		st, err := warhola.Load(s)
		if err != nil {
			return nil, err
		}
		ret = append(ret, st)
	}
	return ret, nil
}

var defaultCanvasOptions = canvasOptions{false, ".", "TAG", "RGBA", "png", "", "", 100, 100}

const canvasUse = `Create and manipulate individual canvases.`

func canvasFlags(o *canvasOptions) *flip.FlagSet {
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

func starArgs(o *canvasOptions) *warhola.StarArgs {
	path := o.Path()
	args := strings.Split(o.StarArgs, ",")
	return warhola.NewStarArgs(path, false, args...)
}

func fResult(cause, path string, err error, env *warhola.Env) flip.ExitStatus {
	env.Printf("%s at %s error: %s", cause, path, err)
	return flip.ExitFailure
}

func sResult(cause, path string, env *warhola.Env) flip.ExitStatus {
	e.Printf("%s at %s successful.", cause, path)
	return flip.ExitSuccess
}

func canvasCommand(o *canvasOptions) flip.ExecutionFunc {
	return func(c context.Context, a []string) flip.ExitStatus {
		path := o.Path()
		cm := warhola.StringToColorModel(o.Color)
		var action string
		var im draw.Image
		switch {
		case o.New:
			action = "create new canvas"
			im = warhola.NewFrom(cm, o.X, o.Y)
		default:
			action = "open existing canvas"
			var oe error
			im, oe = warhola.OpenImage(path, cm)
			if oe != nil {
				return fResult(action, path, oe, e)
			}
		}
		cv := warhola.New(path, im, cm)
		if o.Stars != "" {
			args := starArgs(o)
			stars, err := loadStars(o)
			if err != nil {
				return fResult("load stars", path, err, e)
			}
			err = cv.Apply(args, stars...)
			if err != nil {
				return fResult("apply stars", path, err, e)
			}
		}
		if err := cv.Save(); err != nil {
			return fResult("save canvas", path, err, e)
		}
		return sResult(action, path, e)
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
