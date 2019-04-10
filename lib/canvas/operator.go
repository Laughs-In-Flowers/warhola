package canvas

import (
	"image"
	"image/color"
	"math"

	"github.com/Laughs-In-Flowers/warhola/lib/util/mth"
	"github.com/Laughs-In-Flowers/warhola/lib/util/prl"
	"github.com/Laughs-In-Flowers/xrr"
)

// An interface for performing (relatively default & easy) operations on an image.
type Operator interface {
	Adjuster
	Blender
	Convoluter
	Noiser
	Transformer
	Translater
}

type AdjustmentFunc func(color.RGBA) color.RGBA

type Adjuster interface {
	Adjust(AdjustmentFunc) error
}

func (c *canvas) Adjust(fn AdjustmentFunc) error {
	return c.mutate(func() (*pxl, error) {
		return adjustment(c.pxl, fn)
	})
}

func adjustment(p *pxl, afn AdjustmentFunc) (*pxl, error) {
	return mutate(p, func() (*pxl, error) {
		srcP := p.clone(WorkingColorModelFn)
		sb := srcP.Bounds()
		w, h := sb.Dx(), sb.Dy()
		dstP := scratch(srcP, WorkingColorModelFn, w, h)
		prl.Run(h, func(start, end int) {
			for y := start; y < end; y++ {
				for x := 0; x < w; x++ {
					srcPos := y*srcP.str + x*4

					c := color.RGBA{}

					c.R = srcP.pix[srcPos+0]
					c.G = srcP.pix[srcPos+1]
					c.B = srcP.pix[srcPos+2]
					c.A = srcP.pix[srcPos+3]

					c = afn(c)

					dstP.pix[srcPos+0] = c.R
					dstP.pix[srcPos+1] = c.G
					dstP.pix[srcPos+2] = c.B
					dstP.pix[srcPos+3] = c.A
				}
			}
		})
		return dstP, nil
	})
}

type BlendPosition int

const (
	NoBlendPosition BlendPosition = iota
	FG
	BG
)

type BlendFunc func(RGBA164, RGBA164) RGBA164

type Blender interface {
	Blend(image.Image, BlendPosition, BlendFunc) error
}

// Blend the provided image with the Canvas, at the provided position using the
// provided BlendFunc
func (c *canvas) Blend(i image.Image, pos BlendPosition, fn BlendFunc) error {
	return c.mutate(func() (*pxl, error) {
		return blend(c.pxl, i, pos, fn)
	})
}

var NoBlendPositionError = xrr.Xrror("no blend position")

func blend(p *pxl, i image.Image, pos BlendPosition, bfn BlendFunc) (*pxl, error) {
	if pos == NoBlendPosition {
		return p, NoBlendPositionError
	}

	np := scratch(p, i.ColorModel(), 0, 0)
	existingTo(i, np)

	var bg, fg *pxl
	switch pos {
	case BG:
		bg = np
		fg = p
	case FG:
		bg = p
		fg = np
	}

	bgBounds := bg.Bounds()
	fgBounds := fg.Bounds()

	var w, h int
	if bgBounds.Dx() < fgBounds.Dx() {
		w = bgBounds.Dx()
	} else {
		w = fgBounds.Dx()
	}
	if bgBounds.Dy() < fgBounds.Dy() {
		h = bgBounds.Dy()
	} else {
		h = fgBounds.Dy()
	}

	bgSrc := bg.clone(WorkingColorModelFn)
	fgSrc := fg.clone(WorkingColorModelFn)
	dstP := scratch(bg, WorkingColorModelFn, w, h)

	prl.Run(h, func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < w; x++ {
				bgPos := y*bgSrc.str + x*4
				fgPos := y*fgSrc.str + x*4
				result := bfn(
					newRGBA164(bgSrc.pix[bgPos+0], bgSrc.pix[bgPos+1], bgSrc.pix[bgPos+2], bgSrc.pix[bgPos+3]),
					newRGBA164(fgSrc.pix[fgPos+0], fgSrc.pix[fgPos+1], fgSrc.pix[fgPos+2], fgSrc.pix[fgPos+3]),
				)
				result.Clamp()
				dstPos := y*dstP.str + x*4
				dstP.pix[dstPos+0] = uint8(result.R * 255)
				dstP.pix[dstPos+1] = uint8(result.G * 255)
				dstP.pix[dstPos+2] = uint8(result.B * 255)
				dstP.pix[dstPos+3] = uint8(result.A * 255)
			}

		}
	})

	return dstP, nil
}

