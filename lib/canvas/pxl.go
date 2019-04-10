package canvas

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"

	"github.com/Laughs-In-Flowers/warhola/lib/util/prl"
	"github.com/Laughs-In-Flowers/xrr"
)

// Interface for core image data. Encapsulates image.Image plus other image
// package image functionality in addition to direct pixel access, paletting,
// pasting, and physical measurement.
type Pxl interface {
	ImageImage
	Measure
	Paletter
	Paster
	Pix
}

// The interface corresponding to the most relevant functionality of Image
// in the image package.
type ImageImage interface {
	ColorModel() color.Model
	Bounds() image.Rectangle
	At(x, y int) color.Color
	Opaque() bool
	PixOffset(int, int) int
	Set(x, y int, c color.Color)
	SubImage(image.Rectangle) image.Image
}

// An interface for pasting.
type Paster interface {
	Paste(draw.Image, image.Point)
	Overlay(draw.Image, image.Point, float64)
}

// An interface for direct pix manipulation.
type Pix interface {
	Pix() []uint8
	SetPix([]uint8)
	Stride() int
}

type pxl struct {
	m         ColorModel
	pix       []uint8
	str       int
	rect      image.Rectangle
	paletteFn PaletteFunc
	*measure
}

func newPxl() *pxl {
	p := &pxl{
		pix:  make([]uint8, 0),
		rect: image.ZR,
	}
	p.measure = newMeasure(&p.rect, defaultPP, defaultPPU)
	return p
}

func Scratch(cm color.Model, X, Y int) *pxl {
	return scratch(newPxl(), cm, X, Y)
}

func scratch(p *pxl, cm color.Model, X, Y int) *pxl {
	np := &pxl{
		pix:       make([]uint8, 0),
		rect:      image.Rect(0, 0, X, Y),
		paletteFn: p.paletteFn,
	}
	icmTocm(cm, np)
	newTo(np)
	np.measure = newMeasure(&np.rect, p.measure.pp, p.measure.ppu)
	return np
}

func (p *pxl) clone(cm color.Model) *pxl {
	r := p.Bounds()
	np := &pxl{
		m:         p.m,
		pix:       make([]uint8, 0),
		rect:      image.Rectangle{r.Min, r.Max},
		paletteFn: p.paletteFn,
	}
	np.measure = newMeasure(&np.rect, p.measure.pp, p.measure.ppu)
	icmTocm(cm, np)
	newTo(np)
	draw.Draw(np, np.rect, p, image.ZP, draw.Src)
	return np
}

type padMode int

const (
	pmNoFill padMode = iota
	pmExtend
	pmWrap
)

func (p *pxl) pad(m padMode, px, py int) *pxl {
	dstP := p.clone(color.RGBAModel)
	switch m {
	case pmNoFill:
		return pxlNoFill(dstP, px, py)
	case pmExtend:
		return pxlExtend(dstP, px, py)
	case pmWrap:
		return pxlWrap(dstP, px, py)
	}
	return p
}

func pxlNoFill(p *pxl, px, py int) *pxl {
	srcBounds := p.Bounds()
	paddedW, paddedH := srcBounds.Dx()+2*px, srcBounds.Dy()+2*py
	newBounds := image.Rect(0, 0, paddedW, paddedH)
	fillBounds := image.Rect(px, py, px+srcBounds.Dx(), py+srcBounds.Dy())
	dstP := scratch(p, p.ColorModel(), newBounds.Max.X, newBounds.Max.Y)
	draw.Draw(dstP, fillBounds, p, srcBounds.Min, draw.Src)
	return dstP
}

func pxlExtend(p *pxl, px, py int) *pxl {
	dstP := pxlNoFill(p, px, py)
	dstPB := dstP.Bounds()
	paddedW, paddedH := dstPB.Dx(), dstPB.Dy()

	prl.Run(paddedH, func(start, end int) {
		for y := start; y < end; y++ {
			iy := y
			if iy < py {
				iy = py
			} else if iy >= paddedH-py {
				iy = paddedH - py - 1
			}

			for x := 0; x < paddedW; x++ {
				ix := x
				if ix < px {
					ix = px
				} else if x >= paddedW-px {
					ix = paddedW - px - 1
				} else if iy == y {
					// This only enters if we are not in a y-padded area or
					// x-padded area, so nothing to extend here.
					// So simply jump to the next padded-x index.
					x = paddedW - px - 1
					continue
				}

				dstPos := y*dstP.str + x*4
				edgePos := iy*dstP.str + ix*4

				dstP.pix[dstPos+0] = dstP.pix[edgePos+0]
				dstP.pix[dstPos+1] = dstP.pix[edgePos+1]
				dstP.pix[dstPos+2] = dstP.pix[edgePos+2]
				dstP.pix[dstPos+3] = dstP.pix[edgePos+3]
			}
		}
	})

	return dstP
}

