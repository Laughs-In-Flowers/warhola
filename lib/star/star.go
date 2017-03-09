package star

import (
	"os"
	"path/filepath"
	"plugin"
)

type Args struct {
	Path  string
	Args  []string
	Debug bool
}

func NewArgs(path string, debug bool, args ...string) *Args {
	return &Args{
		Path:  path,
		Args:  args,
		Debug: debug,
	}
}

type Star func(string, bool, ...string) error

type Loader interface {
	PluginDirectory() string
	Plugins() ([]string, error)
	Load() error
	Get(...string) ([]Star, error)
}

type loader struct {
	pluginsDir string
	plugins    map[string]Star
}

var PluginsDirCreateError = Xrror("could not create plugins dir: %v").Out

func defaultDir(tag string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", PluginsDirCreateError(err)
	}
	pDir := filepath.Join(wd, tag)
	_, err = os.Stat(pDir)
	if err != nil {
		err = os.Mkdir(pDir, 0755)
		if err != nil {
			return "", PluginsDirCreateError(err)
		}
	}
	return pDir, nil
}

var FileDoesNotExist = Xrror("file does not exist: %s").Out

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func New(pluginsDir string) (*loader, error) {
	if !fileExists(pluginsDir) {
		pDir, err := defaultDir("plugins")
		if err != nil {
			return nil, err
		}
		pluginsDir = pDir
	}

	return &loader{
		pluginsDir,
		make(map[string]Star),
	}, nil
}

func (l *loader) PluginDirectory() string {
	return l.pluginsDir
}

func (l *loader) Plugins() ([]string, error) {
	dir, err := os.Open(l.pluginsDir)
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
		srcPath = filepath.Join(l.PluginDirectory(), p)
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
	var value func(string, bool, ...string) error
	var ok bool

	if key, ok = u1.(*string); !ok {
		return OpenPluginError(path, "error with plugin name")
	}

	if value, ok = u2.(func(string, bool, ...string) error); !ok {
		return OpenPluginError(path, "error with plugin apply function")
	}

	l.plugins[*key] = value

	return nil
}

var StarDoesNotExistError = Xrror("star does not exist: %s").Out

func (l *loader) Get(tags ...string) ([]Star, error) {
	var ret []Star
	for _, t := range tags {
		var st Star
		var ok bool
		if st, ok = l.plugins[t]; !ok {
			return nil, StarDoesNotExistError(t)
		}
		ret = append(ret, st)
	}
	return ret, nil
}
