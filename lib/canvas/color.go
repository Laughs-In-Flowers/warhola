package canvas

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/Laughs-In-Flowers/warhola/lib/util/mth"
)

type ColorModel int

const (
	COLORNOOP ColorModel = iota
	ALPHA
	ALPHA16
	CMYK
	GRAY
	GRAY16
	NRGBA
	NRGBA64
	RGBA
	RGBA64
)

func (c ColorModel) String() string {
	switch c {
	case ALPHA:
		return "ALPHA"
	case ALPHA16:
		return "ALPHA16"
	case CMYK:
		return "CMYK"
	case GRAY:
		return "GRAY"
	case GRAY16:
		return "GRAY16"
	case NRGBA:
		return "NRGBA"
	case NRGBA64:
		return "NRGBA64"
	case RGBA:
		return "RGBA"
	case RGBA64:
		return "RGBA64"
	}
	return "COLORNOOP"
}

func stringToColorModel(s string) ColorModel {
	switch strings.ToUpper(s) {
	case "ALPHA":
		return ALPHA
	case "ALPHA16":
		return ALPHA16
	case "CMYK":
		return CMYK
	case "GRAY":
		return GRAY
	case "GRAY16":
		return GRAY16
	case "NRGBA":
		return NRGBA
	case "NRGBA64":
		return NRGBA64
	case "RGBA":
		return RGBA
	case "RGBA64":
		return RGBA64
	}
	return COLORNOOP
}

func (c ColorModel) toColorModel() color.Model {
	switch c {
	case ALPHA:
		return color.AlphaModel
	case ALPHA16:
		return color.Alpha16Model
	case CMYK:
		return color.CMYKModel
	case GRAY:
		return color.GrayModel
	case GRAY16:
		return color.Gray16Model
	case NRGBA:
		return color.NRGBAModel
	case NRGBA64:
		return color.NRGBA64Model
	case RGBA:
		return color.RGBAModel
	case RGBA64:
		return color.RGBA64Model
	}
	return nil
}

func icmTocm(m color.Model, p *pxl) {
	switch m {
	case color.AlphaModel:
		p.m = ALPHA
		return
	case color.Alpha16Model:
		p.m = ALPHA16
		return
	case color.CMYKModel:
		p.m = CMYK
		return
	case color.GrayModel:
		p.m = GRAY
		return
	case color.Gray16Model:
		p.m = GRAY16
		return
	case color.NRGBAModel:
		p.m = NRGBA
		return
	case color.NRGBA64Model:
		p.m = NRGBA64
		return
	case color.RGBAModel:
		p.m = RGBA
		return
	case color.RGBA64Model:
		p.m = RGBA64
		return
	}

	p.m = COLORNOOP
}

var (
	WorkingColorModel       = RGBA
	WorkingColorModelString = "RGBA"
	WorkingColorModelFn     = color.RGBAModel
	WorkingColorModelNew    = image.NewRGBA
)

var AvailableColorModels = []ColorModel{
	ALPHA,
	ALPHA16,
	CMYK,
	GRAY,
	GRAY16,
	NRGBA,
	NRGBA64,
	RGBA,
	RGBA64,
}

type Paletter interface {
	Palettize(color.Color) color.Color
}

type PaletteFunc func(color.Color) color.Color

func mkUniformPaletteFunc(l color.Color) PaletteFunc {
	return func(color.Color) color.Color {
		return l
	}
}

func mkPaletteFunc(p color.Palette) PaletteFunc {
	return func(l color.Color) color.Color {
		return p.Convert(l)
	}
}

// RGBA color using the range of 0-1 in float64
type RGBA164 struct {
	R, G, B, A float64
}

func newRGBA164(r, g, b, a uint8) RGBA164 {
	return RGBA164{float64(r) / 255, float64(g) / 255, float64(b) / 255, float64(a) / 255}
}

func (c *RGBA164) Clamp() {
	c.R = mth.Clamp(c.R, 0, 1)
	c.G = mth.Clamp(c.G, 0, 1)
	c.B = mth.Clamp(c.B, 0, 1)
	c.A = mth.Clamp(c.A, 0, 1)
}

func AlphaComp(bg, fg RGBA164) RGBA164 {
	fg.Clamp()
	fga := fg.A

	r := (fg.R * fga / 1) + ((1 - fga) * bg.R / 1)
	g := (fg.G * fga / 1) + ((1 - fga) * bg.G / 1)
	b := (fg.B * fga / 1) + ((1 - fga) * bg.B / 1)
	a := bg.A + fga

	return RGBA164{R: r, G: g, B: b, A: a}
}