type Convoluter interface {
	Convolve(mth.Matrix, float64, bool, bool) error
}

func (c *canvas) Convolve(m mth.Matrix, bias float64, wrap, keepAlpha bool) error {
	return c.mutate(func() (*pxl, error) {
		return convolve(c.pxl, m, bias, wrap, keepAlpha)
	})
}

func convolve(p *pxl, m mth.Matrix, bias float64, wrap, keepAlpha bool) (*pxl, error) {
	return mutate(p, func() (*pxl, error) {
		srcP := p.clone(color.RGBAModel)

		// Kernel attributes
		lenX := m.MaxX()
		lenY := m.MaxY()
		radiusX := lenX / 2
		radiusY := lenY / 2

		// Pad the source image, basically pre-computing the pixels outside of image bounds
		switch {
		case wrap:
			srcP = p.pad(pmWrap, radiusX, radiusY)
		default:
			srcP = p.pad(pmExtend, radiusX, radiusY)
		}

		// src bounds now includes padded pixels
		srcPBounds := srcP.Bounds()
		srcW, srcH := srcPBounds.Dx(), srcPBounds.Dy()
		dstP := scratch(srcP, color.RGBAModel, srcW, srcH)

		// To keep alpha we simply don't convolve it
		switch {
		case keepAlpha:
			// Notice we can't use lenY since it will be larger than the actual padding pixels
			// as it includes the identity element
			prl.Run(srcH-(radiusY*2), func(start, end int) {
				// Correct range so we don't iterate over the padded pixels on the main loop
				for y := start + radiusY; y < end+radiusY; y++ {
					for x := radiusX; x < srcW-radiusX; x++ {
						var r, g, b float64
						// Kernel has access to the padded pixels
						for ky := 0; ky < lenY; ky++ {
							iy := y - radiusY + ky

							for kx := 0; kx < lenX; kx++ {
								ix := x - radiusX + kx

								kvalue := m.At(kx, ky)
								ipos := iy*srcP.str + ix*4
								r += float64(srcP.pix[ipos+0]) * kvalue
								g += float64(srcP.pix[ipos+1]) * kvalue
								b += float64(srcP.pix[ipos+2]) * kvalue
							}
						}

						// Map x and y indices to non-padded range
						pos := (y-radiusY)*dstP.str + (x-radiusX)*4

						dstP.pix[pos+0] = uint8(math.Max(math.Min(r+bias, 255), 0))
						dstP.pix[pos+1] = uint8(math.Max(math.Min(g+bias, 255), 0))
						dstP.pix[pos+2] = uint8(math.Max(math.Min(b+bias, 255), 0))
						dstP.pix[pos+3] = srcP.pix[y*srcP.str+x*4+3]
					}
				}
			})
		default:
			// Notice we can't use lenY since it will be larger than the actual padding pixels
			// as it includes the identity element
			prl.Run(srcH-(radiusY*2), func(start, end int) {
				// Correct range so we don't iterate over the padded pixels on the main loop
				for y := start + radiusY; y < end+radiusY; y++ {
					for x := radiusX; x < srcW-radiusX; x++ {

						var r, g, b, a float64
						// Kernel has access to the padded pixels
						for ky := 0; ky < lenY; ky++ {
							iy := y - radiusY + ky

							for kx := 0; kx < lenX; kx++ {
								ix := x - radiusX + kx

								kvalue := m.At(kx, ky)
								ipos := iy*srcP.str + ix*4
								r += float64(srcP.pix[ipos+0]) * kvalue
								g += float64(srcP.pix[ipos+1]) * kvalue
								b += float64(srcP.pix[ipos+2]) * kvalue
								a += float64(srcP.pix[ipos+3]) * kvalue
							}
						}

						// Map x and y indices to non-padded range
						pos := (y-radiusY)*dstP.str + (x-radiusX)*4

						dstP.pix[pos+0] = uint8(math.Max(math.Min(r+bias, 255), 0))
						dstP.pix[pos+1] = uint8(math.Max(math.Min(g+bias, 255), 0))
						dstP.pix[pos+2] = uint8(math.Max(math.Min(b+bias, 255), 0))
						dstP.pix[pos+3] = uint8(math.Max(math.Min(a, 255), 0))
					}
				}
			})
		}

		return dstP, nil
	})
}

