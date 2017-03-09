package factory

import (
	"fmt"
	"os"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/star"
)

type Factory struct {
	Configuration
	log.Logger
	canvas.Canvaser
	star.Loader
}

func New(cnf ...Config) *Factory {
	f := &Factory{}
	f.Configuration = newConfiguration(f, cnf...)
	err := f.Configure()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		f.Fatal(err)
	}
	return f
}

var Current *Factory

func init() {
	Current = New()
}
