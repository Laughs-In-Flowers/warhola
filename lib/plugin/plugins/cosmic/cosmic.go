package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/ctx"
	"github.com/fatih/structs"
)

// https://www.shadertoy.com/view/XlfGRj
type Cosmic struct {
	name                                 string
	VolumeSteps, Iterations              int
	Zoom, Speed, Tile, Search            float64
	Magic, Brightness, Saturation        float64
	DarkMatter, DistanceFading, StepSize float64
	seedCoordinatesSet                   bool
	seedCoordinates                      [2]float64
	rotationSet                          bool
	Rotation1                            M2
	Rotation2                            M2
	directionSet                         bool
	Direction                            V3
	fromSet                              bool
	From                                 V3
	reuse                                bool
}

func newCosmic(
	name string,
	steps, iterations int,
	zoom,
	speed,
	tile,
	search,
	magic,
	brightness,
	saturation,
	dark,
	fading,
	step float64,
	reuse bool) Cosmic {
	return Cosmic{
		name,
		steps,
		iterations,
		zoom,
		speed,
		tile,
		search,
		magic,
		brightness,
		saturation,
		dark,
		fading,
		step,
		false,
		[2]float64{},
		false,
		M2{},
		M2{},
		false,
		V3{},
		false,
		V3{},
		reuse,
	}
}

var DefaultCosmic = newCosmic(
	"cosmic",
	20,
	17,
	0.8,
	0.010,
	0.85,
	10000000.0,
	0.53,
	0.0015,
	0.85,
	0.300,
	0.730,
	0.1,
	false,
)

func (c *Cosmic) Tag() string {
	return c.name
}

func (c *Cosmic) Apply(i canvas.Canvas) error {
	size := i.Bounds().Size()
	res := V2{float64(size.X), float64(size.Y)}
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			p := V2{float64(x), float64(y)}
			uv := V2{float64(x) / res.X, float64(y) / res.Y}
			cl := c.Shade(p, uv, res)
			i.Set(x, y, cl.RGBA64())
		}
	}
	if !c.reuse {
		c.Reset()
	}
	return nil
}

func (c *Cosmic) currentSeedCoords() (float64, float64) {
	if !c.seedCoordinatesSet {
		c.seedCoordinates[0] = float64(rand.Intn(int(c.Search)))
		c.seedCoordinates[1] = float64(rand.Intn(int(c.Search)))
		c.seedCoordinatesSet = true
	}
	return c.seedCoordinates[0], c.seedCoordinates[1]
}

func rot(v float64) M2 {
	return M2{math.Cos(v), math.Sin(v), -math.Sin(v), math.Cos(v)}
}

func (c *Cosmic) rotation(resolution V2) (M2, M2) {
	if !c.rotationSet {
		mx, my := c.currentSeedCoords()
		rot1 := rot(.5 + mx/resolution.X*2)
		rot2 := rot(.8 + my/resolution.Y*2)
		c.Rotation1 = rot1
		c.Rotation2 = rot2
	}
	return c.Rotation1, c.Rotation2
}

func (c *Cosmic) direction(uv V2, rot1, rot2 M2) V3 {
	if !c.directionSet {
		dir := V3{(uv.X * c.Zoom), (uv.Y * c.Zoom), 1}
		dr := rot1.Mul2x1(V2{dir.X, dir.Y})
		dr1 := rot2.Mul2x1(dr)
		c.Direction = V3{dr1.X, dr1.Y, dir.Z}
	}
	return c.Direction
}

func (c *Cosmic) from(rot1, rot2 M2) V3 {
	if !c.fromSet {
		from := V3{1, 0.5, 0.5}
		fr := rot1.Mul2x1(V2{from.X, from.Y})
		fr1 := rot2.Mul2x1(fr)
		c.From = V3{fr1.X, fr1.Y, from.Z}
	}
	return c.From
}

func (c *Cosmic) tileFold(in V3) V3 {
	t1 := V3{c.Tile, c.Tile, c.Tile}
	t2 := V3{math.Mod(in.X, c.Tile*2), math.Mod(in.Y, c.Tile*2), math.Mod(in.Z, c.Tile*2)}
	return V3{math.Abs(t1.X - t2.X), math.Abs(t1.Y - t2.Y), math.Abs(t1.Z - t2.Z)}
}

func lerp(a, b, p float64) float64 { return (1.0-p)*a + p*b }