func pxlWrap(p *pxl, px, py int) *pxl {
	dstP := pxlNoFill(p, px, py)
	dstPB := dstP.Bounds()
	paddedW, paddedH := dstPB.Dx(), dstPB.Dy()

	prl.Run(paddedH, func(start, end int) {
		for y := start; y < end; y++ {
			iy := y
			if iy < py {
				iy = (paddedH - py) - ((py - y) % (paddedH - py*2))
			} else if iy >= paddedH-py {
				iy = py - ((py - y) % (paddedH - py*2))
			}

			for x := 0; x < paddedW; x++ {
				ix := x
				if ix < px {
					ix = (paddedW - px) - ((px - x) % (paddedW - px*2))
				} else if ix >= paddedW-px {
					ix = px - ((px - x) % (paddedW - px*2))
				} else if iy == y {
					// This only enters if we are not in a y-padded area or
					// x-padded area, so nothing to extend here.
					// So simply jump to the next padded-x index.
					x = paddedW - px - 1
					continue
				}

				dstPos := y*dstP.str + x*4
				edgePos := iy*dstP.str + ix*4

				dstP.pix[dstPos+0] = dstP.pix[edgePos+0]
				dstP.pix[dstPos+1] = dstP.pix[edgePos+1]
				dstP.pix[dstPos+2] = dstP.pix[edgePos+2]
				dstP.pix[dstPos+3] = dstP.pix[edgePos+3]
			}
		}
	})

	return dstP
}

func (p *pxl) ColorModel() color.Model {
	return p.m.toColorModel()
}

func (p *pxl) Bounds() image.Rectangle {
	return p.rect
}

type atFunc func(*pxl, int, int) color.Color

var atFuncs = map[ColorModel]atFunc{
	ALPHA: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.Alpha{}
		}
		i := p.PixOffset(x, y)
		return color.Alpha{p.pix[i]}
	},
	ALPHA16: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.Alpha16{}
		}
		i := p.PixOffset(x, y)
		return color.Alpha16{uint16(p.pix[i+0])<<8 | uint16(p.pix[i+1])}
	},
	CMYK: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.CMYK{}
		}
		i := p.PixOffset(x, y)
		return color.CMYK{p.pix[i+0],
			p.pix[i+1],
			p.pix[i+2],
			p.pix[i+3]}
	},
	GRAY: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.Gray{}
		}
		i := p.PixOffset(x, y)
		return color.Gray{p.pix[i]}
	},
	GRAY16: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.Gray16{}
		}
		i := p.PixOffset(x, y)
		return color.Gray16{uint16(p.pix[i+0])<<8 | uint16(p.pix[i+1])}
	},
	NRGBA: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.NRGBA{}
		}
		i := p.PixOffset(x, y)
		return color.NRGBA{p.pix[i+0],
			p.pix[i+1],
			p.pix[i+2],
			p.pix[i+3]}
	},
	NRGBA64: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.NRGBA64{}
		}
		i := p.PixOffset(x, y)
		return color.NRGBA64{
			uint16(p.pix[i+0])<<8 | uint16(p.pix[i+1]),
			uint16(p.pix[i+2])<<8 | uint16(p.pix[i+3]),
			uint16(p.pix[i+4])<<8 | uint16(p.pix[i+5]),
			uint16(p.pix[i+6])<<8 | uint16(p.pix[i+7]),
		}
	},
	RGBA: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.RGBA{}
		}

		i := p.PixOffset(x, y)
		return color.RGBA{p.pix[i+0],
			p.pix[i+1],
			p.pix[i+2],
			p.pix[i+3]}
	},
	RGBA64: func(p *pxl, x, y int) color.Color {
		if !(image.Point{x, y}.In(p.rect)) {
			return color.RGBA64{}
		}
		i := p.PixOffset(x, y)
		return color.RGBA64{
			uint16(p.pix[i+0])<<8 | uint16(p.pix[i+1]),
			uint16(p.pix[i+2])<<8 | uint16(p.pix[i+3]),
			uint16(p.pix[i+4])<<8 | uint16(p.pix[i+5]),
			uint16(p.pix[i+6])<<8 | uint16(p.pix[i+7]),
		}
	},
}