// RGBToHSL converts from  RGB to HSL color model.
// Parameter c is the RGBA color and must implement the color.RGBA interface.
// Returned values h, s and l correspond to the hue, saturation and lightness.
// The hue is of range 0 to 360 and the saturation and lightness are of range 0.0 to 1.0.
func RGBToHSL(c color.RGBA) (float64, float64, float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	var h, s, l float64
	l = (max + min) / 2

	// Achromatic
	if delta <= 0 {
		return h, s, l
	}

	// Should it be smaller than or equals instead?
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2 - max - min)
	}

	if r >= max {
		h = (g - b) / delta
	} else if g >= max {
		h = (b-r)/delta + 2
	} else {
		h = (r-g)/delta + 4
	}

	h *= 60
	if h < 0 {
		h += 360
	}

	return h, s, l
}

// HSLToRGB converts from HSL to RGB color model.
// Parameter h is the hue and its range is from 0 to 360 degrees.
// Parameter s is the saturation and its range is from 0.0 to 1.0.
// Parameter l is the lightness and its range is from 0.0 to 1.0.
func HSLToRGB(h, s, l float64) color.RGBA {
	var r, g, b float64
	if s == 0 {
		r = l
		g = l
		b = l
	} else {
		var temp0, temp1 float64
		if l < 0.5 {
			temp0 = l * (1 + s)
		} else {
			temp0 = (l + s) - (s * l)
		}
		temp1 = 2*l - temp0

		h /= 360

		hueFn := func(v float64) float64 {
			if v < 0 {
				v++
			} else if v > 1 {
				v--
			}

			if v < 1.0/6.0 {
				return temp1 + (temp0-temp1)*6*v
			}
			if v < 1.0/2.0 {
				return temp0
			}
			if v < 2.0/3.0 {
				return temp1 + (temp0-temp1)*(2.0/3.0-v)*6
			}
			return temp1
		}

		r = hueFn(h + 1.0/3.0)
		g = hueFn(h)
		b = hueFn(h - 1.0/3.0)

	}

	outR := uint8(mth.Clamp(r*255+0.5, 0, 255))
	outG := uint8(mth.Clamp(g*255+0.5, 0, 255))
	outB := uint8(mth.Clamp(b*255+0.5, 0, 255))

	return color.RGBA{outR, outG, outB, 0xFF}
}

// RGBToHSV converts from  RGB to HSV color model.
// Parameter c is the RGBA color and must implement the color.RGBA interface.
// Returned values h, s and v correspond to the hue, saturation and value.
// The hue is of range 0 to 360 and the saturation and value are of range 0.0 to 1.0.
func RGBToHSV(c color.RGBA) (h, s, v float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	v = max
	delta := max - min

	// Avoid division by zero
	if max > 0 {
		s = delta / max
	} else {
		h = 0
		s = 0
		return
	}

	// Achromatic
	if max == min {
		h = 0
		return
	}

	if r >= max {
		h = (g - b) / delta
	} else if g >= max {
		h = (b-r)/delta + 2
	} else {
		h = (r-g)/delta + 4
	}

	h *= 60
	if h < 0 {
		h += 360
	}

	return
}

// HSVToRGB converts from HSV to RGB color model.
// Parameter h is the hue and its range is from 0 to 360 degrees.
// Parameter s is the saturation and its range is from 0.0 to 1.0.
// Parameter v is the value and its range is from 0.0 to 1.0.
func HSVToRGB(h, s, v float64) color.RGBA {
	var i, f, p, q, t float64

	// Achromatic
	if s == 0 {
		outV := uint8(mth.Clamp(v*255+0.5, 0, 255))
		return color.RGBA{outV, outV, outV, 0xFF}
	}

	h /= 60
	i = math.Floor(h)
	f = h - i
	p = v * (1 - s)
	q = v * (1 - s*f)
	t = v * (1 - s*(1-f))

	var r, g, b float64
	switch i {
	case 0:
		r = v
		g = t
		b = p
	case 1:
		r = q
		g = v
		b = p
	case 2:
		r = p
		g = v
		b = t
	case 3:
		r = p
		g = q
		b = v
	case 4:
		r = t
		g = p
		b = v
	default:
		r = v
		g = p
		b = q
	}

	outR := uint8(mth.Clamp(r*255+0.5, 0, 255))
	outG := uint8(mth.Clamp(g*255+0.5, 0, 255))
	outB := uint8(mth.Clamp(b*255+0.5, 0, 255))
	return color.RGBA{outR, outG, outB, 0xFF}
}