type NoiseFunc func() uint8

type Noiser interface {
	Noise(NoiseFunc, bool) error
}

func (c *canvas) Noise(fn NoiseFunc, monochrome bool) error {
	return c.mutate(func() (*pxl, error) {
		return generateNoise(c.pxl, fn, monochrome)
	})
}

func generateNoise(p *pxl, fn NoiseFunc, monochrome bool) (*pxl, error) {
	return mutate(p, func() (*pxl, error) {
		if !monochrome {
			return generateColorNoise(p, fn)
		}
		return generateMonoNoise(p, fn)
	})
}

func generateColorNoise(p *pxl, nfn NoiseFunc) (*pxl, error) {
	dstP := p.clone(WorkingColorModelFn)

	width, height := dstP.Bounds().Dx(), dstP.Bounds().Dy()

	prl.Run(height, func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < width; x++ {
				pos := y*dstP.str + x*4
				dstP.pix[pos+0] = nfn()
				dstP.pix[pos+1] = nfn()
				dstP.pix[pos+2] = nfn()
				dstP.pix[pos+3] = 0xFF
			}
		}
	})
	return dstP.clone(p.ColorModel()), nil
}

func generateMonoNoise(p *pxl, nfn NoiseFunc) (*pxl, error) {
	dstP := p.clone(WorkingColorModelFn)

	width, height := dstP.Bounds().Dx(), dstP.Bounds().Dy()

	prl.Run(height, func(start, end int) {
		for y := start; y < end; y++ {
			for x := 0; x < width; x++ {
				pos := y*dstP.str + x*4
				v := nfn()
				dstP.pix[pos+0] = v
				dstP.pix[pos+1] = v
				dstP.pix[pos+2] = v
				dstP.pix[pos+3] = 0xFF
			}
		}
	})

	return dstP.clone(p.ColorModel()), nil
}

type Transformer interface {
	Cropper
	Resizer
}

// An interface for cropping a Canvas.
type Cropper interface {
	Crop(image.Rectangle) error
}

var EmptyIntersectError = xrr.Xrror("Unable to crop empty intersect of %v and provided %v").Out

// crops the canvas
func (c *canvas) Crop(r image.Rectangle) error {
	return c.mutate(func() (*pxl, error) {
		return crop(c.pxl, r)
	})
}
func crop(p *pxl, r image.Rectangle) (*pxl, error) {
	exp := r
	r = r.Intersect(p.rect)
	if r.Empty() {
		return p, EmptyIntersectError(p.rect, exp)
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	nr := image.Rectangle{image.Point{0, 0}, r.Size()}
	return &pxl{
		m:         p.m,
		pix:       p.pix[i:],
		str:       p.str,
		rect:      nr,
		paletteFn: p.paletteFn,
		measure:   newMeasure(&r, p.measure.pp, p.measure.ppu),
	}, nil
}

type ResampleFilterFunc func(float64) float64

type ResampleFilter struct {
	Key     string
	Support float64
	Fn      ResampleFilterFunc
}

func (r ResampleFilter) String() string {
	return r.Key
}

var (
	NearestNeighbor = ResampleFilter{
		"nearestneighbor",
		0,
		nil,
	}

	Linear = ResampleFilter{
		"linear",
		1.0,
		func(x float64) float64 {
			x = math.Abs(x)
			if x < 1.0 {
				return 1.0 - x
			}
			return 0
		},
	}
)

type Resizer interface {
	Resize(w, h int, f ResampleFilter) error
}

var ZeroResizeError = xrr.Xrror("zero value prevents resizing\n\twidth %d\n\theight %d\n\tempty canvas: %t").Out

// resize the canvas
func (c *canvas) Resize(w, h int, filter ResampleFilter) error {
	return c.mutate(func() (*pxl, error) {
		return resize(c.pxl, w, h, filter)
	})
}

func resize(p *pxl, w, h int, f ResampleFilter) (*pxl, error) {
	return mutate(p, func() (*pxl, error) {
		b := p.Bounds().Empty()
		if w <= 0 || h <= 0 || b {
			return p, ZeroResizeError(w, h, b)
		}
		switch {
		case f.Support <= 0:
			p = nearest(p, w, h)
		default:
			p = resampleH(p, w, f)
			p = resampleV(p, h, f)
		}
		return p, nil
	})
}

func nearest(p *pxl, w, h int) *pxl {
	srcP := p.clone(color.RGBAModel)
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

	return dstP.clone(p.ColorModel())
}

func resampleH(p *pxl, w int, f ResampleFilter) *pxl {
	srcP := p.clone(WorkingColorModelFn)
	srcWidth, srcHeight := srcP.Bounds().Dx(), srcP.Bounds().Dy()
	srcStride := srcP.Stride()

	delta := float64(srcWidth) / float64(w)
	scale := math.Max(delta, 1.0)

	dstP := scratch(srcP, srcP.ColorModel(), w, srcHeight)
	dstStride := dstP.Stride()

	filterRadius := math.Ceil(scale * f.Support)

	prl.Run(srcHeight, func(start, end int) {
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
				dstP.pix[dstPos+0] = uint8(mth.Clamp((r/sum)+0.5, 0, 255))
				dstP.pix[dstPos+1] = uint8(mth.Clamp((g/sum)+0.5, 0, 255))
				dstP.pix[dstPos+2] = uint8(mth.Clamp((b/sum)+0.5, 0, 255))
				dstP.pix[dstPos+3] = uint8(mth.Clamp((a/sum)+0.5, 0, 255))
			}
		}
	})

	return dstP.clone(p.ColorModel())
}

