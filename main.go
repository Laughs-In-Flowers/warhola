package main

import (
	"context"
	"os"
	"path"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/factory"
)

func tExecute(o *tOptions, c context.Context, a []string) flip.ExitStatus {
	for _, fn := range executing {
		fn(o)
	}
	return flip.ExitNo
}

type execution func(o *tOptions)

var executing = []execution{
	logSetting,
}

type tOptions struct {
	LogFormatter string
}

var defaultOptions tOptions = tOptions{"null"}

func tFlags(fs *flip.FlagSet, o *tOptions) *flip.FlagSet {
	fs.StringVar(&o.LogFormatter, "formatter", o.LogFormatter, "Specify the log formatter.")
	return fs
}

func logSetting(o *tOptions) {
	if o.LogFormatter != "null" {
		switch o.LogFormatter {
		case "text", "stdout":
			F.SwapFormatter(log.GetFormatter("warhola_text"))
		}
	}
}

func TopCommand() flip.Command {
	fs := flip.NewFlagSet("top", flip.ContinueOnError)

	eo := &defaultOptions
	fs = tFlags(fs, eo)

	return flip.NewCommand(
		"",
		"warhola",
		"Top level flag usage.",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			return tExecute(eo, c, a)
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
	F *factory.Factory
	C *flip.Commander
)

func init() {
	log.SetFormatter("warhola_text", log.MakeTextFormatter(versionPackage))
	F = factory.Current
	C = flip.BaseWithVersion(versionPackage, versionTag, versionHash, versionDate)
	C.RegisterGroup("top", -1, TopCommand())
}

func main() {
	ctx := context.Background()
	C.Execute(ctx, os.Args)
	os.Exit(0)
}