func (p *pxl) At(x, y int) color.Color {
	if fn, ok := atFuncs[p.m]; ok {
		return fn(p, x, y)
	}
	return color.RGBA{}
}

type opaqueFunc func(*pxl) bool

var opaqueFuncs = map[ColorModel]opaqueFunc{
	ALPHA: func(p *pxl) bool {
		if p.rect.Empty() {
			return true
		}
		i0, i1 := 0, p.rect.Dx()
		for y := p.rect.Min.Y; y < p.rect.Max.Y; y++ {
			for i := i0; i < i1; i++ {
				if p.pix[i] != 0xff {
					return false
				}
			}
			i0 += p.str
			i1 += p.str
		}
		return true
	},
	ALPHA16: func(p *pxl) bool {
		if p.rect.Empty() {
			return true
		}
		i0, i1 := 0, p.rect.Dx()*2
		for y := p.rect.Min.Y; y < p.rect.Max.Y; y++ {
			for i := i0; i < i1; i += 2 {
				if p.pix[i+0] != 0xff || p.pix[i+1] != 0xff {
					return false
				}
			}
			i0 += p.str
			i1 += p.str
		}
		return true
	},
	CMYK: func(p *pxl) bool {
		return true
	},
	GRAY: func(p *pxl) bool {
		return true
	},
	GRAY16: func(p *pxl) bool {
		return true
	},
	NRGBA: func(p *pxl) bool {
		if p.rect.Empty() {
			return true
		}
		i0, i1 := 3, p.rect.Dx()*4
		for y := p.rect.Min.Y; y < p.rect.Max.Y; y++ {
			for i := i0; i < i1; i += 4 {
				if p.pix[i] != 0xff {
					return false
				}
			}
			i0 += p.str
			i1 += p.str
		}
		return true
	},
	NRGBA64: func(p *pxl) bool {
		if p.rect.Empty() {
			return true
		}
		i0, i1 := 6, p.rect.Dx()*8
		for y := p.rect.Min.Y; y < p.rect.Max.Y; y++ {
			for i := i0; i < i1; i += 8 {
				if p.pix[i+0] != 0xff || p.pix[i+1] != 0xff {
					return false
				}
			}
			i0 += p.str
			i1 += p.str
		}
		return true
	},
	RGBA: func(p *pxl) bool {
		if p.rect.Empty() {
			return true
		}
		i0, i1 := 3, p.rect.Dx()*4
		for y := p.rect.Min.Y; y < p.rect.Max.Y; y++ {
			for i := i0; i < i1; i += 4 {
				if p.pix[i] != 0xff {
					return false
				}
			}
			i0 += p.str
			i1 += p.str
		}
		return true
	},
	RGBA64: func(p *pxl) bool {
		if p.rect.Empty() {
			return true
		}
		i0, i1 := 6, p.rect.Dx()*8
		for y := p.rect.Min.Y; y < p.rect.Max.Y; y++ {
			for i := i0; i < i1; i += 8 {
				if p.pix[i+0] != 0xff || p.pix[i+1] != 0xff {
					return false
				}
			}
			i0 += p.str
			i1 += p.str
		}
		return true
	},
}

func (p *pxl) Opaque() bool {
	if fn, ok := opaqueFuncs[p.m]; ok {
		return fn(p)
	}
	return true
}

type setFunc func(*pxl, int, int, color.Color)

