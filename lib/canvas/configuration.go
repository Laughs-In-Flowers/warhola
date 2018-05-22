package canvas

import (
	"fmt"
	"image"
	"os"
	"sort"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/util/xrr"
)

type ConfigFn func(*canvas) error

type Config interface {
	Order() int
	Configure(*canvas) error
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

func (c config) Configure(v *canvas) error {
	return c.fn(v)
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
	v          *canvas
	configured bool
	list       configList
}

func newConfiguration(v *canvas, conf ...Config) *configuration {
	c := &configuration{
		v:    v,
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

func configure(v *canvas, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.v, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{0, setUp},

	config{1000, checkLogger},
	config{1001, checkDebug},
	config{1002, checkColorModel},
	config{1003, checkPath},
	config{1004, checkKind},
	config{1005, checkRect},
	config{1006, checkMeasure},
	config{1007, newAction},
	config{1008, checkAction},
	config{1009, action},
	config{1010, checkPalette},

	config{9999, tearDown},
}

type note struct {
	prefix string
	has    []string
}

func newNote(prefix string) *note {
	return &note{prefix, make([]string, 0)}
}

func (n *note) add(msg string, items ...interface{}) {
	act := fmt.Sprintf(msg, items...)
	res := fmt.Sprintf("%s %s", n.prefix, act)
	n.has = append(n.has, res)
}

func (n *note) addUn(msg string, items ...interface{}) {
	act := fmt.Sprintf(msg, items...)
	n.has = append(n.has, act)
}

var expected, results *note

func setUp(c *canvas) error {
	expected = newNote("canvas expected")
	results = newNote("canvas actual")
	return nil
}

func SetLogger(l log.Logger) Config {
	return NewConfig(-1,
		func(c *canvas) error {
			c.Logger = l
			return nil
		})
}

func checkLogger(c *canvas) error {
	if c.Logger == nil {
		c.Logger = log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
	}
	c.Print("canvas start")
	return nil
}

func SetDebug(d bool) Config {
	return NewConfig(1,
		func(c *canvas) error {
			c.debug = d
			return nil
		})
}

func checkDebug(c *canvas) error {
	expected.addUn("canvas debug is %t", c.debug)
	return nil
}

func SetColorModel(m string) Config {
	return NewConfig(2,
		func(c *canvas) error {
			c.pxl.m = stringToColorModel(m)
			return nil
		})
}

func checkColorModel(c *canvas) error {
	expected.add("color model is %s", c.pxl.m)
	return nil
}

func setPath(c *canvas, path string) {
	c.path = path
}

func SetPath(p string) Config {
	return NewConfig(3,
		func(c *canvas) error {
			setPath(c, p)
			return nil
		})
}

func checkPath(c *canvas) error {
	expected.add("path is %s", c.path)
	return nil
}

func SetKind(k string) Config {
	return NewConfig(4,
		func(c *canvas) error {
			c.kind = stringToKind(k)
			return nil
		})
}

func checkKind(c *canvas) error {
	expected.add("kind is %s", c.kind)
	return nil
}

func SetRect(x, y int) Config {
	return NewConfig(5,
		func(c *canvas) error {
			c.rect = image.Rect(0, 0, x, y)
			return nil
		})
}

var ZeroRectError = xrr.Xrror("rectangle is zero rectangle")

func checkRect(c *canvas) error {
	if c.pxl.rect == image.ZR {
		return ZeroRectError
	}
	expected.add("rectangle dimensions are min %v max %v", c.rect.Min, c.rect.Max)
	return nil
}

func SetMeasure(pp float64, ppu string) Config {
	return NewConfig(6,
		func(c *canvas) error {
			c.pxl.measure.SetPP(ppu, pp)
			return nil
		})
}

func checkMeasure(c *canvas) error {
	expected.addUn("canvas measure imperial: %F points per inch", c.ppi)
	expected.addUn("canvas measure metric: %F points per cm", c.ppc)
	return nil
}

func newAction(c *canvas) error {
	var act Action = ACTIONNOOP
	if c.path != PATHNOOP {
		_, err := os.Stat(c.path)
		switch {
		case os.IsNotExist(err):
			act = ACTIONNEW
		default:
			act = ACTIONOPEN
		}
	}
	c.action = act
	return nil
}

var noopError = xrr.Xrror("canvas is NOOP: %s").Out

func checkAction(c *canvas) error {
	if !c.Noop() {
		switch {
		case c.pxl.m == COLORNOOP:
			results.addUn("noop color is %s", COLORNOOP)
			return noopError(COLORNOOP)
		case c.path == PATHNOOP:
			results.addUn("noop path is %s", PATHNOOP)
			return noopError(PATHNOOP)
		case c.kind == KINDNOOP:
			results.addUn("noop kind is %s", KINDNOOP)
			return noopError(KINDNOOP)
		case c.action == ACTIONNOOP:
			results.add("noop action is %s", ACTIONNOOP)
			return noopError(ACTIONNOOP)
		}
	}
	expected.add("action: %s, %s, %s, %s", c.action, c.pxl.m, c.path, c.kind)
	return nil
}

func action(c *canvas) error {
	var err error
	var nk Kind
	var cm ColorModel
	switch c.action {
	case ACTIONNEW:
		err = newTo(c.pxl)
		nk, cm = c.kind, c.m
	case ACTIONOPEN:
		nk, cm, err = openTo(c.path, c.pxl)
		c.kind = nk
	default:
		err = noopError(ACTIONNOOP)
	}
	if err != nil {
		c.Printf("unable to perform action: %s", c.action)
	}
	results.add("color model is %s", cm)
	results.add("path is %s", c.path)
	results.add("kind is %s", nk)
	results.add("rectangle dimensions are min %v max %v", c.rect.Min, c.rect.Max)
	results.add("action: %s, %s, %s, %s", c.action, cm, c.path, nk)
	return err
}

func setPaletteFn(c *canvas, fn PaletteFunc) {
	c.pxl.paletteFn = fn
}

func checkPalette(c *canvas) error {
	if c.pxl.paletteFn == nil {
		expected.addUn("canvas palette func is default")
	}
	return nil
}

func SetPaletteFn(fn PaletteFunc) Config {
	return NewConfig(5,
		func(c *canvas) error {
			setPaletteFn(c, fn)
			expected.addUn("canvas set custom palette func")
			return nil
		})
}

func tearDown(c *canvas) error {
	if c.debug {
		for _, v := range expected.has {
			c.Print(v)
		}
		for _, v := range results.has {
			c.Print(v)
		}
	}
	expected, results = nil, nil
	return nil
}
