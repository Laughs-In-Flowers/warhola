package builtins

import (
	"context"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/xrr"
)

func cropInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	req := o.ToString("crop.rect")
	rect := stringToRect(req)
	cv.Printf("cropping %s", rect)
	cv.CropTo(rect)
	return cv, flip.ExitNo
}

var (
	cropExecuting = buildExecuting("crop", cropx)
	cropx         = execution{50, cropStep}
	cropStep      = builtInMainStep(cropInner)
)

func cropFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("crop", flip.ContinueOnError)
	fs.StringVectorVar(v, "rectangle", "crop.rect", "0,0,0,0", "a rectangle to crop the image to")
	return fs
}

func cropCommand() flip.Command {
	return flip.NewCommand(
		"",
		"crop",
		"crop an image",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, cropExecuting)
		},
		cropFlags(BuiltInsOptions),
	)
}

var ZeroResizeError = xrr.Xrror("unable to resize with any zero value (width %d height %d empty canvas: %t)").Out

func resize(p canvas.Pxl, width, height int, filter string) error {
	//pb := p.Bounds().Empty()
	//if width <= 0 || height <= 0 || pb {
	//	return ZeroResizeError(width, height, pb)
	//}

	//f := stringToFilter(filter)
	//switch {
	//case f.Support <= 0:
	//	nearest(p, width, height)
	//default:
	//	resampleH(p, width, f)
	//	resampleV(p, height, f)
	//}

	//p.Resize(image.Rect(0, 0, width, height))

	return nil
}

func resizeInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	w, h := o.ToInt("resize.width"), o.ToInt("resize.height")
	filter := o.ToString("resize.filter")
	err := cv.ResizeTo(w, h, filter)
	return cv, builtInErrorHandler(o, err)
}

var (
	resizeExecuting = buildExecuting("resize", resizex)
	resizex         = execution{50, resizeStep}
	resizeStep      = builtInMainStep(resizeInner)
)

func resizeFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("resize", flip.ContinueOnError)
	fs.IntVector(v, "height", "resize.height", "the height of the resized image")
	fs.IntVector(v, "width", "resize.width", "the width of the resized image")
	fs.StringVectorVar(v, "filter", "resize.filter", "nearest", "the resample filter to use in resizing\n\t\t[nearest|\n\t\tbox|\n\t\tlinear|\n\t\tgaussian|\n\t\tmitchellnetravali|\n\t\tcatmullrom|\n\t\tlanczos|\n\t\tbartlett|\n\t\thermite|\n\t\tbspline|\n\t\thann|\n\t\thamming|\n\t\tblackman|\n\t\twelch|\n\t\tcosine]")
	return fs
}

func resizeCommand() flip.Command {
	return flip.NewCommand(
		"",
		"resize",
		"Resize an image",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, resizeExecuting)
		},
		resizeFlags(BuiltInsOptions),
	)
}

