package main

import (
	"context"
	"fmt"
	"image/draw"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/factory"
	"github.com/Laughs-In-Flowers/warhola/lib/star"
)

type cOptions struct {
	New, Debug, View                                  bool
	Directory, Tag, Color, Extension, Stars, StarArgs string
	X, Y                                              int
}

func (o *cOptions) File() string {
	return fmt.Sprintf("%s.%s", o.Tag, o.Extension)
}

func (o *cOptions) Path() string {
	return filepath.Join(o.Directory, o.File())
}

func loadStars(o *Options) ([]star.Star, error) {
	ss := strings.Split(o.Stars, ",")
	reqs, err := F.GetStars(o.Path(), o.Debug, ss...)
	if err != nil {
		return nil, err
	}
	return reqs, nil
}

var defaultCanvasOptions = cOptions{false, false, false, ".", "TAG", "RGBA", "png", "", "", 100, 100}

const canvasUse = `Create and manipulate individual canvases.`

func canvasFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("canvas", flip.ContinueOnError)
	fs.BoolVar(&o.New, "new", o.New, "Create a new canvas.")
	fs.BoolVar(&o.Debug, "debug", o.Debug, "Run any actions in debug mode where available.")
	fs.BoolVar(&o.View, "view", o.View, "Start a valhalla viewer for the canvas image after all actions.")
	fs.StringVar(&o.Directory, "directory", o.Directory, "The directory of the canvas.")
	fs.StringVar(&o.Tag, "tag", o.Tag, "The name tag of the canvas.")
	fs.StringVar(&o.Color, "color", o.Color, "The color model of the canvas. [GRAY|ALPHA|RGBA|RGBA64|NRGBA|NRGBA64|CMYK]")
	fs.StringVar(&o.Extension, "extension", o.Extension, "The file extension for the canvas. [png|jpeg]")
	fs.StringVar(&o.Stars, "stars", o.Stars, "A comma delimited list of stars to apply to the canvas.")
	fs.IntVar(&o.X, "X", o.X, "X dimension of the canvas.")
	fs.IntVar(&o.Y, "Y", o.Y, "Y dimension of the canvas.")
	return fs
}

func failure(cause, path string, err error, f *factory.Factory) flip.ExitStatus {
	f.Printf("%s at %s error: %s", cause, path, err)
	return flip.ExitFailure
}

func success(cause, path string, f *factory.Factory) flip.ExitStatus {
	f.Printf("%s at %s successful.", cause, path)
	return flip.ExitSuccess
}

func view(formatter, path string) *exec.Cmd {
	ret := exec.Command("valhalla", "-formatter", formatter, "view", "-path", path)
	ret.Stdout = os.Stdout
	return ret
}

func canvasCommand(o *Options) flip.ExecutionFunc {
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
				return failure(action, path, oe, F)
			}
		}
		cv := canvas.New(path, im, cm)
		if o.Stars != "" {
			stars, err := loadStars(o)
			if err != nil {
				return failure("load stars requests", path, err, F)
			}
			err = cv.Apply(stars...)
			if err != nil {
				return failure("apply star requests", path, err, F)
			}
		}
		if err := cv.Save(); err != nil {
			return failure("save canvas", path, err, F)
		}
		if o.View {
			vc := view(o.formatter, path)
			vc.Start()
		}
		return success(action, path, F)
	}
}

func CanvasCommand() flip.Command {
	fs := canvasFlags(O)
	cmd := canvasCommand(O)
	return flip.NewCommand(
		"",
		"canvas",
		canvasUse,
		1,
		cmd,
		fs,
	)
}
