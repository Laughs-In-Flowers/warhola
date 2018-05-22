package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/plugin"
	"github.com/Laughs-In-Flowers/warhola/lib/plugin/plugins/builtins"
	"github.com/Laughs-In-Flowers/warhola/lib/util/ctx"
)

var pluginDirs = []string{"plugins"}

func dirs(paths string, b *bytes.Buffer) []string {
	ret := pluginDirs
	if paths != "" {
		var pth string
		var err error
		spl := strings.Split(paths, ",")
		for _, v := range spl {
			pth, err = filepath.Abs(v)
			if err != nil {
				failure(b, "plugin dirs", err)
			}
			ret = append(ret, pth)
		}
	}
	return ret
}

func pluginSetting(a []string) {
	var err error
	var d string
	for idx, flg := range a {
		switch flg {
		case "-plugins", "--plugins":
			d = a[idx+1]
		default:
			continue
		}
	}

	msg := new(bytes.Buffer)

	P, err = plugin.New(dirs(d, msg)...)
	if err != nil {
		failure(msg, "plugins creation", err)
	}
	err = P.Load()
	if err != nil {
		failure(msg, "plugin initialization", err)
	}
	cmds, err := P.Get("ALL")
	if err != nil {
		failure(msg, "plugin command loading", err)
	}

	F.SetGroup("aux", 100, cmds...)
}

type Options struct {
	*tOptions
	*cOptions
}

func defaultOptions() *Options {
	return &Options{
		&defaultTopOptions,
		&defaultCanvasOptions,
	}
}

func tExecute(o *Options, c context.Context, a []string) (context.Context, flip.ExitStatus) {
	var status flip.ExitStatus
	for _, fn := range executing {
		c, status = fn(o, c)
		if status != flip.ExitNo {
			return c, status
		}
	}
	return c, flip.ExitNo
}

type execution func(o *Options, c context.Context) (context.Context, flip.ExitStatus)

var executing = []execution{
	debugPush,
	debugInfoPush,
	debugCleanup,
	logSetting,
	logPush,
	pluginPush,
	canvasSetting,
	canvasPush,
	canvasCleanup,
	mainCleanup,
}

type tOptions struct {
	formatter string
	log.Logger
}

var defaultTopOptions tOptions = tOptions{
	"null",
	log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter()),
}

func tFlags(fs *flip.FlagSet, o *Options) *flip.FlagSet {
	fs.StringVar(&o.formatter, "formatter", o.formatter, "Specify the log formatter. [null|raw|stdout]")
	return fs
}

func debugPush(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	c = context.WithValue(c, 0, o.Debug)
	return c, flip.ExitNo
}

func debugInfoPush(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	di := make(map[string]string)
	c = context.WithValue(c, 1, di)
	return c, flip.ExitNo
}

func debugCleanup(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	if o.Debug {
		F.SetCleanup(flip.ExitAny, debugCleanupFunc)
	}
	return c, flip.ExitNo
}

func debugCleanupFunc(c context.Context) {
	if ctx.Debug(c) {
		if l := ctx.Log(c); l != nil {
			l.Println("- start debug information -----")
			if d := ctx.DebugMap(c); d != nil {
				cv := ctx.Canvas(c)
				builtins.TextDebugReset(
					builtins.BuiltInsOptions,
					ctx.DebugMapCollapse(d),
				)
				builtins.WriteText(cv, builtins.BuiltInsOptions)
				for k, v := range d {
					l.Printf("%s: %s", k, v)
				}
			}
			l.Println("- end debug information -----")
		}
	}
}

func logSetting(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	if o.formatter != "null" {
		switch o.formatter {
		case "text", "stdout":
			o.SwapFormatter(log.GetFormatter("warhola_text"))
		default:
			o.SwapFormatter(log.GetFormatter(o.formatter))
		}
	}
	return c, flip.ExitNo
}

func logPush(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	c = context.WithValue(c, 2, o.Logger)
	return c, flip.ExitNo
}

func pluginPush(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	c = context.WithValue(c, 3, P)
	return c, flip.ExitNo
}

type cOptions struct {
	Debug      bool
	Color      string
	File, Kind string
	X, Y       int
	PP         float64
	PPU        string
}

var defaultCanvasOptions = cOptions{
	false,
	"RGBA",
	"",
	"png",
	100,
	100,
	300,
	"inch",
}

func cFlags(fs *flip.FlagSet, o *Options) *flip.FlagSet {
	fs.BoolVar(&o.Debug, "debug", o.Debug, "Run any actions in debug mode where available.")
	fs.StringVar(&o.Color, "color", o.Color, "The color model of the canvas. [GRAY|ALPHA|RGBA|RGBA64|NRGBA|NRGBA64|CMYK]")
	fs.StringVar(&o.File, "file", o.File, "The full path of the new or existing image")
	fs.StringVar(&o.Kind, "kind", o.Kind, "Kind of file for the canvas. [png|jpeg]")
	fs.IntVar(&o.X, "X", o.X, "X dimension of the canvas.")
	fs.IntVar(&o.Y, "Y", o.Y, "Y dimension of the canvas.")
	fs.Float64Var(&o.PP, "PP", o.PP, "points per unit where unit is specified in option PP")
	fs.StringVar(&o.PPU, "PPU", o.PPU, "unit of measurement for points per")
	return fs
}

