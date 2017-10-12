package star

import (
	"image/draw"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

type Req struct {
	*Args
	fn Star
}

func newReq(path string, debug bool, star Star, args ...string) *Req {
	a := NewArgs(path, debug, args...)
	return &Req{a, star}
}

func (r *Req) Apply(i draw.Image) (draw.Image, error) {
	return r.fn(i, r.Debug, r.Value...)
}

type Args struct {
	Path  string
	Debug bool
	Value []string
}

func NewArgs(path string, debug bool, args ...string) *Args {
	return &Args{
		Path:  path,
		Debug: debug,
		Value: args,
	}
}

type Star func(draw.Image, bool, ...string) (draw.Image, error)

type Loaders interface {
	AddStarDir(string) error
	Load() error
	Plugins() ([]string, error)
	Get(string) (Star, error)
	Request(string, bool, ...string) ([]*Req, error)
}

type loaders struct {
	has []Loader
}

func New(dirs ...string) (*loaders, error) {
	ret := &loaders{
		make([]Loader, 0),
	}
	for _, d := range dirs {
		err := ret.AddStarDir(d)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (l *loaders) AddStarDir(dir string) error {
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

func (l *loaders) Plugins() ([]string, error) {
	var ret []string
	for _, ld := range l.has {
		ps, err := ld.Plugins()
		if err != nil {
			return nil, err
		}
		ret = append(ret, ps...)
	}
	return ret, nil
}

var StarDoesNotExistError = Xrror("star does not exist: %s").Out

func (l *loaders) Get(tag string) (Star, error) {
	var st Star
	for _, ld := range l.has {
		st = ld.Get(tag)
		if st != nil {
			return st, nil
		}
	}
	return nil, StarDoesNotExistError(tag)
}

var ZeroLengthStarRequestError = Xrror("%v is not long enough to request a star").Out

func (l *loaders) Request(path string, debug bool, reqs ...string) ([]*Req, error) {
	var ret = make([]*Req, 0)
	for _, req := range reqs {
		params := strings.Split(req, "+")
		if len(params) < 1 {
			return nil, ZeroLengthStarRequestError(params)
		}
		st, err := l.Get(params[0])
		if err != nil {
			return nil, err
		}
		nr := newReq(path, debug, st, params[0:]...)
		ret = append(ret, nr)
	}
	return ret, nil
}

type Loader interface {
	Directory() string
	Plugins() ([]string, error)
	Load() error
	Get(string) Star
}

type loader struct {
	dir    string
	loaded map[string]Star
}

func starDir(d string) string {
	_, err := os.Stat(d)
	if err != nil {
		os.MkdirAll(d, 0755)
	}
	return d
}

func newLoader(dir string) *loader {
	return &loader{
		starDir(dir),
		make(map[string]Star),
	}
}

func (l *loader) Directory() string {
	return l.dir
}

func (l *loader) Plugins() ([]string, error) {
	dir, err := os.Open(l.dir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, name := range names {
		if filepath.Ext(name) == ".so" {
			res = append(res, name)
		}
	}
	return res, nil
}

func (l *loader) Load() error {
	var plugins []string
	var err error
	plugins, err = l.Plugins()
	if err != nil {
		return err
	}
	var srcPath string
	for _, p := range plugins {
		srcPath = filepath.Join(l.dir, p)
		err = load(l, srcPath)
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	OpenPluginError  = Xrror("Unable to open plugin at %s: %s").Out
	DoesntExistError = Xrror("Plugin at %s has no %s.").Out
)

func load(l *loader, path string) error {
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

func (l *loader) Get(tag string) Star {
	if st, ok := l.loaded[tag]; ok {
		return st
	}
	return nil
}
