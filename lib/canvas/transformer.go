package canvas

import (
	"image"
	"image/color"
	"math"

	"github.com/Laughs-In-Flowers/warhola/lib/util/xrr"
)

type Transformer interface {
	Cropper
	Resizer
}

// An interface for cropping a Canvas, modify the existing canvas or return a cropped clone.
type Cropper interface {
	Crop(image.Rectangle) error
	CropTo(image.Rectangle) (Canvas, error)
}

var EmptyIntersectError = xrr.Xrror("Unable to crop empty intersect of %v and provided %v").Out

// crops the canvas
func (c *canvas) Crop(r image.Rectangle) error {
	exp := r
	p := c.pxl
	r = r.Intersect(p.rect)
	if r.Empty() {
		return EmptyIntersectError(p.rect, exp)
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	nr := image.Rectangle{image.Point{0, 0}, r.Size()}
	c.Printf("canvas dimensions are now min %v max %v", nr.Min, nr.Max)
	c.pxl = &pxl{
		m:         p.m,
		pix:       p.pix[i:],
		str:       p.str,
		rect:      nr,
		paletteFn: c.paletteFn,
		measure:   newMeasure(&r, p.measure.pp, p.measure.ppu),
	}
	return nil
}

// returns a cropped clone
func (c *canvas) CropTo(r image.Rectangle) (Canvas, error) {
	nc := c.Clone()
	err := nc.Crop(r)
	return nc, err
}

type Resizer interface {
	Resize(w, h int, filter string) error
	ResizeTo(w, h int, filter string) (Canvas, error)
}

var ZeroResizeError = xrr.Xrror("zero value prevents resizing\n\twidth %d\n\theight %d\n\tempty canvas: %t)").Out

// resize the canvas
func (c *canvas) Resize(w, h int, filter string) error {
	b := c.Bounds().Empty()
	if w <= 0 || h <= 0 || b {
		return ZeroResizeError(w, h, b)
	}
	f := stringToFilter(filter)
	switch {
	case f.Support <= 0:
		nearest(c, w, h)
	default:
		resampleH(c, w, f)
		resampleV(c, h, f)
	}
	return nil
}

// return a resized clone of the canvas
func (c *canvas) ResizeTo(w, h int, filter string) (Canvas, error) {
	nc := c.Clone()
	err := nc.Resize(w, h, filter)
	return nc, err
}

func resampleH(src *canvas, w int, f ResampleFilter) {
	srcP := src.pxl.clone(color.RGBAModel)
	srcWidth, srcHeight := srcP.Bounds().Dx(), srcP.Bounds().Dy()
	srcStride := srcP.Stride()

	delta := float64(srcWidth) / float64(w)
	scale := math.Max(delta, 1.0)

	dstP := scratch(srcP, srcP.ColorModel(), w, srcHeight)
	dstStride := dstP.Stride()

	filterRadius := math.Ceil(scale * f.Support)

	parallel(srcHeight, func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < w; x++ {
				ix := (float64(x)+0.5)*delta - 0.5
				istart, iend := int(ix-filterRadius+0.5), int(ix+filterRadius)

				if istart < 0 {
					istart = 0
				}
				if iend >= srcWidth {
					iend = srcWidth - 1
				}

				var r, g, b, a float64
				var sum float64
				for kx := istart; kx <= iend; kx++ {
					srcPos := y*srcStride + kx*4

					normPos := (float64(kx) - ix) / scale
					fValue := f.Fn(normPos)

					r += float64(srcP.pix[srcPos+0]) * fValue
					g += float64(srcP.pix[srcPos+1]) * fValue
					b += float64(srcP.pix[srcPos+2]) * fValue
					a += float64(srcP.pix[srcPos+3]) * fValue
					sum += fValue
				}

				dstPos := y*dstStride + x*4
				dstP.pix[dstPos+0] = uint8(clamp((r/sum)+0.5, 0, 255))
				dstP.pix[dstPos+1] = uint8(clamp((g/sum)+0.5, 0, 255))
				dstP.pix[dstPos+2] = uint8(clamp((b/sum)+0.5, 0, 255))
				dstP.pix[dstPos+3] = uint8(clamp((a/sum)+0.5, 0, 255))
			}
		}
	})

	src.pxl = dstP.clone(src.ColorModel())
}

func resampleV(src *canvas, h int, f ResampleFilter) {
	srcP := src.pxl.clone(color.RGBAModel)
	srcWidth, srcHeight := srcP.Bounds().Dx(), srcP.Bounds().Dy()
	srcStride := srcP.Stride()

	delta := float64(srcHeight) / float64(h)
	scale := math.Max(delta, 1.0)

	dstP := scratch(srcP, srcP.ColorModel(), srcWidth, h)
	dstStride := dstP.Stride()

	filterRadius := math.Ceil(scale * f.Support)

	parallel(h, func(start, end int) {
		for y := start; y < end; y++ {
			iy := (float64(y)+0.5)*delta - 0.5

			istart, iend := int(iy-filterRadius+0.5), int(iy+filterRadius)

			if istart < 0 {
				istart = 0
			}
			if iend >= srcHeight {
				iend = srcHeight - 1
			}

			for x := 0; x < srcWidth; x++ {
				var r, g, b, a float64
				var sum float64
				for ky := istart; ky <= iend; ky++ {
					srcPos := ky*srcStride + x*4
					normPos := (float64(ky) - iy) / scale
					fValue := f.Fn(normPos)

					r += float64(srcP.pix[srcPos+0]) * fValue
					g += float64(srcP.pix[srcPos+1]) * fValue
					b += float64(srcP.pix[srcPos+2]) * fValue
					a += float64(srcP.pix[srcPos+3]) * fValue
					sum += fValue
				}

				dstPos := y*dstStride + x*4
				dstP.pix[dstPos+0] = uint8(clamp((r/sum)+0.5, 0, 255))
				dstP.pix[dstPos+1] = uint8(clamp((g/sum)+0.5, 0, 255))
				dstP.pix[dstPos+2] = uint8(clamp((b/sum)+0.5, 0, 255))
				dstP.pix[dstPos+3] = uint8(clamp((a/sum)+0.5, 0, 255))
			}
		}
	})

	src.pxl = dstP.clone(src.ColorModel())
}