func resampleV(p *pxl, h int, f ResampleFilter) *pxl {
	srcP := p.clone(WorkingColorModelFn)
	srcWidth, srcHeight := srcP.Bounds().Dx(), srcP.Bounds().Dy()
	srcStride := srcP.Stride()

	delta := float64(srcHeight) / float64(h)
	scale := math.Max(delta, 1.0)

	dstP := scratch(srcP, srcP.ColorModel(), srcWidth, h)
	dstStride := dstP.Stride()

	filterRadius := math.Ceil(scale * f.Support)

	prl.Run(h, func(start, end int) {
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
				dstP.pix[dstPos+0] = uint8(mth.Clamp((r/sum)+0.5, 0, 255))
				dstP.pix[dstPos+1] = uint8(mth.Clamp((g/sum)+0.5, 0, 255))
				dstP.pix[dstPos+2] = uint8(mth.Clamp((b/sum)+0.5, 0, 255))
				dstP.pix[dstPos+3] = uint8(mth.Clamp((a/sum)+0.5, 0, 255))
			}
		}
	})

	return dstP.clone(p.ColorModel())
}

type Translater interface {
	Flip(TDir) error
	Rotate(float64, bool, image.Point) error
	Shear(TDir, float64) error
	Translate(int, int) error
}

type TDir int

const (
	NoTDir TDir = iota
	THorizontal
	TVertical
)

var NoDirectionError = xrr.Xrror("'%s' is not a direction to flip").Out

func (c *canvas) Flip(dir TDir) error {
	return c.mutate(func() (*pxl, error) {
		return flip(c.pxl, dir)
	})
}

func flip(p *pxl, dir TDir) (*pxl, error) {
	return mutate(p, func() (*pxl, error) {
		srcP := p.clone(WorkingColorModelFn)
		dstP := srcP.clone(WorkingColorModelFn)
		b := dstP.Bounds()
		w, h := b.Dx(), b.Dy()
		switch dir {
		case THorizontal:
			prl.Run(h, func(start, end int) {
				for y := start; y < end; y++ {
					for x := 0; x < w; x++ {
						iy := y * dstP.str
						pos := iy + (x * 4)
						flippedX := w - x - 1
						flippedPos := iy + (flippedX * 4)
						dstP.pix[pos+0] = srcP.pix[flippedPos+0]
						dstP.pix[pos+1] = srcP.pix[flippedPos+1]
						dstP.pix[pos+2] = srcP.pix[flippedPos+2]
						dstP.pix[pos+3] = srcP.pix[flippedPos+3]
					}
				}
			})
		case TVertical:
			prl.Run(h, func(start, end int) {
				for y := start; y < end; y++ {
					for x := 0; x < w; x++ {
						pos := y*dstP.str + (x * 4)
						flippedY := h - y - 1
						flippedPos := flippedY*dstP.str + (x * 4)

						dstP.pix[pos+0] = srcP.pix[flippedPos+0]
						dstP.pix[pos+1] = srcP.pix[flippedPos+1]
						dstP.pix[pos+2] = srcP.pix[flippedPos+2]
						dstP.pix[pos+3] = srcP.pix[flippedPos+3]
					}
				}
			})
		default:
			return p, NoDirectionError(dir)
		}
		return dstP, nil
	})
}

