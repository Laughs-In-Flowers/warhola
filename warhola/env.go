package main

import (
	"context"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
)

func envExecute(o *envOptions, c context.Context, a []string) flip.ExitStatus {
	for _, fn := range executing {
		fn(o)
	}
	return flip.ExitNo
}

type execution func(o *envOptions)

var executing = []execution{
	logLoading,
}

type envOptions struct {
	LogFormatter string
}

var defaultEnvOptions envOptions = envOptions{"null"}

func envFlags(fs *flip.FlagSet, o *envOptions) *flip.FlagSet {
	fs.StringVar(&o.LogFormatter, "formatter", o.LogFormatter, "Specify the log formatter.")
	return fs
}

func logLoading(o *envOptions) {
	if o.LogFormatter != "null" {
		switch o.LogFormatter {
		case "text", "stdout":
			e.SwapFormatter(log.GetFormatter("warhola_text"))
		}
		e.Printf("swapping to log formatter: %s", o.LogFormatter)
	}
}
