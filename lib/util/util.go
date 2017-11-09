package util

import (
	"image/color"
	"unicode"

	"github.com/Laughs-In-Flowers/log"
	"golang.org/x/image/math/fixed"
)

// Returns a color.Model from the provided string, defaulting to color.RGBAModel.
func StringToColorModel(s string) color.Model {
	switch s {
	case "gray":
		return color.GrayModel
	case "gray16":
		return color.Gray16Model
	case "alpha":
		return color.AlphaModel
	case "alpha16":
		return color.Alpha16Model
	case "rgba":
		return color.RGBAModel
	case "rgba64":
		return color.RGBA64Model
	case "nrgba":
		return color.NRGBAModel
	case "nrgba64":
		return color.NRGBA64Model
	case "cmyk":
		return color.CMYKModel
	}
	return color.RGBAModel
}

// Returns fixed.Point26_6
func Fixp(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{Fix(x), Fix(y)}
}

// Returns fixed.Int26_6
func Fix(x float64) fixed.Int26_6 {
	return fixed.Int26_6(x * 64)
}

func Success(l log.Logger, module, msg string) {
	l.Printf("%s: %s", module, msg)
}

func Failure(l log.Logger, module string, err error) {
	l.FatalErrorf("%s: %s", module, err)
}

type AlphaS []string

func (a AlphaS) Len() int      { return len(a) }
func (a AlphaS) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AlphaS) Less(i, j int) bool {
	iRunes := []rune(a[i])
	jRunes := []rune(a[j])

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}
	return false
}

//func Flatten(t draw.Interpolator, images ...draw.Image) draw.Image {
//	base := images[0]
//	over := 1
//	length := len(images)
//	for over <= length {
//		Combine(base, images[over], t)
//		over++
//	}
//	return base
//}

//var M3I = [6]float64{1, 0, 0, 1, 0, 0}

//func Combine(under draw.Image, over draw.Image, t draw.Interpolator) {
//	tr := f64.Aff3{M3I[0], M3I[1], M3I[4], M3I[2], M3I[3], M3I[5]}
//	t.Transform(
//		under,
//		tr,
//		over,
//		over.Bounds(),
//		draw.Over,
//		nil,
//	)
//}
