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
)

var pluginDirs = []string{"plugins"}

func dirs(paths string) []string {
	ret := pluginDirs
	if paths != "" {
		var pth string
		var err error
		spl := strings.Split(paths, ",")
		for _, v := range spl {
			pth, err = filepath.Abs(v)
			if err != nil {
				msg := new(bytes.Buffer)
				failure(msg, "plugin dirs", err)
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

	P, err = plugin.New(dirs(d)...)
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
	if Debug(c) {
		if l := Log(c); l != nil {
			l.Println("- start debug information -----")
			if d := DebugMap(c); d != nil {
				for k, v := range d {
					l.Printf("%s: %s", k, v)
				}
			}
			l.Println("- end debug information -----")
		}
	}
}

func Debug(c context.Context) bool {
	dv := c.Value(0)
	if d, ok := dv.(bool); ok {
		return d
	}
	return false
}

func DebugMap(c context.Context) map[string]string {
	di := c.Value(1)
	if div, ok := di.(map[string]string); ok {
		return div
	}
	return nil
}

func Log(c context.Context) log.Logger {
	l := c.Value(2)
	var ll log.Logger
	var ok bool
	if ll, ok = l.(log.Logger); ok {
		return ll
	}
	return nil
}

func logSetting(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	if o.formatter != "null" {
		switch o.formatter {
		case "raw":
			o.SwapFormatter(log.GetFormatter("raw"))
		case "text", "stdout":
			o.SwapFormatter(log.GetFormatter("warhola_text"))
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
	Debug           bool
	Color           string
	File, Extension string
	X, Y            int
	PP              int
	PPU             string
}

var defaultCanvasOptions = cOptions{
	false,
	"RGBA",
	canvas.DefaultPath,
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
	fs.StringVar(&o.Extension, "extension", o.Extension, "The file extension for the canvas. [png|jpeg]")
	fs.IntVar(&o.X, "X", o.X, "X dimension of the canvas.")
	fs.IntVar(&o.Y, "Y", o.Y, "Y dimension of the canvas.")
	fs.IntVar(&o.PP, "PP", o.PP, "points per unit where unit is specified in option PP")
	fs.StringVar(&o.PPU, "PPU", o.PPU, "unit of measurement for points per")
	return fs
}

func canvasSetting(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	CV = canvas.New(canvas.Options(o.Color,
		o.File,
		o.Extension,
		o.X,
		o.Y,
		o.PP,
		o.PPU,
	), o)
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
	if l := Log(c); l != nil {
		l.Println("cleaning up canvas")
		err := CV.Save()
		if err != nil {
			l.Printf("canvas cleanup error: %s", err)
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
	plugins bool
}

var defaultStatusOptions sOptions = sOptions{false}

func sFlags(fs *flip.FlagSet, o *sOptions) *flip.FlagSet {
	fs.BoolVar(&o.plugins, "pluginList", o.plugins, "List of all plugins by directory")
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
			case so.plugins:
				writePlugins(msg)
			}
			return c, flip.ExitSuccess
		},
		fs,
	)
}

func writeOnce(msg *bytes.Buffer, w io.Writer) {
	w.Write(msg.Bytes())
}

func failure(msg *bytes.Buffer, cause string, err error) flip.ExitStatus {
	msg.WriteString(fmt.Sprintf("%s: %s", cause, err))
	writeOnce(msg, os.Stderr)
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
		msg.WriteString(fmt.Sprintf("%s\n", k))
		for _, vv := range v {
			msg.WriteString(fmt.Sprintf("   %s\n", vv))
		}
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
	O  *Options
	F  flip.Flip
	P  plugin.Loader
	CV canvas.Canvas
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
