package builtins

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"strings"
	"unicode"

	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type TextLocation struct {
	X, Y float64
}

func (t *TextLocation) Locate() fixed.Point26_6 {
	return fixp(t.X, t.Y)
}

func NewTextLocation(x, y float64) *TextLocation {
	return &TextLocation{x, y}
}

type TextFont struct {
	font.Face
	tag    string
	height float64
	color  color.Color
}

func (t *TextFont) MeasureString(s string) (w, h float64) {
	d := &font.Drawer{
		Face: t.Face,
	}
	a := d.MeasureString(s)
	return float64(a >> 6), t.height
}

type Fonts struct {
	has map[string]*truetype.Font
}

func (f *Fonts) Get(k string) *truetype.Font {
	if gf, exists := f.has[k]; exists {
		return gf
	}
	ft := f.has["default"]
	return ft
}

func (f *Fonts) TextFont(k string, z float64, c color.Color) *TextFont {
	ft := f.Get(k)

	face := truetype.NewFace(ft, &truetype.Options{
		Size: z,
		//Hinting: font.HintingFull,
	})

	return &TextFont{face, k, z, c}
}

func parseFont(f []byte) (*truetype.Font, error) {
	ft, err := truetype.Parse(f)
	if err != nil {
		return nil, err
	}
	return ft, nil
}

func (f *Fonts) Set(name string, ft *truetype.Font) bool {
	f.has[name] = ft
	return true
}

func loadFont(path string) (*truetype.Font, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseFont(f)
}

func (f *Fonts) SetPath(name, path string) bool {
	if ft, err := loadFont(path); err == nil {
		return f.Set(name, ft)
	}
	return false
}

func (f *Fonts) SetByte(name string, fb []byte) bool {
	if ff, err := parseFont(fb); err == nil {
		return f.Set(name, ff)
	}
	return false
}

func splitOnSpace(x string) []string {
	var result []string
	pi := 0
	ps := false
	for i, c := range x {
		s := unicode.IsSpace(c)
		if s != ps && i > 0 {
			result = append(result, x[pi:i])
			pi = i
		}
		ps = s
	}
	result = append(result, x[pi:])
	return result
}

func wordWrap(t *TextFont, s string, width float64) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		fields := splitOnSpace(line)
		if len(fields)%2 == 1 {
			fields = append(fields, "")
		}
		x := ""
		for i := 0; i < len(fields); i += 2 {
			w, _ := t.MeasureString(x + fields[i])
			if w > width {
				if x == "" {
					result = append(result, fields[i])
					x = ""
					continue
				} else {
					result = append(result, x)
					x = ""
				}
			}
			x += fields[i] + fields[i+1]
		}
		if x != "" {
			result = append(result, x)
		}
	}
	for i, line := range result {
		result[i] = strings.TrimSpace(line)
	}
	return result
}

type Text struct {
	*TextLocation
	*TextFont
	raw string
}

func NewText(x, y float64, f string, z float64, c color.Color, txt string) *Text {
	return &Text{
		TextLocation: NewTextLocation(x, y),
		TextFont:     Local.TextFont(f, z, c),
		raw:          txt,
	}
}

func (t *Text) String() string {
	return t.raw
}

func drawer(i draw.Image, tf *TextFont, x, y float64) *font.Drawer {
	return &font.Drawer{
		Dst:  i,
		Src:  image.NewUniform(tf.color),
		Face: tf,
		Dot:  fixp(x, y),
	}
}

func drawString(i draw.Image, tf *TextFont, s string, x, y float64) {
	d := drawer(i, tf, x, y)
	d.DrawString(s)
}

func (t *Text) Draw(i draw.Image) {
	drawString(i, t.TextFont, t.raw, t.X, t.Y)
}

func drawStringAnchored(i draw.Image, tf *TextFont, s string, x, y, ax, ay float64) {
	w, h := tf.MeasureString(s)
	x -= ax * w
	y += ay * h
	drawString(i, tf, s, x, y)
}

func (t *Text) DrawStringAnchored(i draw.Image, ax, ay float64) {
	drawStringAnchored(i, t.TextFont, t.raw, t.X, t.Y, ax, ay)
}

type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

func (t *Text) DrawStringWrapped(i draw.Image, ax, ay, width, lineSpacing float64, a Align) {
	drawStringWrapped(i, t.TextFont, t.raw, t.X, t.Y, ax, ay, width, lineSpacing, a)
}

func drawStringWrapped(
	i draw.Image,
	tf *TextFont,
	s string,
	x, y, ax, ay, width, lineSpacing float64,
	a Align,
) {
	lines := wordWrap(tf, s, width)
	h := float64(len(lines)) * tf.height * lineSpacing
	h -= (lineSpacing - 1) * tf.height
	x -= ax * width
	y -= ay * h
	switch a {
	case AlignLeft:
		ax = 0
	case AlignCenter:
		ax = 0.5
		x += width / 2
	case AlignRight:
		ax = 1
		x += width
	}
	ay = 1
	for _, line := range lines {
		drawStringAnchored(i, tf, line, x, y, ax, ay)
		y += tf.height * lineSpacing
	}
}

var Local *Fonts

func init() {
	Local = &Fonts{make(map[string]*truetype.Font)}
	ft, _ := Asset("ft/DroidSansMono.ttf")
	Local.SetByte("default", ft)
}