func rotate(src canvas.Canvas, angle float64, resizeBounds bool, pivot canvas.Moint) error {
	/*
		srcW, srcH := src.Bounds().Dx(), src.Bounds().Dy()

		supersample := false
		absAngle := int(math.Abs(angle) + 0.5)
		if absAngle%360 == 0 {
			// Return early if nothing to do
			return nil
		} else if absAngle%90 != 0 {
			// Supersampling is required for non-special angles
			// Special angles = 90, 180, 270...
			supersample = true
		}

		// Config defaults
		resizeBounds = false
		// Default pivot position is center of image
		pivotX, pivotY := float64(srcW/2), float64(srcH/2)

		if supersample {
			// Supersample, currently hard set to 2x
			srcW, srcH = srcW*2, srcH*2
			resize(src, srcW, srcH, "NearestNeighbor")
			pivotX, pivotY = pivotX*2, pivotY*2
		}

		// Convert to radians, positive degree maps to clockwise rotation
		angleRadians := -angle * (math.Pi / 180)

		var dstW, dstH int
		if resizeBounds {
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

		dst := canvas.Scratch(src.ColorModel(), dstW, dstH) //image.NewRGBA(image.Rect(0, 0, dstW, dstH))

		// Calculate offsets in case entire image is being displayed
		// Otherwise areas clipped by rotation won't be available
		offsetX := (dstW - srcW) / 2
		offsetY := (dstH - srcH) / 2

		parallel(srcH, func(start, end int) {
			srcPx := src.Pix()
			dstPx := dst.Pix()
			// Correct range to include the pixels visible in new bounds
			// Note that cannot be done in parallelize function input height, otherwise ranges would overlap
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

					srcPos := iy*src.Stride() + ix*4
					dstPos := (y+offsetY)*dst.Stride() + (x+offsetX)*4
					copy(dstPx[dstPos:dstPos+4], srcPx[srcPos:srcPos+4])
				}
			}
			src.SetPix(dstPx)
		})

		if supersample {
			// Downsample to original bounds as part of the Supersampling
			resize(dst, dstW/2, dstH/2, "Linear")
		}
	*/
	return nil
}

func rotateInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	angle := o.ToFloat64("rotate.angle")
	preserve := o.ToBool("rotate.preserve")
	x, y := o.ToFloat64("rotate.pivot.x"), o.ToFloat64("rotate.pivot.y")
	err := rotate(cv, angle, preserve, canvas.Moint{x, y})
	return cv, builtInErrorHandler(o, err)
}

var (
	rotateExecuting = buildExecuting("rotate", rotatex)
	rotatex         = execution{50, rotateStep}
	rotateStep      = builtInMainStep(rotateInner)
)

func rotateFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("rotate", flip.ContinueOnError)
	fs.Float64Vector(v, "angle", "rotate.angle", "the angle of rotation to apply")
	fs.BoolVector(v, "preserveSize", "rotate.preserve", "preserve the size the image")
	fs.Float64Vector(v, "pivotX", "rotate.pivot.x", "the x value of the rotation pivot point")
	fs.Float64Vector(v, "pivotY", "rotate.pivot.y", "the y value of the rotation pivot point")
	return fs
}

func rotateCommand() flip.Command {
	return flip.NewCommand(
		"",
		"rotate",
		"Rotate an image",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, rotateExecuting)
		},
		rotateFlags(BuiltInsOptions),
	)
}

func flipV(src canvas.Canvas) error {
	/*
		bounds := img.Bounds()
			src := clone.AsRGBA(img)
			dst := image.NewRGBA(bounds)
			w, h := dst.Bounds().Dx(), dst.Bounds().Dy()

			parallel(h, func(start, end int) {
				for y := start; y < end; y++ {
					for x := 0; x < w; x++ {
						pos := y*dst.Stride + (x * 4)
						flippedY := h - y - 1
						flippedPos := flippedY*dst.Stride + (x * 4)

						dst.Pix[pos+0] = src.Pix[flippedPos+0]
						dst.Pix[pos+1] = src.Pix[flippedPos+1]
						dst.Pix[pos+2] = src.Pix[flippedPos+2]
						dst.Pix[pos+3] = src.Pix[flippedPos+3]
					}
				}
			})

		return dst
	*/
	return nil
}

func flipH(src canvas.Canvas) error {
	/*
		bounds := img.Bounds()
			src := clone.AsRGBA(img)
			dst := image.NewRGBA(bounds)
			w, h := dst.Bounds().Dx(), dst.Bounds().Dy()

			parallel.Line(h, func(start, end int) {
				for y := start; y < end; y++ {
					for x := 0; x < w; x++ {
						iy := y * dst.Stride
						pos := iy + (x * 4)
						flippedX := w - x - 1
						flippedPos := iy + (flippedX * 4)

						dst.Pix[pos+0] = src.Pix[flippedPos+0]
						dst.Pix[pos+1] = src.Pix[flippedPos+1]
						dst.Pix[pos+2] = src.Pix[flippedPos+2]
						dst.Pix[pos+3] = src.Pix[flippedPos+3]
					}
				}
			})

		return dst
	*/
	return nil
}

