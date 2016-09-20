package warhola

import (
	"github.com/Laughs-In-Flowers/log"
)

type Env struct {
	Configuration
	log.Logger
	Canvaser
}

func NewEnv(cnf ...Config) *Env {
	e := &Env{}
	e.Configuration = newConfiguration(e, cnf...)
	e.Configure()
	return e
}

type Factory interface {
	Canvaser
}

type factory struct {
	*Env
}

func NewFactory(c ...Config) Factory {
	return &factory{NewEnv(c...)}
}

var Current *Env

func init() {
	Current = NewEnv()
}
