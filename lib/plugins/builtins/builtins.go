package builtins

import "image/draw"

type builtinsMap map[string]func(draw.Image, bool, ...string) (draw.Image, error)

func (b builtinsMap) Register(tag string, fn func(draw.Image, bool, ...string) (draw.Image, error)) {
	b[tag] = fn
}

var BuiltIns builtinsMap

func init() {
	BuiltIns = make(builtinsMap)
}
