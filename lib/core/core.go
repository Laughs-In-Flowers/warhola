package core

import (
	"context"
	"fmt"
	"sort"

	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/ctx"
	"github.com/Laughs-In-Flowers/warhola/lib/util/geo"
)

// A string key map of functions returning a flip.Command.
var Core cmdMap

type cmdMap map[string]func() flip.Command

// Register a string tag corresponding to a function providing a flip.Command.
func (c cmdMap) Register(tag string, fn func() flip.Command) {
	c[tag] = fn
}

// Register a function that passes in the cmdMap instance to the func.
func (c cmdMap) RegisterFunc(fn func(cmdMap)) {
	fn(c)
}

func execute(o *Options, c context.Context, a []string, e executionGroup) (context.Context, flip.ExitStatus) {
	e.sort()
	var status flip.ExitStatus
	for _, x := range e {
		c, status = x.efn(o, c)
		if status != flip.ExitNo {
			return c, status
		}
	}
	return c, status
}

type (
	execution struct {
		pri int
		efn executionFn
	}

	executionFn func(o *Options, c context.Context) (context.Context, flip.ExitStatus)

	executionGroup []execution
)

func (e executionGroup) Len() int { return len(e) }

func (e executionGroup) Less(i, j int) bool { return e[i].pri < e[j].pri }

func (e executionGroup) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

func (e executionGroup) sort() { sort.Sort(e) }

func defaultExecutionGroup() executionGroup {
	ret := make(executionGroup, 0)
	ret = append(ret, setLogStep, endStep)
	return ret
}

func buildExec(debugKey string, fns ...execution) executionGroup {
	ret := defaultExecutionGroup()
	ret = append(ret, fns...)
	ret = append(ret, debugStep(debugKey))
	return ret
}

func coreExec(e ...innerExec) executionGroup {
	var ret []execution
	for i, v := range e {
		ret = append(ret, execution{50 + i, coreMainStep(v)})
	}
	return ret
}

type innerExec func(*Options, canvas.Canvas) (canvas.Canvas, flip.ExitStatus)

func coreMainStep(ie innerExec) executionFn {
	return func(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
		cv := ctx.Canvas(c)
		cv, exit := ie(o, cv)
		c = context.WithValue(c, 4, cv)
		return c, exit
	}
}

var (
	setLogStep = execution{
		0,
		func(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
			o.Logger = ctx.Log(c)
			return c, flip.ExitNo
		},
	}

	// Nothing uses this, yet could it be useful (or should it be removed)?
	//setCtxOptsStep = execution{
	//	1,
	//	func(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	//		// 'global' stuff from main in ctx to make 'local'
	//		return c, flip.ExitNo
	//	},
	//}

	debugStep = func(key string) execution {
		return execution{
			99,
			func(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
				if d := ctx.DebugMap(c); d != nil {
					ii := CoreOptions.Match(key)
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

type (
	command struct {
		group, tag, instruction string
		priority                int
		x                       []execution
		ffn                     flagSetFunc
		cfn                     cmdFunc
	}

	flagSetFunc func(*Options) *flip.FlagSet

	cmdFunc func(*command) flip.Command
)

func NewCommand(group, tag, instruction string,
	prio int,
	ffn flagSetFunc,
	cfn cmdFunc,
	x ...execution) *command {
	return &command{
		group,
		tag,
		instruction,
		prio,
		x,
		ffn,
		cfn,
	}
}

func (c *command) executing() executionGroup {
	return buildExec(c.tag, c.x...)
}

func (c *command) setExecute(x ...execution) {
	c.x = x
}

func defaultCommandFunc(c *command) flip.Command {
	return flip.NewCommand(
		c.group,
		c.tag,
		c.instruction,
		c.priority,
		false,
		func(ctx context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(CoreOptions, ctx, a, c.executing())
		},
		c.ffn(CoreOptions),
	)
}

func (c *command) Command() flip.Command {
	return c.cfn(c)
}

// A package level Options struct consisting of a logger and a data.Vector.
type Options struct {
	log.Logger
	*data.Vector
}

func (o *Options) pullGeometry(keys ...string) *geo.Geometry {
	var g string = o.ToString("default.geometry")
	for _, k := range keys {
		v := o.ToString(k)
		if v != "" {
			g = v
			break
		}
	}
	return geo.New(g)
}

func coreErrorHandler(o *Options, err error) flip.ExitStatus {
	if err != nil {
		o.Println(err)
		return flip.ExitFailure
	}
	return flip.ExitNo
}

// A package level Options instance.
var CoreOptions *Options

func init() {
	initCoreFonts()
	CoreOptions = &Options{nil, data.New("core_options")}
	Core = make(cmdMap)
	//adjustment
	Core.Register("adjust", adjust)
	//blend
	Core.Register("blend", blend)
	//blur
	Core.Register("blur", blur)
	//channel
	//BuitIns.Register()
	//convolute
	Core.Register("convolve", convolve)
	//effect
	//BuiltIns.RegisterFunc()
	//histogram
	//BuiltIns.Register()
	//noise
	Core.Register("noise", noise)
	//text
	Core.Register("text", text)
	//transform
	Core.RegisterFunc(registerTransformCmds)
	//translate
	Core.RegisterFunc(registerTranslateCmds)
}
