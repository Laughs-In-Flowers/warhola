package main

import (
	"context"
	"fmt"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
)

type sOptions struct {
	plugins bool
}

var defaultStatusOptions sOptions = sOptions{false}

func sFlags(fs *flip.FlagSet, o *Options) *flip.FlagSet {
	fs.BoolVar(&o.plugins, "plugins", o.plugins, "List all plugins")
	return fs
}

func StatusCommand() flip.Command {
	fs := flip.NewFlagSet("top", flip.ContinueOnError)
	fs = sFlags(fs, O)

	return flip.NewCommand(
		"",
		"status",
		"warhola status information",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			O.SwapFormatter(log.GetFormatter("raw"))
			switch {
			case O.plugins:
				ps, err := F.Plugins()
				if err != nil {
					return flip.ExitFailure
				}
				for k, v := range ps {
					O.Println(k)
					for _, vv := range v {
						O.Println(fmt.Sprintf("   %s", vv))
					}
				}
			}
			return flip.ExitSuccess
		},
		fs,
	)
}
