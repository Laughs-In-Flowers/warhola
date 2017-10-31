package builtins

import (
	"context"
	"fmt"
	"image"

	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/math/fixed"
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

func Canvas(c context.Context) canvas.Canvas {
	cv := c.Value(4)
	var cvv canvas.Canvas
	var ok bool
	if cvv, ok = cv.(canvas.Canvas); ok {
		return cvv
	}
	return nil
}

func Flatten(t draw.Interpolator, images ...draw.Image) draw.Image {
	base := images[0]
	over := 1
	length := len(images)
	for over <= length {
		Combine(base, images[over], t)
		over++
	}
	return base
}

type M3 [6]float64

func M3IdentityMatrix() M3 {
	return M3{1, 0, 0, 1, 0, 0}
}

func Combine(under draw.Image, over draw.Image, t draw.Interpolator) {
	tr := M3IdentityMatrix()
	tra := f64.Aff3{tr[0], tr[1], tr[4], tr[2], tr[3], tr[5]}
	t.Transform(
		under,
		tra,
		over,
		over.Bounds(),
		draw.Over,
		nil,
	)
}

func DebugText(u draw.Image, t *Text) {
	b := u.Bounds()

	o := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	t.DrawStringWrapped(o, 0, 0, float64(b.Dx()/2), 1.5, AlignLeft)

	Combine(u, o, draw.NearestNeighbor)
}

func fixp(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{fix(x), fix(y)}
}

func fix(x float64) fixed.Int26_6 {
	return fixed.Int26_6(x * 64)
}
