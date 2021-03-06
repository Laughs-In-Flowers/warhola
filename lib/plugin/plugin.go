package plugin

import (
	"os"
	"path/filepath"
	p "plugin"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/core"
	"github.com/Laughs-In-Flowers/xrr"
)

type pluginCmd func() flip.Command

// An interface for plugin loading.
type Loader interface {
	AddDir(string) error
	Load() error
	Plugins() (map[string][]string, error)
	Get(...string) ([]flip.Command, error)
}

type loaders struct {
	has []Loader
}

// Provides a new, multiple directory handling Loader.
func New(dirs ...string) (*loaders, error) {
	def := make([]Loader, 0)
	def = append(def, Core)
	ret := &loaders{def}
	for _, d := range dirs {
		err := ret.AddDir(d)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// Adds a new directory to this *loaders instance.
func (l *loaders) AddDir(dir string) error {
	nl, err := newLoader(dir)
	if err != nil {
		return err
	}
	l.has = append(l.has, nl)
	return nil
}

// Loads plugins from all directories, returning any error.
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

// Provides a map[string][]string of all plugins managed by this *loaders instance
// and any error. The returned map is keyed by managed directory, and lists plugins
// in that directory.
func (l *loaders) Plugins() (map[string][]string, error) {
	ret := make(map[string][]string)
	for _, sl := range l.has {
		ps, _ := sl.Plugins()
		for k, v := range ps {
			ret[k] = v
		}
	}
	return ret, nil
}

// An error indicating the named plugin does not exist.
var PluginDoesNotExistError = xrr.Xrror("plugin does not exist: %s").Out

// Provided any number of string tags, returns an array of flip.Command and any error.
func (l *loaders) Get(tags ...string) ([]flip.Command, error) {
	var ret = make([]flip.Command, 0)
	var err error
	for _, ll := range l.has {
		var ps []flip.Command
		ps, err = ll.Get(tags...)
		ret = append(ret, ps...)
	}
	return ret, err
}

type loader struct {
	dir    string
	llfn   func(*loader) error
	plfn   func(*loader) (map[string][]string, error)
	loaded map[string]pluginCmd
}

func loaderDir(d string) string {
	pth, _ := filepath.Abs(d)
	return pth
}

func newLoader(dir string) (*loader, error) {
	return &loader{
		loaderDir(dir),
		defaultLoaderFunc,
		defaultPluginLister,
		nil,
	}, nil
}

// Does not add a directory, single directory is specified at instantiation.
func (l *loader) AddDir(string) error { return nil }

func defaultLoaderFunc(l *loader) error {
	if l.loaded == nil {
		l.loaded = make(map[string]pluginCmd)
	}
	var plugins map[string][]string
	var err error
	plugins, err = l.Plugins()
	if err != nil {
		return nil // pass through here and do nothing, its less mess
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
	p, err := p.Open(path)
	if err != nil {
		return OpenPluginError(path, err)
	}
	u1, err := p.Lookup("PluginName")
	if err != nil {
		return DoesntExistError(path, "name")
	}
	u2, err := p.Lookup("Command")
	if err != nil {
		return DoesntExistError(path, "command function")
	}

	var key *string
	var value func() flip.Command
	var ok bool

	if key, ok = u1.(*string); !ok {
		return OpenPluginError(path, "error with plugin name")
	}

	if value, ok = u2.(func() flip.Command); !ok {
		return OpenPluginError(path, "error with plugin command function")
	}

	l.loaded[*key] = value

	return nil
}

// Satisfies the interface Loader.Load function for this *loader
func (l *loader) Load() error {
	return l.llfn(l)
}

var (
	//
	OpenPluginError = xrr.Xrror("Unable to open plugin at %s:\n\t%s").Out
	//
	DoesntExistError = xrr.Xrror("Plugin at %s has no %s.").Out
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

// Satisfies the interface Loader.Plugins function for this *loader
func (l *loader) Plugins() (map[string][]string, error) {
	return l.plfn(l)
}

// Satisfies the interface Loader.Get function for this *loader
func (l *loader) Get(tags ...string) ([]flip.Command, error) {
	var ret = make([]flip.Command, 0)

	switch {
	case len(tags) > 0 && tags[0] == "ALL":
		for _, pc := range l.loaded {
			ret = append(ret, pc())
		}
		return ret, nil
	default:
		for _, tag := range tags {
			var pc pluginCmd
			var ok bool
			pc, ok = l.loaded[tag]
			switch {
			case ok:
				ret = append(ret, pc())
			case !ok:
				return nil, PluginDoesNotExistError(tag)
			}
		}
	}
	return ret, nil
}

// A customised always loaded Loader encapsulating core functionality
var Core *loader = &loader{
	"core",
	func(l *loader) error {
		if l.loaded == nil {
			l.loaded = make(map[string]pluginCmd)
			for k, fn := range core.Core {
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
		ret["core"] = res
		return ret, nil
	},
	nil,
}
