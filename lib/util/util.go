package util

/*
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
*/

// Returns fixed.Point26_6
//func Fixp(x, y float64) fixed.Point26_6 {
//	return fixed.Point26_6{Fix(x), Fix(y)}
//}

// Returns fixed.Int26_6
//func Fix(x float64) fixed.Int26_6 {
//	return fixed.Int26_6(x * 64)
//}

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
