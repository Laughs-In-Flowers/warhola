package main

import (
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"strings"

	"github.com/Laughs-In-Flowers/warhola/lib/text"
)

type options struct {
	message  string
	fontSize float64
	color    color.Color
}

func defaults() *options {
	return &options{
		"NO MESSAGE", 30, color.White,
	}
}

func extract(args []string) *options {
	o := defaults()
	for _, a := range args {
		as := strings.Split(a, ":")
		if len(as) == 2 {
			k, v := as[0], as[1]
			switch k {
			case "message":
				o.message = v
			case "font_size":
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					o.fontSize = f
				}
			case "color":
				list := strings.Split(v, ",")
				if len(list) == 4 {
					rr, gg, bb, aa := list[0], list[1], list[2], list[3]
					var rgb []uint8
					for _, v := range []string{rr, gg, bb, aa} {
						if s, err := strconv.ParseUint(v, 10, 8); err == nil {
							rgb = append(rgb, uint8(s))
						}
					}
					o.color = color.Color(color.RGBA{rgb[0], rgb[1], rgb[2], rgb[3]})
				}
			}
		}
	}
	return o
}

func clone(src draw.Image) draw.Image {
	bounds := src.Bounds()
	img := image.NewRGBA64(bounds)
	draw.Draw(img, bounds, src, bounds.Min, draw.Src)
	return img
}

var PluginName = "TextOnImage"

func Apply(i draw.Image, debug bool, args ...string) (draw.Image, error) {
	si := clone(i)
	opts := extract(args)
	t := text.NewText(10, 10, "default", opts.fontSize, opts.color, opts.message)
	t.Draw(si)
	return si, nil
}
