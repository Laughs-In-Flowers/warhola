package geo

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/xrr"
)

type Geometry struct {
	Raw              string
	ScaleX, ScaleY   float64
	X, OffsetX       int
	Y, OffsetY       int
	AspectX, AspectY float64
	Area             int
	Gravity          Gravity
	Error            []error
}

func New(s string) *Geometry {
	g := defaultGeometry()
	if s != "" {
		return parseGeometry(s, g)
	}
	return g
}

func defaultGeometry() *Geometry {
	return &Geometry{"", 0, 0, 0, 0, 0, 0, 1, 1, 0, Center, make([]error, 0)}
}

func parseGeometry(s string, g *Geometry) *Geometry {
	g.Raw = s
	for _, v := range grxs {
		if v.MatchString(s) {
			err := v.call(s, g)
			if err != nil {
				g.Error = append(g.Error, parseGeometryError(s, err))
			}
		}
	}
	return g
}

type rx2geo func(*regexp.Regexp, string, *Geometry) error

type grx struct {
	gfn rx2geo
	*regexp.Regexp
}

func newGrx(gfn rx2geo, rfn *regexp.Regexp) grx {
	return grx{gfn, rfn}
}

var emptyRxError = xrr.Xrror("regex '%s' returned nothing").Out

func (g grx) call(s string, gg *Geometry) error {
	var rs string
	rs = g.FindString(s)
	if rs != "" {
		return g.gfn(g.Regexp, rs, gg)
	}
	return emptyRxError(s)
}

func paramsInt(r *regexp.Regexp, s string) (map[string]int, error) {
	ret := make(map[string]int)
	match := r.FindStringSubmatch(s)
	for i, name := range r.SubexpNames() {
		if i > 0 && i <= len(match) {
			v, err := strconv.Atoi(match[i])
			if err == nil {
				ret[name] = v
			}

		}
	}
	return ret, nil
}

func paramsFloat(r *regexp.Regexp, s string) (map[string]float64, error) {
	ret := make(map[string]float64)
	match := r.FindStringSubmatch(s)
	for i, name := range r.SubexpNames() {
		if i > 0 && i <= len(match) {
			v, err := strconv.ParseFloat(match[i], 64)
			if err == nil {
				ret[name] = v
			}
		}
	}
	return ret, nil
}