var setFuncs = map[ColorModel]setFunc{
	ALPHA: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		p.pix[i] = color.AlphaModel.Convert(l).(color.Alpha).A
	},
	ALPHA16: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		ll := color.Alpha16Model.Convert(l).(color.Alpha16)
		p.pix[i+0] = uint8(ll.A >> 8)
		p.pix[i+1] = uint8(ll.A)
	},
	CMYK: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		ll := color.CMYKModel.Convert(l).(color.CMYK)
		p.pix[i+0] = ll.C
		p.pix[i+1] = ll.M
		p.pix[i+2] = ll.Y
		p.pix[i+3] = ll.K
	},
	GRAY: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		p.pix[i] = color.GrayModel.Convert(l).(color.Gray).Y
	},
	GRAY16: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		ll := color.Gray16Model.Convert(l).(color.Gray16)
		p.pix[i+0] = uint8(ll.Y >> 8)
		p.pix[i+1] = uint8(ll.Y)
	},
	NRGBA: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		ll := color.NRGBAModel.Convert(l).(color.NRGBA)
		p.pix[i+0] = ll.R
		p.pix[i+1] = ll.G
		p.pix[i+2] = ll.B
		p.pix[i+3] = ll.A
	},
	NRGBA64: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		ll := color.NRGBA64Model.Convert(l).(color.NRGBA64)
		p.pix[i+0] = uint8(ll.R >> 8)
		p.pix[i+1] = uint8(ll.R)
		p.pix[i+2] = uint8(ll.G >> 8)
		p.pix[i+3] = uint8(ll.G)
		p.pix[i+4] = uint8(ll.B >> 8)
		p.pix[i+5] = uint8(ll.B)
		p.pix[i+6] = uint8(ll.A >> 8)
		p.pix[i+7] = uint8(ll.A)
	},
	RGBA: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		ll := color.RGBAModel.Convert(l).(color.RGBA)
		p.pix[i+0] = ll.R
		p.pix[i+1] = ll.G
		p.pix[i+2] = ll.B
		p.pix[i+3] = ll.A
	},
	RGBA64: func(p *pxl, x, y int, l color.Color) {
		if !(image.Point{x, y}.In(p.rect)) {
			return
		}
		i := p.PixOffset(x, y)
		ll := color.RGBA64Model.Convert(l).(color.RGBA64)
		p.pix[i+0] = uint8(ll.R >> 8)
		p.pix[i+1] = uint8(ll.R)
		p.pix[i+2] = uint8(ll.G >> 8)
		p.pix[i+3] = uint8(ll.G)
		p.pix[i+4] = uint8(ll.B >> 8)
		p.pix[i+5] = uint8(ll.B)
		p.pix[i+6] = uint8(ll.A >> 8)
		p.pix[i+7] = uint8(ll.A)
	},
}

func (p *pxl) Set(x, y int, l color.Color) {
	if fn, ok := setFuncs[p.m]; ok {
		fn(p, x, y, p.Palettize(l))
	}
}

type pixOffsetFunc func(*pxl, int, int) int

func offset(i int) pixOffsetFunc {
	return func(p *pxl, x, y int) int {
		return (y-p.rect.Min.Y)*p.str + (x-p.rect.Min.X)*i
	}
}

var pixOffsetFuncs = map[ColorModel]pixOffsetFunc{
	ALPHA:   offset(1),
	ALPHA16: offset(2),
	CMYK:    offset(4),
	GRAY:    offset(1),
	GRAY16:  offset(2),
	NRGBA:   offset(4),
	NRGBA64: offset(8),
	RGBA:    offset(4),
	RGBA64:  offset(8),
}

func (p *pxl) PixOffset(x, y int) int {
	if fn, ok := pixOffsetFuncs[p.m]; ok {
		return fn(p, x, y)
	}
	return 0
}

func (p *pxl) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.rect)
	if r.Empty() {
		return &pxl{m: p.m}
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	return &pxl{
		m:       p.m,
		pix:     p.pix[i:],
		str:     p.str,
		rect:    r,
		measure: newMeasure(&r, p.measure.pp, p.measure.ppu),
	}
}

func newTo(p *pxl) error {
	switch p.m {
	case ALPHA:
		newFromOffset(1, p)
	case ALPHA16:
		newFromOffset(2, p)
	case CMYK:
		newFromOffset(4, p)
	case GRAY:
		newFromOffset(1, p)
	case GRAY16:
		newFromOffset(2, p)
	case NRGBA:
		newFromOffset(4, p)
	case NRGBA64:
		newFromOffset(8, p)
	case RGBA:
		newFromOffset(4, p)
	case RGBA64:
		newFromOffset(8, p)
	}
	return nil
}

func newFromOffset(o int, p *pxl) {
	r := p.rect
	w, h := r.Dx(), r.Dy()
	p.str = o * w
	p.pix = make([]uint8, o*w*h)
}

func OpenTo(path string) (image.Image, error) {
	np := newPxl()
	_, _, err := openTo(path, np)
	if err != nil {
		return nil, err
	}
	return np, nil
}

func openTo(path string, p *pxl) (FileType, ColorModel, error) {
	file, fErr := openFile(path)
	if fErr != nil {
		return FILETYPENOOP, COLORNOOP, fErr
	}
	defer file.Close()

	i, ext, dErr := image.Decode(file)
	if dErr != nil {
		return FILETYPENOOP, COLORNOOP, dErr
	}
	cm, eErr := existingTo(i, p)
	return stringToFileType(ext), cm, eErr
}

