package builtins

import (
	"context"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
)

var (
	adjustmentExecuting = buildExecuting("adjust", adjustmentx)
	adjustmentx         = execution{50, adjustmentStep}
	adjustmentInner     = []innerExec{
		brightnessInner,
		gammaInner,
		contrastInner,
		hueInner,
		saturationInner,
	}
	adjustmentStep = builtInMainStep(func(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
		cv.Print("execute adjustments")
		var ret flip.ExitStatus
		for _, fn := range adjustmentInner {
			cv, ret = fn(o, cv)
			if ret != flip.ExitNo {
				return cv, ret
			}
		}
		return cv, ret
	})
)

func adjustmentFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("adjust", flip.ContinueOnError)
	fs.Float64Vector(v, "brightness", "adjust.brightness", "Amount of brightness adjustment to apply -1.0 to 1.0")
	fs.Float64Vector(v, "gamma", "adjust.gamma", "Amount of gamma adjustment to apply, greater than 0")
	fs.Float64Vector(v, "constrast", "adjust.contrast", "Amount of constrast adjustment to apply, -1.0 to 1.0")
	fs.Float64Vector(v, "hue", "adjust.hue", "Amount of hue adjustment to apply, -360 to 360")
	fs.Float64Vector(v, "saturation", "adjust.saturation", "Amount of saturation adjustment to apply,-1.0 to 1.0 (-1.0 being -100% and 1.0 being 100%)")
	return fs
}

func adjustmentCommand() flip.Command {
	return flip.NewCommand(
		"",
		"adjust",
		"Adjust the brightness,gamma,contrast,hue or saturation of an image",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, adjustmentExecuting)
		},
		adjustmentFlags(BuiltInsOptions),
	)
}

var (
	brightnessExecuting = buildExecuting("brightness", brightnessx)
	brightnessx         = execution{50, brightnessStep}
	brightnessStep      = builtInMainStep(brightnessInner)
)

func brightnessInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	cv.Print("execute brightness")

	brightness := o.ToFloat64("adjust.brightness")

	if brightness != 0 {
		err := cv.Adjust("brightness", brightness)
		return cv, builtInErrorHandler(o, err)
	}

	return cv, flip.ExitNo
}

func brightnessFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("brightness", flip.ContinueOnError)
	fs.Float64Vector(v, "adjust", "adjust.brightness", "Amount of brightness adjustment to apply -1.0 to 1.0")
	return fs
}

func brightnessCommand() flip.Command {
	return flip.NewCommand(
		"",
		"brightness",
		"Adjust the brightness of an image",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, brightnessExecuting)
		},
		brightnessFlags(BuiltInsOptions),
	)
}

var (
	gammaExecuting = buildExecuting("gamma", gammax)
	gammax         = execution{50, gammaStep}
	gammaStep      = builtInMainStep(gammaInner)
)

func gammaInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	cv.Print("execute gamma")
	gamma := o.ToFloat64("adjust.gamma")
	if gamma != 0 {
		err := cv.Adjust("gamma", gamma)
		return cv, builtInErrorHandler(o, err)
	}
	return cv, flip.ExitNo
}

func gammaFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("gamma", flip.ContinueOnError)
	fs.Float64Vector(v, "adjust", "adjust.gamma", "Amount of gamma adjustment to apply, greater than 0")
	return fs
}

func gammaCommand() flip.Command {
	return flip.NewCommand(
		"",
		"gamma",
		"Adjust image gamma",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, gammaExecuting)
		},
		gammaFlags(BuiltInsOptions),
	)
}

var (
	contrastExecuting = buildExecuting("contrast", contrastx)
	contrastx         = execution{50, contrastStep}
	contrastStep      = builtInMainStep(contrastInner)
)

func contrastInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	cv.Print("execute contrast")
	contrast := o.ToFloat64("adjust.contrast")
	if contrast != 0 {
		err := cv.Adjust("contrast", contrast)
		return cv, builtInErrorHandler(o, err)
	}

	return cv, flip.ExitNo
}

func contrastFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("constrast", flip.ContinueOnError)
	fs.Float64Vector(v, "adjust", "adjust.contrast", "Amount of constrast adjustment to apply, -1.0 to 1.0")
	return fs
}

func contrastCommand() flip.Command {
	return flip.NewCommand(
		"",
		"contrast",
		"Adjust image contrast",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, contrastExecuting)
		},
		contrastFlags(BuiltInsOptions),
	)
}

var (
	hueExecuting = buildExecuting("hue", huex)
	huex         = execution{50, hueStep}
	hueStep      = builtInMainStep(hueInner)
)

func hueInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	cv.Print("execute hue")
	hue := o.ToFloat64("adjust.hue")
	if hue != 0 {
		err := cv.Adjust("hue", hue)
		return cv, builtInErrorHandler(o, err)
	}

	return cv, flip.ExitNo
}

func hueFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("hue", flip.ContinueOnError)
	fs.Float64Vector(v, "adjust", "adjust.hue", "Amount of hue adjustment to apply, -360 to 360")
	return fs
}

func hueCommand() flip.Command {
	return flip.NewCommand(
		"",
		"hue",
		"Adjust image hue",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, hueExecuting)
		},
		hueFlags(BuiltInsOptions),
	)
}

var (
	saturationExecuting = buildExecuting("saturation", saturationx)
	saturationx         = execution{50, saturationStep}
	saturationStep      = builtInMainStep(saturationInner)
)

func saturationInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	cv.Print("execute saturation")
	saturation := o.ToFloat64("adjust.saturation")
	if saturation != 0 {
		err := cv.Adjust("saturation", saturation)
		return cv, builtInErrorHandler(o, err)
	}
	return cv, flip.ExitNo
}

func saturationFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("saturation", flip.ContinueOnError)
	fs.Float64Vector(v, "adjust", "adjust.saturation", "Amount of saturation adjustment to apply,-1.0 to 1.0 (-1.0 being -100% and 1.0 being 100%)")
	return fs
}

func saturationCommand() flip.Command {
	return flip.NewCommand(
		"",
		"saturation",
		"Adjust image saturation",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, saturationExecuting)
		},
		saturationFlags(BuiltInsOptions),
	)
}