var (
	area = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			vals, err := paramsInt(r, s)
			if err != nil {
				return err
			}
			var ok bool
			if g.Area, ok = vals["area"]; !ok {
				return emptyRxError(s)
			}
			return nil
		},
		regexp.MustCompile("(?P<area>[0-9]+)[@]"),
	)

	aspect = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			vals, err := paramsFloat(r, s)
			if err != nil {
				return err
			}
			var ok bool
			if g.AspectX, ok = vals["aspectX"]; !ok {
				return emptyRxError(s)
			}
			if g.AspectY, ok = vals["aspectY"]; !ok {
				return emptyRxError(s)
			}
			return nil
		},
		regexp.MustCompile("(?P<aspectX>[0-9]*\\.?[0-9]+)[:](?P<aspectY>[0-9]*\\.?[0-9]+)"),
	)

	gravityError = xrr.Xrror("%s is not a gravity specification").Out

	gravity = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			var err error
			g.Gravity = stringToGravity(s)
			if g.Gravity == GravityNo {
				err = gravityError(s)
			}
			return err
		},
		regexp.MustCompile("(?i)(?P<gravity>NorthWest|NorthEast|North|West|Center|East|SouthWest|SouthEast|South)"),
	)

	height = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			vals, err := paramsInt(r, s)
			if err != nil {
				return err
			}
			var ok bool
			if g.Y, ok = vals["y"]; !ok {
				return emptyRxError(s)
			}
			return nil
		},
		regexp.MustCompile("(?i)(height|y)[:](?P<y>[0-9]+)"),
	)

	offsetrx2geo = func(r *regexp.Regexp, s string, g *Geometry) error {
		vals, err := paramsInt(r, s)
		if err != nil {
			return err
		}
		var ok bool
		if g.OffsetX, ok = vals["offsetX"]; !ok {
			return emptyRxError(s)
		}
		if g.OffsetY, ok = vals["offsetY"]; !ok {
			return emptyRxError(s)
		}
		return nil
	}

	offset1 = newGrx(
		offsetrx2geo,
		regexp.MustCompile("[{](?P<offsetX>[+|-]?[0-9]+)[/](?P<offsetY>[+|-]?[0-9]+)[}]"),
	)

	offset2 = newGrx(
		offsetrx2geo,
		regexp.MustCompile("(?i)(offset)[:](?P<offsetX>[+|-]?[0-9]+)[,](?P<offsetY>[+|-]?[0-9]+)"),
	)

	rect = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			vals, err := paramsInt(r, s)
			if err != nil {
				return err
			}
			var ok bool
			if g.OffsetX, ok = vals["minX"]; !ok {
				return emptyRxError(s)
			}
			if g.OffsetY, ok = vals["minY"]; !ok {
				return emptyRxError(s)
			}
			if g.X, ok = vals["maxX"]; !ok {
				return emptyRxError(s)
			}
			if g.Y, ok = vals["maxY"]; !ok {
				return emptyRxError(s)
			}
			return nil
		},
		regexp.MustCompile("(?i)((rect)[:]|\\A)(?P<rect>(?P<minX>[0-9]+),(?P<minY>[0-9]+),(?P<maxX>[0-9]+),(?P<maxY>[0-9]+))"),
	)

	scale = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			vals, err := paramsFloat(r, s)
			if err != nil {
				return err
			}
			var sc float64
			var ok bool
			if sc, ok = vals["scale"]; !ok {
				return emptyRxError(s)
			}
			g.ScaleX, g.ScaleY = sc, sc
			return nil
		},
		regexp.MustCompile("(?P<scale>[-]?[0-9]+)[%]"),
	)

	sxs = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			vals, err := paramsFloat(r, s)
			if err != nil {
				return err
			}
			var ok bool
			if g.ScaleX, ok = vals["scaleX"]; !ok {
				return emptyRxError(s)
			}
			if g.ScaleY, ok = vals["scaleY"]; !ok {
				return emptyRxError(s)
			}
			return nil

		},
		regexp.MustCompile("(?P<scaleX>[0-9]+)[%](?P<scaleY>[0-9]+)"),
	)

	wxh = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			vals, err := paramsInt(r, s)
			if err != nil {
				return err
			}
			var ok bool
			if g.X, ok = vals["x"]; !ok {
				return emptyRxError(s)
			}
			if g.Y, ok = vals["y"]; !ok {
				return emptyRxError(s)
			}
			return nil
		},
		regexp.MustCompile("(?P<x>[0-9]+)[x](?P<y>[0-9]+)"),
	)

	width = newGrx(
		func(r *regexp.Regexp, s string, g *Geometry) error {
			vals, err := paramsInt(r, s)
			if err != nil {
				return err
			}
			var ok bool
			if g.X, ok = vals["x"]; !ok {
				return emptyRxError(s)
			}
			return nil
		},
		regexp.MustCompile("(?i)(width|x):(?P<x>[0-9]+)"),
	)

	grxs = []grx{area, aspect, gravity, height, offset1, offset2, rect, scale, sxs, wxh, width}

	parseGeometryError = xrr.Xrror("error parsing '%s' as Geometry: %s").Out
)

type Gravity int

func stringToGravity(s string) Gravity {
	switch strings.ToLower(s) {
	case "northwest":
		return NorthWest
	case "north":
		return North
	case "northeast":
		return NorthEast
	case "west":
		return West
	case "center":
		return Center
	case "east":
		return East
	case "southwest":
		return SouthWest
	case "south":
		return South
	case "southeast":
		return SouthEast
	}
	return GravityNo
}

const (
	GravityNo Gravity = iota
	NorthWest
	North
	NorthEast
	West
	Center
	East
	SouthWest
	South
	SouthEast
)

func GeometryFlag(fs *flip.FlagSet, val *string, dflt string) {
	fs.StringVar(val, "geometry", dflt, GeometryInstruction)
}

func GeometryVectorFlag(fs *flip.FlagSet, v *data.Vector, key string) {
	fs.StringVectorVar(v, "geometry", key, "", GeometryInstruction)
}

var GeometryInstruction string = "set geometry by string"