func (c *canvas) Rotate(angle float64, preserve bool, at image.Point) error {
	return c.mutate(func() (*pxl, error) {
		return rotate(c.pxl, angle, preserve, at)
	})
}

func rotate(p *pxl, angle float64, preserve bool, at image.Point) (*pxl, error) {
	return mutate(p, func() (*pxl, error) {
		srcP := p.clone(WorkingColorModelFn)
		b := srcP.Bounds()
		srcW, srcH := b.Dx(), b.Dy()

		supersample := false
		absAngle := int(math.Abs(angle) + 0.5)
		if absAngle%360 == 0 {
			return p, nil
		} else if absAngle%90 != 0 {
			// Supersampling is required for non-special angles
			// Special angles = 90, 180, 270...
			supersample = true
		}

		pivotX, pivotY := float64(srcW/2), float64(srcH/2)
		if at != image.ZP {
			pivotX, pivotY = float64(at.X), float64(at.Y)
		}

		var rErr error

		if supersample {
			// Supersample, currently hard set to 2x
			srcW, srcH = srcW*2, srcH*2
			srcP, rErr = resize(srcP, srcW, srcH, NearestNeighbor)
			pivotX, pivotY = pivotX*2, pivotY*2
		}

		if rErr != nil {
			return p, rErr
		}

		// Convert to radians, positive degree maps to clockwise rotation
		angleRadians := -angle * (math.Pi / 180)

		var dstW, dstH int
		if preserve {
			// Reserve larger size in destination image for full image bounds rotation
			// If not preserving size, always take image center as pivot
			pivotX, pivotY = float64(srcW)/2, float64(srcH)/2

			a := math.Abs(float64(srcW) * math.Sin(angleRadians))
			b := math.Abs(float64(srcW) * math.Cos(angleRadians))
			c := math.Abs(float64(srcH) * math.Sin(angleRadians))
			d := math.Abs(float64(srcH) * math.Cos(angleRadians))

			dstW, dstH = int(c+b+0.5), int(a+d+0.5)
		} else {
			dstW, dstH = srcW, srcH
		}

		dstP := scratch(srcP, WorkingColorModelFn, dstW, dstH)

		// Calculate offsets in case entire image is being displayed
		// Otherwise areas clipped by rotation won't be available
		offsetX := (dstW - srcW) / 2
		offsetY := (dstH - srcH) / 2

		prl.Run(srcH, func(start, end int) {
			// Correct range to include the pixels visible in new bounds
			// Note that cannot be done in prl.Runize function input height, otherwise ranges would overlap
			yStart := int((float64(start)/float64(srcH))*float64(dstH)) - offsetY
			yEnd := int((float64(end)/float64(srcH))*float64(dstH)) - offsetY
			xStart := -offsetX
			xEnd := srcW + offsetX

			for y := yStart; y < yEnd; y++ {
				dy := float64(y) - pivotY + 0.5
				for x := xStart; x < xEnd; x++ {
					dx := float64(x) - pivotX + 0.5

					ix := int((math.Cos(angleRadians)*dx - math.Sin(angleRadians)*dy + pivotX))
					iy := int((math.Sin(angleRadians)*dx + math.Cos(angleRadians)*dy + pivotY))

					if ix < 0 || ix >= srcW || iy < 0 || iy >= srcH {
						continue
					}

					srcPos := iy*srcP.str + ix*4
					dstPos := (y+offsetY)*dstP.str + (x+offsetX)*4
					copy(dstP.pix[dstPos:dstPos+4], srcP.pix[srcPos:srcPos+4])
				}
			}
		})

		if supersample {
			// Downsample to original bounds as part of the Supersampling
			dstP, rErr = resize(dstP, dstW/2, dstH/2, Linear)
		}

		if rErr != nil {
			return p, rErr
		}
		return dstP, nil
	})
}

func (c *canvas) Shear(dir TDir, angle float64) error {
	return c.mutate(func() (*pxl, error) {
		return shear(c.pxl, dir, angle)
	})
}

