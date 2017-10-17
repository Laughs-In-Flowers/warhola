package wand

import (
	"fmt"
	"image"

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

func Debug(u draw.Image, t *Text) {
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
