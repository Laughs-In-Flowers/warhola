package factory

import (
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
)

type Factory struct {
	Configuration
	log.Logger
	canvas.Canvaser
}

func New(cnf ...Config) *Factory {
	e := &Factory{}
	e.Configuration = newConfiguration(e, cnf...)
	e.Configure()
	return e
}

func (f *Factory) SetTextFormatter(tag, packageTag string) {
	log.SetFormatter(tag, log.MakeTextFormatter(packageTag))
	f.SwapFormatter(log.GetFormatter(tag))
}

var Current *Factory

func init() {
	Current = New()
}