func flipCV(o *Options, cv canvas.Canvas, fn func(canvas.Canvas) error) error {
	err := fn(cv)
	if err != nil {
		o.Println(err)
		return err
	}
	return nil
}

func flipInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	v, h := o.ToBool("flip.vertical"), o.ToBool("flip.horizontal")
	var err error
	if v {
		err = flipCV(o, cv, flipV)
	}
	if h {
		err = flipCV(o, cv, flipH)
	}
	return cv, builtInErrorHandler(o, err)
}

var (
	flipExecuting = buildExecuting("flip", flipx)
	flipx         = execution{50, flipStep}
	flipStep      = builtInMainStep(flipInner)
)

func flipFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("flip", flip.ContinueOnError)
	fs.BoolVector(v, "vertical", "flip.vertical", "flip the image vertically")
	fs.BoolVector(v, "horizontal", "flip.horizontal", "flip the image horizontally")
	return fs
}

func flipCommand() flip.Command {
	return flip.NewCommand(
		"",
		"flip",
		"Flip an image in an opposite direction",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, flipExecuting)
		},
		flipFlags(BuiltInsOptions),
	)
}

func shearV(cv canvas.Canvas, angle float64) error {
	/*
		src := clone.AsRGBA(img)
			srcW, srcH := src.Bounds().Dx(), src.Bounds().Dy()

			// Supersample, currently hard set to 2x
			srcW, srcH = srcW*2, srcH*2
			src = Resize(src, srcW, srcH, NearestNeighbor)

			// Calculate shear factor
			ky := math.Tan(angle * (math.Pi / 180))

			dstW, dstH := srcW, srcH+int(float64(srcW)*math.Abs(ky))
			dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

			pivotX := float64(dstW) / 2
			pivotY := float64(dstH) / 2

			// Calculate offset since we are resizing the bounds to
			// fit the sheared image.
			dx := (dstW - srcW) / 2
			dy := (dstH - srcH) / 2

			parallel.Line(dstH, func(start, end int) {
				for y := start; y < end; y++ {
					for x := 0; x < dstW; x++ {
						// Move positions to revolve around pivot
						ix := x - int(pivotX) - dx
						iy := y - int(pivotY) - dy

						// Apply linear transformation
						iy = iy + int(float64(ix)*ky)

						// Move positions back to image coordinates
						ix += int(pivotX)
						iy += int(pivotY)

						if ix < 0 || ix >= srcW || iy < 0 || iy >= srcH {
							continue
						}

						srcPos := iy*src.Stride + ix*4
						dstPos := y*dst.Stride + x*4

						dst.Pix[dstPos+0] = src.Pix[srcPos+0]
						dst.Pix[dstPos+1] = src.Pix[srcPos+1]
						dst.Pix[dstPos+2] = src.Pix[srcPos+2]
						dst.Pix[dstPos+3] = src.Pix[srcPos+3]
					}
				}
			})

			// Downsample to original bounds as part of the Supersampling
			dst = Resize(dst, dstW/2, dstH/2, Linear)

		return dst
	*/
	return nil
}