func shear(p *pxl, dir TDir, angle float64) (*pxl, error) {
	return mutate(p, func() (*pxl, error) {
		srcP := p.clone(WorkingColorModelFn)
		b := srcP.Bounds()
		srcW, srcH := b.Dx(), b.Dy()

		// Supersample, currently hard set to 2x
		srcW, srcH = srcW*2, srcH*2

		var sErr error

		srcP, sErr = resize(srcP, srcW, srcH, NearestNeighbor)
		if sErr != nil {
			return p, sErr
		}

		// Calculate shear factor
		k := math.Tan(angle * (math.Pi / 180))

		var dstH, dstW int
		var dstP *pxl

		switch dir {
		case THorizontal:
			dstW, dstH = srcW+int(float64(srcH)*math.Abs(k)), srcH
			dstP = scratch(srcP, WorkingColorModelFn, dstW, dstH)
			pivotX := float64(dstW) / 2
			pivotY := float64(dstH) / 2

			dx := (dstW - srcW) / 2
			dy := (dstH - srcH) / 2
			prl.Run(dstH, func(start, end int) {
				for y := start; y < end; y++ {
					for x := 0; x < dstW; x++ {
						// Move positions to revolve around pivot
						ix := x - int(pivotX) - dx
						iy := y - int(pivotY) - dy

						// Apply linear transformation
						ix = ix + int(float64(iy)*k)

						// Move positions back to image coordinates
						ix += int(pivotX)
						iy += int(pivotY)

						if ix < 0 || ix >= srcW || iy < 0 || iy >= srcH {
							continue
						}

						srcPos := iy*srcP.str + ix*4
						dstPos := y*dstP.str + x*4

						dstP.pix[dstPos+0] = srcP.pix[srcPos+0]
						dstP.pix[dstPos+1] = srcP.pix[srcPos+1]
						dstP.pix[dstPos+2] = srcP.pix[srcPos+2]
						dstP.pix[dstPos+3] = srcP.pix[srcPos+3]
					}
				}
			})
		case TVertical:
			dstW, dstH = srcW, srcH+int(float64(srcW)*math.Abs(k))
			dstP = scratch(srcP, WorkingColorModelFn, dstW, dstH)
			pivotX := float64(dstW) / 2
			pivotY := float64(dstH) / 2
			dx := (dstW - srcW) / 2
			dy := (dstH - srcH) / 2
			prl.Run(dstH, func(start, end int) {
				for y := start; y < end; y++ {
					for x := 0; x < dstW; x++ {
						// Move positions to revolve around pivot
						ix := x - int(pivotX) - dx
						iy := y - int(pivotY) - dy

						// Apply linear transformation
						iy = iy + int(float64(ix)*k)

						// Move positions back to image coordinates
						ix += int(pivotX)
						iy += int(pivotY)

						if ix < 0 || ix >= srcW || iy < 0 || iy >= srcH {
							continue
						}

						srcPos := iy*srcP.str + ix*4
						dstPos := y*dstP.str + x*4

						dstP.pix[dstPos+0] = srcP.pix[srcPos+0]
						dstP.pix[dstPos+1] = srcP.pix[srcPos+1]
						dstP.pix[dstPos+2] = srcP.pix[srcPos+2]
						dstP.pix[dstPos+3] = srcP.pix[srcPos+3]
					}
				}
			})
		default:
			return p, NoDirectionError(dir)
		}

		// Downsample to original bounds as part of the Supersampling
		dstP, sErr = resize(dstP, dstW/2, dstH/2, Linear)
		if sErr != nil {
			return p, sErr
		}

		return dstP, nil
	})
}

func (c *canvas) Translate(dx, dy int) error {
	return c.mutate(func() (*pxl, error) {
		return translate(c.pxl, dx, dy)
	})
}

func translate(p *pxl, dx, dy int) (*pxl, error) {
	return mutate(p, func() (*pxl, error) {
		srcP := p.clone(WorkingColorModelFn)

		if dx == 0 && dy == 0 {
			return p, nil
		}

		b := srcP.Bounds()
		w, h := b.Dx(), b.Dy()

		dstP := scratch(srcP, WorkingColorModelFn, w, h)

		prl.Run(h, func(start, end int) {
			for y := start; y < end; y++ {
				for x := 0; x < w; x++ {
					ix, iy := x-dx, y+dy

					if ix < 0 || ix >= w || iy < 0 || iy >= h {
						continue
					}

					srcPos := iy*srcP.str + ix*4
					dstPos := y*srcP.str + x*4

					copy(dstP.pix[dstPos:dstPos+4], srcP.pix[srcPos:srcPos+4])
				}
			}
		})

		return dstP, nil
	})
}
