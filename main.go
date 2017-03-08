package main

import (
	"context"
	"os"
	"path"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/factory"
)

func fExecute(o *fOptions, c context.Context, a []string) flip.ExitStatus {
	for _, fn := range executing {
		fn(o)
	}
	return flip.ExitNo
}

type execution func(o *fOptions)

var executing = []execution{
	logSetting,
}

type fOptions struct {
	LogFormatter string
}

var defaultOptions fOptions = fOptions{"null"}

func fFlags(fs *flip.FlagSet, o *fOptions) *flip.FlagSet {
	fs.StringVar(&o.LogFormatter, "formatter", o.LogFormatter, "Specify the log formatter.")
	return fs
}

func logSetting(o *fOptions) {
	if o.LogFormatter != "null" {
		switch o.LogFormatter {
		case "text", "stdout":
			f.SetTextFormatter("warhola_text", versionPackage)
		}
	}
}

func TopCommand() flip.Command {
	fs := flip.NewFlagSet("top", flip.ContinueOnError)

	eo := &defaultOptions
	fs = fFlags(fs, eo)

	return flip.NewCommand(
		"",
		"warhola",
		"Top level flag usage.",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			var ret flip.ExitStatus
			ret = fExecute(eo, c, a)
			return ret
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
	f *factory.Factory
	c *flip.Commander
)

func init() {
	f = factory.Current
	c = flip.BaseWithVersion(versionPackage, versionTag, versionHash, versionDate)
	c.RegisterGroup("top", -1, TopCommand())
}

func main() {
	ctx := context.Background()
	c.Execute(ctx, os.Args)
	os.Exit(0)
}