func shearH(cv canvas.Canvas, angle float64) error {
	/*
			src := clone.AsRGBA(img)
			srcW, srcH := src.Bounds().Dx(), src.Bounds().Dy()

			// Supersample, currently hard set to 2x
			srcW, srcH = srcW*2, srcH*2
			src = Resize(src, srcW, srcH, NearestNeighbor)

			// Calculate shear factor
			kx := math.Tan(angle * (math.Pi / 180))

			dstW, dstH := srcW+int(float64(srcH)*math.Abs(kx)), srcH
			dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

			pivotX := float64(dstW) / 2
			pivotY := float64(dstH) / 2

			// Calculate offset since we are resizing the bounds to
			// fit the sheared image.
			dx := (dstW - srcW) / 2
			dy := (dstH - srcH) / 2

			parallel.Line(dstH, func(start, end int) {
				for y := start; y < end; y++ {
					for x := 0; x < dstW; x++ {
						// Move positions to revolve around pivot
						ix := x - int(pivotX) - dx
						iy := y - int(pivotY) - dy

						// Apply linear transformation
						ix = ix + int(float64(iy)*kx)

						// Move positions back to image coordinates
						ix += int(pivotX)
						iy += int(pivotY)

						if ix < 0 || ix >= srcW || iy < 0 || iy >= srcH {
							continue
						}

						srcPos := iy*src.Stride + ix*4
						dstPos := y*dst.Stride + x*4

						dst.Pix[dstPos+0] = src.Pix[srcPos+0]
						dst.Pix[dstPos+1] = src.Pix[srcPos+1]
						dst.Pix[dstPos+2] = src.Pix[srcPos+2]
						dst.Pix[dstPos+3] = src.Pix[srcPos+3]
					}
				}
			})

			// Downsample to original bounds as part of the Supersampling
			dst = Resize(dst, dstW/2, dstH/2, Linear)

		return dst
	*/
	return nil
}

func shearInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	v, h := o.ToFloat64("shear.vertical.angle"), o.ToFloat64("shear.horizontal.angle")
	var err error
	if v != 0 {
		err = shearV(cv, v)
	}
	if h != 0 {
		err = shearH(cv, h)
	}
	return cv, builtInErrorHandler(o, err)
}

var (
	shearExecuting = buildExecuting("shear", shearx)
	shearx         = execution{50, shearStep}
	shearStep      = builtInMainStep(shearInner)
)

func shearFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("shear", flip.ContinueOnError)
	fs.Float64Vector(v, "vertical", "shear.vertical.angle", "the angle of vertical shear to apply")
	fs.Float64Vector(v, "horizontal", "shear.horizontal.angle", "the angle of horizontal  shear to apply")
	return fs
}

func shearCommand() flip.Command {
	return flip.NewCommand(
		"",
		"shear",
		"Linear transformation along an axis",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, shearExecuting)
		},
		shearFlags(BuiltInsOptions),
	)
}

func translate(cv canvas.Canvas, dx, dy int) error {
	/*
		src := clone.AsRGBA(img)

		if dx == 0 && dy == 0 {
			return src
		}

		w, h := src.Bounds().Dx(), src.Bounds().Dy()
		dst := image.NewRGBA(src.Bounds())

		parallel.Line(h, func(start, end int) {
			for y := start; y < end; y++ {
				for x := 0; x < w; x++ {
					ix, iy := x-dx, y+dy

					if ix < 0 || ix >= w || iy < 0 || iy >= h {
						continue
					}

					srcPos := iy*src.Stride + ix*4
					dstPos := y*src.Stride + x*4

					copy(dst.Pix[dstPos:dstPos+4], src.Pix[srcPos:srcPos+4])
				}
			}
		})

		return dst
	*/
	return nil
}

func translateInner(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	dx, dy := o.ToInt("translate.dx"), o.ToInt("translate.dy")
	err := translate(cv, dx, dy)
	return cv, builtInErrorHandler(o, err)
}

var (
	translateExecuting = buildExecuting("translate", translatex)
	translatex         = execution{50, translateStep}
	translateStep      = builtInMainStep(translateInner)
)

func translateFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("translate", flip.ContinueOnError)
	fs.IntVector(v, "dx", "translate.dx", "reposition by dx on the x-axis")
	fs.IntVector(v, "dy", "translate.dy", "reposition by dy on the y-axis")
	return fs
}

func translateCommand() flip.Command {
	return flip.NewCommand(
		"",
		"translate",
		"repositions a copy of an image by dx on the x-axis and dy on the y-axis",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, translateExecuting)
		},
		translateFlags(BuiltInsOptions),
	)
}
