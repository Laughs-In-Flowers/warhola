package star

import (
	"image/draw"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/Laughs-In-Flowers/warhola/lib/plugins/builtins"
)

type Star interface {
	Tag() string
	Path() string
	Debug() bool
	Args() []string
	Apply(draw.Image) (draw.Image, error)
}

type StarFunc func(draw.Image, bool, ...string) (draw.Image, error)

type arguments struct {
	pth string
	dbg bool
	arg []string
}

func NewArgs(path string, debug bool, args ...string) *arguments {
	return &arguments{path, debug, args}
}

func (a *arguments) Path() string { return a.pth }

func (a *arguments) Debug() bool { return a.dbg }

func (a *arguments) Args() []string { return a.arg }

type star struct {
	tag string
	*arguments
	fn StarFunc
}

func NewStar(tag string, s StarFunc, path string, debug bool, args ...string) *star {
	return &star{
		tag,
		NewArgs(path, debug, args...),
		s,
	}
}

func (s *star) Tag() string { return s.tag }

func (s *star) Apply(i draw.Image) (draw.Image, error) {
	return s.fn(i, s.Debug(), s.Args()...)
}

type Loader interface {
	AddDir(string) error
	Load() error
	Plugins() (map[string][]string, error)
	Get(string) (StarFunc, error)
	GetStars(string, bool, ...string) ([]Star, error)
}

type loaders struct {
	has []Loader
}

func New(dirs ...string) (*loaders, error) {
	def := make([]Loader, 0)
	def = append(def, BuiltIns)
	ret := &loaders{def}
	for _, d := range dirs {
		err := ret.AddDir(d)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (l *loaders) AddDir(dir string) error {
	nl := newLoader(dir)
	l.has = append(l.has, nl)
	return nil
}

func (l *loaders) Load() error {
	var err error
	for _, ld := range l.has {
		err = ld.Load()
		if err != nil {
			return err
		}
	}
	return err
}

func (l *loaders) Plugins() (map[string][]string, error) {
	ret := make(map[string][]string)
	for _, sl := range l.has {
		ps, err := sl.Plugins()
		if err != nil {
			return nil, err
		}
		for k, v := range ps {
			ret[k] = v
		}
	}
	return ret, nil
}

var StarDoesNotExistError = Xrror("star does not exist: %s").Out

func (l *loaders) Get(tag string) (StarFunc, error) {
	var st StarFunc
	for _, sl := range l.has {
		st, _ = sl.Get(tag)
		if st != nil {
			return st, nil
		}
	}
	return nil, StarDoesNotExistError(tag)
}

var ZeroLengthStarRequestError = Xrror("%v is not long enough to request a star").Out

func (l *loaders) GetStars(path string, debug bool, requests ...string) ([]Star, error) {
	var ret = make([]Star, 0)
	for _, req := range requests {
		params := strings.Split(req, "+")
		if len(params) < 1 {
			return nil, ZeroLengthStarRequestError(params)
		}
		stfn, err := l.Get(params[0])
		if err != nil {
			return nil, err
		}
		ns := NewStar(params[0], stfn, path, debug, params[0:]...)
		ret = append(ret, ns)
	}
	return ret, nil
}

type loader struct {
	dir    string
	llfn   func(*loader) error
	plfn   func(*loader) (map[string][]string, error)
	loaded map[string]StarFunc
}

func loaderDir(d string) string {
	_, err := os.Stat(d)
	if err != nil {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return d
}

func newLoader(dir string) *loader {
	return &loader{
		loaderDir(dir),
		defaultLoaderFunc,
		defaultPluginLister,
		nil,
	}
}

func (l *loader) AddDir(string) error { return nil }

func defaultLoaderFunc(l *loader) error {
	if l.loaded == nil {
		l.loaded = make(map[string]StarFunc)
	}
	var plugins map[string][]string
	var err error
	plugins, err = l.Plugins()
	if err != nil {
		return err
	}
	for _, v := range plugins {
		var srcPath string
		for _, plugin := range v {
			srcPath = filepath.Join(l.dir, plugin)
			err = loadPath(l, srcPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func loadPath(l *loader, path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return OpenPluginError(path, err)
	}
	u1, err := p.Lookup("PluginName")
	if err != nil {
		return DoesntExistError(path, "name")
	}
	u2, err := p.Lookup("Apply")
	if err != nil {
		return DoesntExistError(path, "apply function")
	}

	var key *string
	var value func(draw.Image, bool, ...string) (draw.Image, error)
	var ok bool

	if key, ok = u1.(*string); !ok {
		return OpenPluginError(path, "error with plugin name")
	}

	if value, ok = u2.(func(draw.Image, bool, ...string) (draw.Image, error)); !ok {
		return OpenPluginError(path, "error with plugin apply function")
	}

	l.loaded[*key] = value

	return nil
}

func (l *loader) Load() error {
	return l.llfn(l)
}

var (
	OpenPluginError  = Xrror("Unable to open plugin at %s: %s").Out
	DoesntExistError = Xrror("Plugin at %s has no %s.").Out
)

func defaultPluginLister(l *loader) (map[string][]string, error) {
	dir, err := os.Open(l.dir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	ret := make(map[string][]string)
	var res []string
	for _, name := range names {
		if filepath.Ext(name) == ".so" {
			res = append(res, name)
		}
	}
	ret[l.dir] = res
	return ret, nil
}

func (l *loader) Plugins() (map[string][]string, error) {
	return l.plfn(l)
}

func (l *loader) Get(tag string) (StarFunc, error) {
	if st, ok := l.loaded[tag]; ok {
		return st, nil
	}
	return nil, StarDoesNotExistError(tag)
}

var NotImplemented = Xrror("'%s' is not implemented").Out

func (l *loader) GetStars(string, bool, ...string) ([]Star, error) {
	return nil, NotImplemented("GetStars")
}

var BuiltIns *loader = &loader{
	"builtins",
	func(l *loader) error {
		if l.loaded == nil {
			l.loaded = make(map[string]StarFunc)
			for k, fn := range builtins.BuiltIns {
				l.loaded[k] = fn
			}
		}
		return nil
	},
	func(l *loader) (map[string][]string, error) {
		l.Load()
		ret := make(map[string][]string)
		var res []string
		for k, _ := range l.loaded {
			res = append(res, k)
		}
		ret["builtins"] = res
		return ret, nil
	},
	nil,
}