func existingTo(in image.Image, p *pxl) (ColorModel, error) {
	var i image.Image
	switch in.(type) {
	case *image.Paletted, *image.Uniform:
		i = in
	default:
		i = toCorrectedColorModel(in, p)
	}
	switch i.(type) {
	case *image.Alpha:
		ii := i.(*image.Alpha)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.Alpha16:
		ii := i.(*image.Alpha16)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.CMYK:
		ii := i.(*image.CMYK)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.Gray:
		ii := i.(*image.Gray)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.Gray16:
		ii := i.(*image.Gray16)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.NRGBA:
		ii := i.(*image.NRGBA)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.NRGBA64:
		ii := i.(*image.NRGBA64)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.Paletted:
		ii := i.(*image.Paletted)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
		p.paletteFn = mkPaletteFunc(ii.Palette)
	case *image.RGBA:
		ii := i.(*image.RGBA)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.RGBA64:
		ii := i.(*image.RGBA64)
		p.pix = ii.Pix
		p.str = ii.Stride
		p.rect = ii.Rect
	case *image.Uniform:
		ii := i.(*image.Uniform)
		p.paletteFn = mkUniformPaletteFunc(ii.C)
		p.rect = image.Rectangle{image.Point{0, 0}, image.Point{5000, 5000}}
	default:
		return p.m, toPxlError("no corresponding way to move opened %v to %s for existing canvas file",
			i.ColorModel(),
			p.m)
	}

	return p.m, nil
}

var toPxlError = xrr.Xrror("to pxl error: %s").Out

func toCorrectedColorModel(in image.Image, p *pxl) image.Image {
	var out draw.Image
	xcm := p.m
	switch xcm {
	case ALPHA:
		out = image.NewAlpha(in.Bounds())
	case ALPHA16:
		out = image.NewAlpha16(in.Bounds())
	case CMYK:
		out = image.NewCMYK(in.Bounds())
	case GRAY:
		out = image.NewGray(in.Bounds())
	case GRAY16:
		out = image.NewGray16(in.Bounds())
	case NRGBA:
		out = image.NewNRGBA(in.Bounds())
	case NRGBA64:
		out = image.NewNRGBA64(in.Bounds())
	case RGBA:
		out = image.NewRGBA(in.Bounds())
	case RGBA64:
		out = image.NewRGBA64(in.Bounds())
	default:
		out = WorkingColorModelNew(in.Bounds())
		p.m = WorkingColorModel
	}
	draw.Draw(out, out.Bounds(), in, image.ZP, draw.Src)
	return out
}

func save(path string, t FileType, p *pxl) error {
	f, err := openFile(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.encode(f, p)
}

func openFile(path string) (*os.File, error) {
	p := filepath.Clean(path)

	dir, name := filepath.Split(p)

	var fp string
	var aErr error
	switch dir {
	case "":
		fp, aErr = filepath.Abs(name)
	default:
		exist(dir)
		fp, aErr = filepath.Abs(p)
	}
	if aErr != nil {
		return nil, aErr
	}

	var file *os.File
	var oErr error
	if file, oErr = os.OpenFile(fp, os.O_RDWR|os.O_CREATE, 0660); oErr != nil {
		return nil, openError(fp, path)
	}

	return file, nil
}

func exist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir|0755)
	}
}

var openError = xrr.Xrror("unable to find or open file %s, provided %s").Out

func (p *pxl) Palettize(in color.Color) color.Color {
	if p.paletteFn != nil {
		return p.paletteFn(in)
	}
	return in
}

// paste the provided image to the canvas at the provided point
func (p *pxl) Paste(src draw.Image, pt image.Point) {
	draw.Draw(p, p.rect, src, pt, draw.Src)
}

// overlay the image to the canvas at the specified point with the specified opacity
func (p *pxl) Overlay(src draw.Image, pt image.Point, o float64) {
	m := image.NewUniform(color.Alpha{uint8(255 * (o / 100))})
	draw.DrawMask(p, p.rect, src, pt, m, image.ZP, draw.Over)
}

func (p *pxl) Pix() []uint8 {
	return p.pix
}

func (p *pxl) SetPix(x []uint8) {
	p.pix = x
}

func (p *pxl) Stride() int {
	return p.str
}

type mutationFn func() (*pxl, error)

func mutate(src *pxl, fn mutationFn) (*pxl, error) {
	ret, err := fn()
	if err != nil {
		return src, err
	}
	return ret.clone(src.ColorModel()), nil
}
