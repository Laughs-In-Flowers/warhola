package builtins

import (
	"context"
	"fmt"
	"sort"

	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/ctx"
)

type builtinsMap map[string]func() flip.Command

// Register a string tag corresponding to a function providing a flip.Command.
func (b builtinsMap) Register(tag string, fn func() flip.Command) {
	b[tag] = fn
}

// A string key map of functions returning a flip.Command.
var BuiltIns builtinsMap

type execution struct {
	pri int
	efn executionFn
}

type executionFn func(o *Options, c context.Context) (context.Context, flip.ExitStatus)

type executionGroup []execution

func (e executionGroup) Len() int { return len(e) }

func (e executionGroup) Less(i, j int) bool { return e[i].pri < e[j].pri }

func (e executionGroup) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

func execute(o *Options, c context.Context, a []string, e executionGroup) (context.Context, flip.ExitStatus) {
	sort.Sort(e)
	var status flip.ExitStatus
	for _, x := range e {
		c, status = x.efn(o, c)
		if status != flip.ExitNo {
			return c, status
		}
	}
	return c, status
}

func buildExecuting(debugKey string, fns ...execution) executionGroup {
	ret := make(executionGroup, 0)
	ret = append(ret, setLogStep)
	ret = append(ret, fns...)
	ret = append(ret, debugStep(debugKey))
	ret = append(ret, endStep)
	return ret
}

type innerExec func(*Options, canvas.Canvas) (canvas.Canvas, flip.ExitStatus)

func builtInMainStep(i innerExec) executionFn {
	return func(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
		cv := ctx.Canvas(c)
		cv, exit := i(o, cv)
		c = context.WithValue(c, 4, cv)
		return c, exit
	}

}

func builtInErrorHandler(o *Options, err error) flip.ExitStatus {
	if err != nil {
		o.Println(err)
		return flip.ExitFailure
	}
	return flip.ExitNo
}

var (
	setLogStep = execution{
		0,
		func(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
			o.Logger = ctx.Log(c)
			return c, flip.ExitNo
		},
	}

	debugStep = func(key string) execution {
		return execution{
			99,
			func(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
				if d := ctx.DebugMap(c); d != nil {
					ii := BuiltInsOptions.Match(key)
					for _, i := range ii {
						key := i.Key()
						val := fmt.Sprintf("%v", i.Provided())
						d[key] = val
					}
					c = context.WithValue(c, 1, d)
				}
				return c, flip.ExitNo
			},
		}
	}

	endStep = execution{
		100,
		func(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
			return c, flip.ExitSuccess
		},
	}
)

// A package wide Options struct consisting of a logger and a data.Vector.
type Options struct {
	log.Logger
	*data.Vector
}

// A package wide Options instance.
var BuiltInsOptions *Options

func init() {
	BuiltInsOptions = &Options{nil, data.New("builtin_options")}
	BuiltIns = make(builtinsMap)
	//adjustment
	BuiltIns.Register("adjust", adjustmentCommand)
	BuiltIns.Register("brightness", brightnessCommand)
	BuiltIns.Register("gamma", gammaCommand)
	BuiltIns.Register("contrast", contrastCommand)
	BuiltIns.Register("hue", hueCommand)
	BuiltIns.Register("saturation", saturationCommand)
	//blend

	//blur

	//channel

	//effect

	//histogram

	//noise

	//text
	BuiltIns.Register("text", textCommand)
	//transform
	BuiltIns.Register("crop", cropCommand)
	BuiltIns.Register("resize", resizeCommand)
	BuiltIns.Register("rotate", rotateCommand)
	BuiltIns.Register("flip", flipCommand)
	BuiltIns.Register("shear", shearCommand)
	BuiltIns.Register("translate", translateCommand)
}
