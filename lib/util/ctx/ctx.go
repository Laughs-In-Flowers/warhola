package ctx

import (
	"context"
	"fmt"
	"strings"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
)

type Ctx struct {
	canvas.Canvas
	log.Logger
	Debug    bool
	DebugMap map[string]string
}

func Is(c context.Context) *Ctx {
	cv := Canvas(c)
	l := Log(c)
	d := Debug(c)
	dm := DebugMap(c)
	return &Ctx{
		cv, l, d, dm,
	}
}

func Debug(c context.Context) bool {
	dv := c.Value(0)
	if d, ok := dv.(bool); ok {
		return d
	}
	return false
}

func DebugMap(c context.Context) map[string]string {
	di := c.Value(1)
	if div, ok := di.(map[string]string); ok {
		return div
	}
	return nil
}

func DebugMapCollapse(m map[string]string) string {
	var inter []string
	for k, v := range m {
		inter = append(inter, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(inter, "\n")
}

func Log(c context.Context) log.Logger {
	l := c.Value(2)
	var ll log.Logger
	var ok bool
	if ll, ok = l.(log.Logger); ok {
		return ll
	}
	return nil
}

func Canvas(c context.Context) canvas.Canvas {
	cv := c.Value(4)
	var cvv canvas.Canvas
	var ok bool
	if cvv, ok = cv.(canvas.Canvas); ok {
		return cvv
	}
	return nil
}