func canvasSetting(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	var err error
	ocm = o.Color
	CV, err = canvas.New(canvas.SetLogger(o.Logger),
		canvas.SetDebug(o.Debug),
		canvas.SetColorModel(canvas.WorkingColorModel),
		canvas.SetPath(o.File),
		canvas.SetKind(o.Kind),
		canvas.SetMeasure(o.PP, o.PPU),
		canvas.SetRect(o.X, o.Y),
	)
	if err != nil {
		CV.Fatalf("canvas error: %s", err)
		return nil, flip.ExitFailure
	}
	return c, flip.ExitNo
}

func canvasPush(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	c = context.WithValue(c, 4, CV)
	return c, flip.ExitNo
}

func canvasCleanup(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	F.SetCleanup(flip.ExitAny, canvasCleanupFunc)
	return c, flip.ExitNo
}

func canvasCleanupFunc(c context.Context) {
	if l := ctx.Log(c); l != nil {
		l.Println("clean up")
		var cuErr error
		switch ocm {
		case canvas.WorkingColorModel:
			cuErr = CV.Save()
		default:
			cuErr = CV.SaveTo(ocm)
		}
		if cuErr != nil {
			l.Printf("cleanup error: %s", cuErr)
		}
	} else {
		CV.Save()
	}
}

func mainCleanup(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	F.SetCleanup(flip.ExitAny, func(context.Context) {
		o.Println("done")
	})
	return c, flip.ExitNo
}

func TopCommand() flip.Command {
	fs := flip.NewFlagSet("top", flip.ContinueOnError)
	fs = tFlags(fs, O)
	fs = cFlags(fs, O)

	return flip.NewCommand(
		"",
		"warhola",
		"Top level options use.",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return tExecute(O, c, a)
		},
		fs,
	)
}

type sOptions struct {
	all, plugin, font bool
	fontDirs          string
}

var defaultStatusOptions sOptions = sOptions{true, false, false, ""}

func sFlags(fs *flip.FlagSet, o *sOptions) *flip.FlagSet {
	fs.BoolVar(&o.all, "all", o.all, "Complete status list")
	fs.BoolVar(&o.plugin, "plugin", o.plugin, "List of all plugins by directory")
	fs.BoolVar(&o.font, "font", o.font, "List of all fonts for the specified fontsDir")
	fs.StringVar(&o.fontDirs, "fontDirs", o.fontDirs, "The fontsDir to look in.")
	return fs
}

func StatusCommand() flip.Command {
	so := &defaultStatusOptions
	fs := flip.NewFlagSet("status", flip.ContinueOnError)
	fs = sFlags(fs, so)

	return flip.NewCommand(
		"",
		"status",
		"warhola status information",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			msg := new(bytes.Buffer)
			switch {
			case so.plugin, so.font:
				so.all = false
			}
			switch {
			case so.plugin:
				writePlugins(msg)
			case so.font:
				writeFonts(msg, so)
			case so.all:
				writePlugins(msg)
				writeFonts(msg, so)
			}
			return c, flip.ExitSuccess
		},
		fs,
	)
}

func writeOnce(msg *bytes.Buffer, w io.Writer) {
	w.Write(msg.Bytes())
	msg.Reset()
}

func failure(msg *bytes.Buffer, cause string, err error) flip.ExitStatus {
	msg.WriteString(fmt.Sprintf("%s error:\n\t%s\n", cause, err))
	writeOnce(msg, os.Stdout)
	return flip.ExitFailure
}

func writePlugins(msg *bytes.Buffer) {
	ps, err := P.Plugins()
	if err != nil {
		msg.WriteString(fmt.Sprintf("unable to write plugins: %s", err.Error()))
		writeOnce(msg, os.Stderr)
	}
	msg.WriteString(fmt.Sprintf("%s\n", "available plugins"))
	for k, v := range ps {
		msg.WriteString(fmt.Sprintf("\t%s\n", k))
		for _, vv := range v {
			msg.WriteString(fmt.Sprintf("\t\t%s\n", vv))
		}
	}
	writeOnce(msg, os.Stdout)
}

func writeFonts(msg *bytes.Buffer, so *sOptions) {
	f := builtins.LF
	f.SetDir(builtins.FontDirectories(so.fontDirs)...)
	fs := f.List()
	msg.WriteString("available fonts\n")
	for _, v := range fs {
		msg.WriteString(fmt.Sprintf("\t%s\n", v))
	}
	writeOnce(msg, os.Stdout)
}

var (
	versionPackage string = path.Base(os.Args[0])
	versionTag     string = "No Tag"
	versionHash    string = "No Hash"
	versionDate    string = "No Date"
)

var (
	O   *Options
	F   flip.Flip
	P   plugin.Loader
	ocm string
	CV  canvas.Canvas
)

func init() {
	O = defaultOptions()
	log.SetFormatter("warhola_text", log.MakeTextFormatter(versionPackage))
	F = flip.Base
	F.AddCommand("version", versionPackage, versionTag, versionHash, versionDate).
		AddCommand("help").
		SetGroup("top", -1, TopCommand(), StatusCommand())
}

// context passes the following
// 0 debug boolean
// 1 debug info
// 2 log.Logger
// 3 plugin loaders
// 4 canvas
func main() {
	pluginSetting(os.Args)
	c := context.Background()
	os.Exit(F.Execute(c, os.Args))
}
