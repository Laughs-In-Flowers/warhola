package plugin

import (
	"context"
	"fmt"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
)

type xrror struct {
	base string
	vals []interface{}
}

func (x *xrror) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(x.base, x.vals...))
}

func (x *xrror) Out(vals ...interface{}) *xrror {
	x.vals = vals
	return x
}

func Xrror(base string) *xrror {
	return &xrror{base: base}
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

func Log(c context.Context) log.Logger {
	l := c.Value(2)
	var ll log.Logger
	var ok bool
	if ll, ok = l.(log.Logger); ok {
		return ll
	}
	return nil
}

//func Plugins(c context.Context) *plugin.Loader {
//	return nil
//}

func Canvas(c context.Context) canvas.Canvas {
	cv := c.Value(4)
	var cvv canvas.Canvas
	var ok bool
	if cvv, ok = cv.(canvas.Canvas); ok {
		return cvv
	}
	return nil
}