func (c *Cosmic) Shade(p, uv, res V2) Color {
	uv.Y *= res.Y / res.X

	rot1, rot2 := c.rotation(res)
	dir := c.direction(uv, rot1, rot2)

	from := c.from(rot1, rot2)
	TIME := float64(time.Now().UTC().UnixNano())*c.Speed + 0.25
	from.Add(V3{TIME * 2, TIME, -2})

	s := float64(0.1)
	fade := float64(1)

	v := V3{}
	for r := 0; r < c.VolumeSteps; r++ {
		p := from.Add(dir.Scale(s * 0.5))
		p = c.tileFold(p)

		var pa, a float64
		for i := 0; i < c.Iterations; i++ {
			p = p.Abs().Scale(1.0 / p.Len2()).Offset(-c.Magic)
			a += math.Abs(p.Len() - pa)
		}

		dm := math.Max(0, c.DarkMatter-a*a*.001)

		a *= a * a

		if r > 6 {
			fade *= 1 - dm
		}

		v = v.Offset(fade)
		v = v.Add(V3{s, s * s, s * s * s * s}.Scale(a * fade * c.Brightness))

		fade *= c.DistanceFading
		s += c.StepSize
	}

	lv := v.Len()
	v = V3{lerp(lv, v.X, c.Saturation), lerp(lv, v.Y, c.Saturation), lerp(lv, v.Z, c.Saturation)}

	v = v.Scale(0.01)

	return C3(v.X, v.Y, v.Z)
}

func (c *Cosmic) Reset() {
	c.seedCoordinatesSet = false
	c.rotationSet = false
	c.directionSet = false
	c.fromSet = false
}

func Apply(c context.Context, l log.Logger, cc *Cosmic) context.Context {
	cv := ctx.Canvas(c)
	if cv != nil {
		if !cv.Noop() {
			err := cc.Apply(cv)
			if err != nil {
				cosmicFatal(l, err)
			}
			c = context.WithValue(c, 4, cv)
			cosmicLog(l, "ran")
			return c
		}
		cosmicLog(l, "image is noop")
	}
	cosmicLog(l, "there is no canvas")
	return c
}

func Debug(c context.Context, l log.Logger, cc *Cosmic) context.Context {
	d := c.Value(0)
	if dv, ok := d.(bool); ok {
		if dv {
			dm := DebugMap(c, cc)
			if dm != nil {
				l.Println("cosmic: wrote debug information")
				c = context.WithValue(c, 1, dm)
				return c
			}
		}
	}
	return c
}

func DebugMap(c context.Context, cc *Cosmic) map[string]string {
	di := c.Value(1)
	if dm, ok := di.(map[string]string); ok {
		return mergeDebugInfo(cc, dm)
	}
	return nil
}

func destructure(c *Cosmic) map[string]interface{} {
	return structs.Map(c)
}

func mapped(c *Cosmic) map[string]string {
	ret := make(map[string]string)
	for k, v := range destructure(c) {
		val := fmt.Sprintf("%v", v)
		ret[k] = val
	}
	return ret
}

func mergeDebugInfo(c *Cosmic, m map[string]string) map[string]string {
	v := mapped(c)
	for k, v := range v {
		m[k] = v
	}
	return m
}

var PluginName = "cosmic"

func cosmicFlags(c *Cosmic, fs *flip.FlagSet) *flip.FlagSet {
	fs.IntVar(&c.VolumeSteps, "steps", c.VolumeSteps, "")
	fs.IntVar(&c.Iterations, "iterations", c.VolumeSteps, "")
	fs.Float64Var(&c.Zoom, "zoom", c.Zoom, "")
	fs.Float64Var(&c.Speed, "speed", c.Speed, "")
	fs.Float64Var(&c.Tile, "tile", c.Tile, "")
	fs.Float64Var(&c.Search, "search", c.Search, "")
	fs.Float64Var(&c.Magic, "magic", c.Magic, "")
	fs.Float64Var(&c.Brightness, "brightness", c.Brightness, "")
	fs.Float64Var(&c.Saturation, "saturation", c.Saturation, "")
	fs.Float64Var(&c.DarkMatter, "darkMatter", c.DarkMatter, "")
	fs.Float64Var(&c.DistanceFading, "distanceFading", c.DistanceFading, "")
	fs.Float64Var(&c.StepSize, "stepSize", c.StepSize, "")
	return fs
}

func cosmicLog(l log.Logger, msg interface{}) {
	l.Printf("cosmic: %s", msg)
}

func cosmicFatal(l log.Logger, msg interface{}) {
	l.Fatalf("cosmic: %s", msg)
}

func Command() flip.Command {
	csmc := &DefaultCosmic
	fs := flip.NewFlagSet("top", flip.ContinueOnError)
	fs = cosmicFlags(csmc, fs)

	return flip.NewCommand(
		"",
		PluginName,
		"A starfield filter",
		100,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			l := ctx.Log(c)
			c = Apply(c, l, csmc)
			c = Debug(c, l, csmc)
			return c, flip.ExitSuccess
		},
		fs,
	)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
