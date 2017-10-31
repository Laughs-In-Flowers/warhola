package builtins

import "github.com/Laughs-In-Flowers/flip"

type builtinsMap map[string]func() flip.Command

func (b builtinsMap) Register(tag string, fn func() flip.Command) {
	b[tag] = fn
}

var BuiltIns builtinsMap

func init() {
	BuiltIns = make(builtinsMap)
}