type ColorStats interface {
	Channel(string) (*image.Gray, error)
	Threshold(l uint8) (*image.Gray, error)
	RGBAHistogram() *RGBAHistogram
}

type channel int

const (
	cNo channel = iota
	cRed
	cGreen
	cBlue
	cAlpha
)

func stringToChannel(s string) channel {
	switch strings.ToLower(s) {
	case "red":
		return cRed
	case "green":
		return cGreen
	case "blue":
		return cBlue
	case "alpha":
		return cAlpha
	}
	return cNo
}

func (c channel) String() string {
	switch c {
	case cRed:
		return "Red"
	case cGreen:
		return "Green"
	case cBlue:
		return "Blue"
	case cAlpha:
		return "Alpha"
	}
	return "no channel"
}

//channel
func (c *canvas) Channel(ch string) (*image.Gray, error) {
	/*
	   	//if c < 0 || 3 < c {
	   	//	panic(fmt.Sprintf("channel index '%v' out of bounds. Red: 0, Green: 1, Blue: 2, Alpha: 3", c))
	   	//}

	   	src := clone.AsRGBA(img)
	   	bounds := src.Bounds()
	   	srcW, srcH := bounds.Dx(), bounds.Dy()

	   	if bounds.Empty() {
	   		return &image.Gray{}
	   	}

	   	dst := image.NewGray(bounds)

	   	parallel.Line(srcH, func(start, end int) {
	   		for y := start; y < end; y++ {
	   			for x := 0; x < srcW; x++ {
	   				srcPos := y*src.Stride + x*4
	   				dstPos := y*dst.Stride + x

	   				dst.Pix[dstPos] = src.Pix[srcPos+int(c)]
	   			}
	   		}
	   	})

	   return dst
	*/
	return nil, nil
}

//threshold
func (c *canvas) Threshold(l uint8) (*image.Gray, error) {
	/*
		src := clone.AsRGBA(img)
			bounds := src.Bounds()

			dst := image.NewGray(bounds)

			for y := 0; y < bounds.Dy(); y++ {
				for x := 0; x < bounds.Dx(); x++ {
					srcPos := y*src.Stride + x*4
					dstPos := y*dst.Stride + x

					c := src.Pix[srcPos : srcPos+4]
					r := util.Rank(color.RGBA{c[0], c[1], c[2], c[3]})

					if uint8(r) >= level {
						dst.Pix[dstPos] = 0xFF
					} else {
						dst.Pix[dstPos] = 0x00
					}
				}
			}

		return dst
	*/
	return nil, nil
}

type Histogram struct {
	Bins []int
}

// Max returns the highest count found in the histogram bins.
func (h *Histogram) Max() int {
	var max int
	if len(h.Bins) > 0 {
		max = h.Bins[0]
		for i := 1; i < len(h.Bins); i++ {
			if h.Bins[i] > max {
				max = h.Bins[i]
			}
		}
	}
	return max
}

// Min returns the lowest count found in the histogram bins.
func (h *Histogram) Min() int {
	var min int
	if len(h.Bins) > 0 {
		min = h.Bins[0]
		for i := 1; i < len(h.Bins); i++ {
			if h.Bins[i] < min {
				min = h.Bins[i]
			}
		}
	}
	return min
}

// Cumulative returns a new Histogram in which each bin is the cumulative
// value of its previous bins
func (h *Histogram) Cumulative() *Histogram {
	/*
		binCount := len(h.Bins)
		out := Histogram{make([]int, binCount)}

		if binCount > 0 {
			out.Bins[0] = h.Bins[0]
		}

		for i := 1; i < binCount; i++ {
			out.Bins[i] = out.Bins[i-1] + h.Bins[i]
		}

		return &out
	*/
	return nil
}

// Image returns a grayscale image representation of the Histogram.
// The width and height of the image will be equivalent to the number of Bins in the Histogram.
func (h *Histogram) Image() *image.Gray {
	/*
		dstW, dstH := len(h.Bins), len(h.Bins)
		dst := image.NewGray(image.Rect(0, 0, dstW, dstH))

		max := h.Max()
		if max == 0 {
			max = 1
		}

		for x := 0; x < dstW; x++ {
			value := ((int(h.Bins[x]) << 16 / max) * dstH) >> 16
			// Fill from the bottom up
			for y := dstH - 1; y > dstH-value-1; y-- {
				dst.Pix[y*dst.Stride+x] = 0xFF
			}
		}
		return dst
	*/
	return nil
}

