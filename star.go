package warhola

import (
	"fmt"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Laughs-In-Flowers/plugin"
)

type Star interface {
	Tag() string
	Apply(*StarArgs) error
}

type StarArgs struct {
	Path  string
	Args  map[string]string
	Debug bool
}

func parseArgs(args ...string) map[string]string {
	m := make(map[string]string)
	for _, v := range args {
		spl := strings.Split(v, ":")
		if len(spl) == 2 {
			m[spl[0]] = spl[1]
		}
	}
	return m
}

func NewStarArgs(path string, debug bool, args ...string) *StarArgs {
	m := parseArgs(args...)
	return &StarArgs{
		Path:  path,
		Args:  m,
		Debug: debug,
	}
}

type StarResult struct {
	Exit  int
	Error error
}

type starPlugin struct {
	tag, path string
	*rpc.Client
}

var StarDoesNotExistError = Xrror("star does not exist: %s").Out

func Load(path string) (Star, error) {
	_, err := exec.LookPath(path)
	if err != nil {
		return nil, StarDoesNotExistError(err.Error())
	}
	tag := filepath.Base(path)
	return newStarPlugin(tag, path), nil
}

func newStarPlugin(tag, path string) *starPlugin {
	return &starPlugin{
		tag:  tag,
		path: path,
	}
}

func (s *starPlugin) Tag() string {
	return s.tag
}

func (s *starPlugin) fmtCall(c string) string {
	return fmt.Sprintf("%s.%s", s.tag, c)
}

func (s *starPlugin) Apply(args *StarArgs) error {
	if s.Client == nil {
		client, err := plugin.StartCodec(
			jsonrpc.NewClientCodec,
			os.Stderr,
			s.path,
		)
		if err != nil {
			return err
		}
		defer client.Close()
		s.Client = client
	}
	var result StarResult
	err := s.Call(s.fmtCall("Apply"), args, &result)
	if err != nil {
		return err
	}
	if result.Exit != 0 {
		log.Fatalf("warhola star %s result returned a non zero exit code: %d", s.tag, result.Exit)
	}
	if result.Error != nil {
		return result.Error
	}
	s.Client = nil
	return nil
}
