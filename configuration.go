package warhola

import (
	"os"
	"sort"

	"github.com/Laughs-In-Flowers/log"
)

type ConfigFn func(*Env) error

type Config interface {
	Order() int
	Configure(*Env) error
}

type config struct {
	order int
	fn    ConfigFn
}

func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

func (c config) Order() int {
	return c.order
}

func (c config) Configure(e *Env) error {
	return c.fn(e)
}

type configList []Config

func (c configList) Len() int {
	return len(c)
}

func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

type Configuration interface {
	Add(...Config)
	AddFn(...ConfigFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	e          *Env
	configured bool
	list       configList
}

func newConfiguration(e *Env, conf ...Config) *configuration {
	c := &configuration{
		e:    e,
		list: builtIns,
	}
	c.Add(conf...)
	return c
}

func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
}

func configure(e *Env, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.e, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{1001, eLogger},
	config{1002, eCanvaser},
}

func eLogger(e *Env) error {
	if e.Logger == nil {
		l := log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
		log.Current = l
		e.Logger = l
	}
	return nil
}

func eCanvaser(e *Env) error {
	if e.Canvaser == nil {
		e.Canvaser = DefaultCanvaser
	}
	return nil
}

func SetLogger(l log.Logger) Config {
	return DefaultConfig(func(e *Env) error {
		log.Current = l
		e.Logger = l
		return nil
	})
}
