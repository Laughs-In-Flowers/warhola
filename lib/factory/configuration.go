package factory

import (
	"os"
	"sort"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
)

type ConfigFn func(*Factory) error

type Config interface {
	Order() int
	Configure(*Factory) error
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

func (c config) Configure(f *Factory) error {
	return c.fn(f)
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
	f          *Factory
	configured bool
	list       configList
}

func newConfiguration(f *Factory, conf ...Config) *configuration {
	c := &configuration{
		f:    f,
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

func configure(f *Factory, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.f, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{1001, fLogger},
	config{1002, fCanvaser},
}

func fLogger(f *Factory) error {
	if f.Logger == nil {
		l := log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
		log.Current = l
		f.Logger = l
	}
	return nil
}

func fCanvaser(f *Factory) error {
	if f.Canvaser == nil {
		f.Canvaser = canvas.DefaultCanvaser
	}
	return nil
}

func SetLogger(l log.Logger) Config {
	return DefaultConfig(func(f *Factory) error {
		log.Current = l
		f.Logger = l
		return nil
	})
}