type RGBAHistogram struct {
	R Histogram
	G Histogram
	B Histogram
	A Histogram
}

// NewRGBAHistogram constructs a RGBAHistogram out of the provided image.
// A sub-histogram is created per RGBA channel with 256 bins each.
func NewRGBAHistogram(img image.Image) *RGBAHistogram {
	/*
		src := clone.AsRGBA(img)

		binCount := 256
		r := Histogram{make([]int, binCount)}
		g := Histogram{make([]int, binCount)}
		b := Histogram{make([]int, binCount)}
		a := Histogram{make([]int, binCount)}

		for y := 0; y < src.Bounds().Dy(); y++ {
			for x := 0; x < src.Bounds().Dx(); x++ {
				pos := y*src.Stride + x*4
				r.Bins[src.Pix[pos+0]]++
				g.Bins[src.Pix[pos+1]]++
				b.Bins[src.Pix[pos+2]]++
				a.Bins[src.Pix[pos+3]]++
			}
		}

		return &RGBAHistogram{R: r, G: g, B: b, A: a}
	*/
	return nil
}

// Cumulative returns a new RGBAHistogram in which each bin is the cumulative
// value of its previous bins per channel.
func (h *RGBAHistogram) Cumulative() *RGBAHistogram {
	/*
		binCount := len(h.R.Bins)

		r := Histogram{make([]int, binCount)}
		g := Histogram{make([]int, binCount)}
		b := Histogram{make([]int, binCount)}
		a := Histogram{make([]int, binCount)}

		out := RGBAHistogram{R: r, G: g, B: b, A: a}

		if binCount > 0 {
			out.R.Bins[0] = h.R.Bins[0]
			out.G.Bins[0] = h.G.Bins[0]
			out.B.Bins[0] = h.B.Bins[0]
			out.A.Bins[0] = h.A.Bins[0]
		}

		for i := 1; i < binCount; i++ {
			out.R.Bins[i] = out.R.Bins[i-1] + h.R.Bins[i]
			out.G.Bins[i] = out.G.Bins[i-1] + h.G.Bins[i]
			out.B.Bins[i] = out.B.Bins[i-1] + h.B.Bins[i]
			out.A.Bins[i] = out.A.Bins[i-1] + h.A.Bins[i]
		}

		return &out
	*/
	return nil
}

// Image returns an RGBA image representation of the RGBAHistogram.
// An image width of 256 represents the 256 Bins per channel and the
// image height of 256 represents the max normalized histogram value per channel.
// Each RGB channel from the histogram is mapped to its corresponding channel in the image,
// so that for example if the red channel is extracted from the image, it corresponds to the
// red channel histogram.
func (h *RGBAHistogram) Image() *image.RGBA {
	/*
		if len(h.R.Bins) != 256 || len(h.G.Bins) != 256 ||
			len(h.B.Bins) != 256 || len(h.A.Bins) != 256 {
			panic("RGBAHistogram bins length not equal to 256")
		}

		dstW, dstH := 256, 256
		dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

		maxR := h.R.Max()
		if maxR == 0 {
			maxR = 1
		}
		maxG := h.G.Max()
		if maxG == 0 {
			maxG = 1
		}
		maxB := h.B.Max()
		if maxB == 0 {
			maxB = 1
		}

		for x := 0; x < dstW; x++ {
			binHeightR := ((int(h.R.Bins[x]) << 16 / maxR) * dstH) >> 16
			binHeightG := ((int(h.G.Bins[x]) << 16 / maxG) * dstH) >> 16
			binHeightB := ((int(h.B.Bins[x]) << 16 / maxB) * dstH) >> 16
			// Fill from the bottom up
			for y := dstH - 1; y >= 0; y-- {
				pos := y*dst.Stride + x*4
				iy := dstH - 1 - y

				if iy < binHeightR {
					dst.Pix[pos+0] = 0xFF
				}
				if iy < binHeightG {
					dst.Pix[pos+1] = 0xFF
				}
				if iy < binHeightB {
					dst.Pix[pos+2] = 0xFF
				}
				dst.Pix[pos+3] = 0xFF
			}
		}

		return dst
	*/
	return nil
}

//histogram
func (c *canvas) RGBAHistogram() *RGBAHistogram {
	return nil
}
