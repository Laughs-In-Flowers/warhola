package main

import (
	"context"
	"os"
	"path"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/factory"
)

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

func tExecute(o *Options, c context.Context, a []string) flip.ExitStatus {
	for _, fn := range executing {
		fn(o)
	}
	return flip.ExitNo
}

type execution func(o *Options)

var executing = []execution{
	logSetting,
}

type tOptions struct {
	LogFormatter string
}

var defaultTopOptions tOptions = tOptions{"null"}

func tFlags(fs *flip.FlagSet, o *Options) *flip.FlagSet {
	fs.StringVar(&o.LogFormatter, "formatter", o.LogFormatter, "Specify the log formatter.")
	return fs
}

func logSetting(o *Options) {
	if o.LogFormatter != "null" {
		switch o.LogFormatter {
		case "raw":
			F.SwapFormatter(log.GetFormatter("raw"))
		case "text", "stdout":
			F.SwapFormatter(log.GetFormatter("warhola_text"))
		}
	}
}

func TopCommand() flip.Command {
	fs := flip.NewFlagSet("top", flip.ContinueOnError)
	fs = tFlags(fs, O)

	return flip.NewCommand(
		"",
		"warhola",
		"Top level flag usage.",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			return tExecute(O, c, a)
		},
		fs,
	)
}

var (
	versionPackage string = path.Base(os.Args[0])
	versionTag     string = "No Tag"
	versionHash    string = "No Hash"
	versionDate    string = "No Date"
)

var (
	O *Options
	F *factory.Factory
	C *flip.Commander
)

func init() {
	O = defaultOptions()
	log.SetFormatter("warhola_text", log.MakeTextFormatter(versionPackage))
	F = factory.Current
	C = flip.BaseWithVersion(versionPackage, versionTag, versionHash, versionDate)
	C.RegisterGroup("top", -1, TopCommand())
	C.RegisterGroup("canvas", 10, CanvasCommand())
}

func main() {
	ctx := context.Background()
	C.Execute(ctx, os.Args)
	os.Exit(0)
}
