package factory

import (
	"fmt"
	"os"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/star"
)

// A context and utility environment for image manipulation.
type Factory struct {
	Configuration
	log.Logger
	canvas.Canvaser
	star.Loader
}

// Configures and produces a new *Factory with the provided Configuration.
func New(cnf ...Config) *Factory {
	f := &Factory{}
	f.Configuration = newConfiguration(f, cnf...)
	err := f.Configure()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal configuration error")
		fmt.Fprintln(os.Stderr, err)
		f.Fatal(err)
	}
	return f
}
