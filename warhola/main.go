package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola"
)

const topUse = `Top level flag usage.`

func TopCommand() flip.Command {
	fs := flip.NewFlagSet("top", flip.ContinueOnError)

	eo := &defaultEnvOptions
	fs = envFlags(fs, eo)

	return flip.NewCommand(
		"",
		"warhola",
		topUse,
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			var ret flip.ExitStatus
			ret = envExecute(eo, c, a)
			return ret
		},
		fs,
	)
}

var (
	pkgVersion     *version
	versionPackage string = path.Base(os.Args[0])
	versionTag     string = "No Tag"
	versionHash    string = "No Hash"
	versionDate    string = "No Date"
	fullVersion    bool
)

type version struct {
	pkge, tag, hash, date string
}

func Version(p, t, h, d string) *version {
	return &version{p, t, h, d}
}

func (v *version) Default() string {
	return fmt.Sprintf("%s %s", v.pkge, v.tag)
}

func (v *version) Full() string {
	return fmt.Sprintf("%s %s(%s %s)", v.pkge, v.tag, v.hash, v.date)
}

func versionExecute(c context.Context, a []string) flip.ExitStatus {
	var p string
	switch {
	case fullVersion:
		p = pkgVersion.Full()
	default:
		p = pkgVersion.Default()
	}
	fmt.Println(p)
	return flip.ExitSuccess
}

const versionUse = `Print the package version and exit.`

func VersionCommand() flip.Command {
	fs := flip.NewFlagSet("version", flip.ContinueOnError)
	fs.BoolVar(&fullVersion, "full", false, "print full version information containing package name, tag, hash and date")
	return flip.NewCommand(
		"",
		"version",
		versionUse,
		2,
		versionExecute,
		fs,
	)
}

var (
	e *warhola.Env
	c *flip.Commander
)

func init() {
	log.SetFormatter("warhola_text", log.MakeTextFormatter(versionPackage))
	e = warhola.Current
	c = flip.Base
	pkgVersion = Version(versionPackage, versionTag, versionHash, versionDate)
	c.RegisterGroup("top", -1, TopCommand(), VersionCommand())
}

func main() {
	ctx := context.Background()
	c.Execute(ctx, os.Args)
	os.Exit(0)
}
