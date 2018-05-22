package canvas

import (
	"image/color"
	"math"
)

type ResampleFilterFunc func(float64) float64

type ResampleFilter struct {
	Support float64
	Fn      ResampleFilterFunc
}

func stringToFilter(s string) ResampleFilter {
	switch s {
	case "box":
		return Box
	case "linear":
		return Linear
	case "gaussian":
		return Gaussian
	case "mitchellnetravali":
		return MitchellNetravali
	case "catmullrom":
		return CatmullRom
	case "lanczos":
		return Lanczos
	case "bartlett":
		return Bartlett
	case "hermite":
		return Hermite
	case "bspline":
		return BSpline
	case "hann":
		return Hann
	case "hamming":
		return Hamming
	case "blackman":
		return Blackman
	case "welch":
		return Welch
	case "cosine":
		return Cosine
	}
	return NearestNeighbor
}

var (
	nearest = func(src *canvas, w, h int) {
		srcP := src.pxl.clone(color.RGBAModel)
		srcW, srcH := srcP.Bounds().Dx(), srcP.Bounds().Dy()
		srcStride := srcP.str

		dstP := scratch(srcP, srcP.ColorModel(), w, h)
		dstStride := dstP.str

		dx := float64(srcW) / float64(w)
		dy := float64(srcH) / float64(h)

		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				pos := y*dstStride + x*4
				ipos := int((float64(y)+0.5)*dy)*srcStride + int((float64(x)+0.5)*dx)*4

				dstP.pix[pos+0] = srcP.pix[ipos+0]
				dstP.pix[pos+1] = srcP.pix[ipos+1]
				dstP.pix[pos+2] = srcP.pix[ipos+2]
				dstP.pix[pos+3] = srcP.pix[ipos+3]
			}
		}

		src.pxl = dstP.clone(src.ColorModel())
	}

	NearestNeighbor = ResampleFilter{
		Support: 0,
		Fn:      nil,
	}

	Box = ResampleFilter{
		Support: 0.5,
		Fn: func(x float64) float64 {
			if math.Abs(x) < 0.5 {
				return 1
			}
			return 0
		},
	}

	Linear = ResampleFilter{
		Support: 1.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 1.0 {
				return 1.0 - x
			}
			return 0
		},
	}

	Gaussian = ResampleFilter{
		Support: 1.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 1.0 {
				exp := 2.0
				x *= 2.0
				y := math.Pow(0.5, math.Pow(x, exp))
				base := math.Pow(0.5, math.Pow(2, exp))
				return (y - base) / (1 - base)
			}
			return 0
		},
	}

	MitchellNetravali = ResampleFilter{
		Support: 2.0,
		Fn: func(x float64) float64 {
			b := 1.0 / 3
			c := 1.0 / 3
			var w [4]float64
			x = math.Abs(x)

			if x < 1.0 {
				w[0] = 0
				w[1] = 6 - 2*b
				w[2] = (-18 + 12*b + 6*c) * x * x
				w[3] = (12 - 9*b - 6*c) * x * x * x
			} else if x <= 2.0 {
				w[0] = 8*b + 24*c
				w[1] = (-12*b - 48*c) * x
				w[2] = (6*b + 30*c) * x * x
				w[3] = (-b - 6*c) * x * x * x
			} else {
				return 0
			}

			return (w[0] + w[1] + w[2] + w[3]) / 6
		},
	}

	CatmullRom = ResampleFilter{
		Support: 2.0,
		Fn: func(x float64) float64 {
			b := 0.0
			c := 0.5
			var w [4]float64
			x = math.Abs(x)

			if x < 1.0 {
				w[0] = 0
				w[1] = 6 - 2*b
				w[2] = (-18 + 12*b + 6*c) * x * x
				w[3] = (12 - 9*b - 6*c) * x * x * x
			} else if x <= 2.0 {
				w[0] = 8*b + 24*c
				w[1] = (-12*b - 48*c) * x
				w[2] = (6*b + 30*c) * x * x
				w[3] = (-b - 6*c) * x * x * x
			} else {
				return 0
			}

			return (w[0] + w[1] + w[2] + w[3]) / 6
		},
	}

	Lanczos = ResampleFilter{
		Support: 3.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x == 0 {
				return 1.0
			} else if x < 3.0 {
				return (3.0 * math.Sin(math.Pi*x) * math.Sin(math.Pi*(x/3.0))) / (math.Pi * math.Pi * x * x)
			}
			return 0.0
		},
	}

	Bartlett = ResampleFilter{
		Support: 3.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 3.0 {
				return sinc(x) * (3.0 - x) / 3.0
			}
			return 0
		},
	}

	Hermite = ResampleFilter{
		Support: 1.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 1.0 {
				return bcspline(x, 0.0, 0.0)
			}
			return 0
		},
	}

	BSpline = ResampleFilter{
		Support: 2.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 2.0 {
				return bcspline(x, 1.0, 0.0)
			}
			return 0
		},
	}

	Hann = ResampleFilter{
		Support: 3.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 3.0 {
				return sinc(x) * (0.5 + 0.5*math.Cos(math.Pi*x/3.0))
			}
			return 0
		},
	}

	Hamming = ResampleFilter{
		Support: 3.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 3.0 {
				return sinc(x) * (0.54 + 0.46*math.Cos(math.Pi*x/3.0))
			}
			return 0
		},
	}

	Blackman = ResampleFilter{
		Support: 3.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 3.0 {
				return sinc(x) * (0.42 - 0.5*math.Cos(math.Pi*x/3.0+math.Pi) + 0.08*math.Cos(2.0*math.Pi*x/3.0))
			}
			return 0
		},
	}

	Welch = ResampleFilter{
		Support: 3.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 3.0 {
				return sinc(x) * (1.0 - (x * x / 9.0))
			}
			return 0
		},
	}

	Cosine = ResampleFilter{
		Support: 3.0,
		Fn: func(x float64) float64 {
			x = math.Abs(x)
			if x < 3.0 {
				return sinc(x) * math.Cos((math.Pi/2.0)*(x/3.0))
			}
			return 0
		},
	}
)

func sinc(x float64) float64 {
	if x == 0 {
		return 1
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}

func bcspline(x, b, c float64) float64 {
	x = math.Abs(x)
	if x < 1.0 {
		return ((12-9*b-6*c)*x*x*x + (-18+12*b+6*c)*x*x + (6 - 2*b)) / 6
	}
	if x < 2.0 {
		return ((-b-6*c)*x*x*x + (6*b+30*c)*x*x + (-12*b-48*c)*x + (8*b + 24*c)) / 6
	}
	return 0
}
