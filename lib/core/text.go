package core

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util/geo"
	"github.com/Laughs-In-Flowers/warhola/lib/util/mth"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var text = NewCommand(
	"", "text", "Apply text to an image", 1,
	func(o *Options) *flip.FlagSet {
		v := o.Vector
		fs := flip.NewFlagSet("text", flip.ContinueOnError)
		fs.StringVectorVar(v, "fontDirs", "text.fonts.dirs", "", "A comma delimited list of directories containing fonts.")
		geo.GeometryVectorFlag(fs, v, "text.geometry")
		fs.StringVectorVar(v, "message", "text.message", "no message", "The text message to draw")
		fs.StringVectorVar(v, "font", "text.font.name", "", "The font to use for drawing text.")
		fs.StringVectorVar(v, "colorType", "text.color.type", "hex", "The color type specification to use. [hex]")
		fs.StringVectorVar(v, "color", "text.color.value", "FFF", "Font color as a string.")
		fs.StringVectorVar(v, "alignment", "text.align", "left", "Text alignment in text box. [left|center|right]")
		fs.Float64VectorVar(v, "fontSize", "text.font.size", 12, "The font size.")
		fs.Float64VectorVar(v, "lineHeight", "text.line.height", 1, "The line height.")
		fs.Float64VectorVar(v, "opacity", "text.opacity", 100, "Opacity of the drawn text 1-100")
		fs.BoolVectorVar(v, "wrap", "text.wrap", true, "")
		fs.BoolVectorVar(v, "anchor", "text.anchor", false, "")
		return fs
	},
	defaultCommandFunc,
	coreExec(fontStep, textStep)...,
).Command

func fontStep(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	dirs := o.ToString("text.fonts.dirs")
	err := LF.SetDir(FontDirs(dirs)...)
	if err != nil {
		return cv, coreErrorHandler(o, err)
	}
	return cv, flip.ExitNo
}

func textStep(o *Options, cv canvas.Canvas) (canvas.Canvas, flip.ExitStatus) {
	o.Printf("executing text")
	msg := WriteText(cv, o)
	o.Printf("text wrote: %s", msg)
	return cv, flip.ExitNo
}

// Given a canvas and instance of Options, will draw text to the canvas.
func WriteText(cv canvas.Canvas, o *Options) string {
	t := OptionsToText(o)
	t.Scrive(cv)
	return t.String()
}

// Translates a set of Options to a Text instance.
func OptionsToText(o *Options) *Text {
	g := o.pullGeometry("text.geometry")
	bw, bh := g.X, g.Y
	lx, ly := float64(g.OffsetX), float64(g.OffsetY)

	tf := o.ToString("text.font")
	tfz := o.ToFloat64("text.font.size")
	lh := o.ToFloat64("text.line.height")
	ta := o.ToString("text.align")

	tct := o.ToString("text.color.type")
	tc := ToColor(tct, o.ToString("text.color.value"))
	op := o.ToFloat64("text.opacity")

	wr, an := o.ToBool("text.wrap"), o.ToBool("text.anchor")

	msg := o.ToString("text.message")

	return NewText(
		msg,
		bw, bh,
		lx, ly,
		tf, tfz, tc, op,
		lh, ta, wr, an,
	)
}

//
type Text struct {
	raw string
	*TextBox
	*TextLocation
	*TextFont
}

//
func NewText(
	msg string,
	bw, bh int,
	lx, ly float64,
	font string,
	fontSize float64,
	fontColor color.Color,
	opacity float64,
	lineHeight float64,
	alignment string,
	wrap, anchor bool) *Text {
	b := NewTextBox(bw, bh)
	l := NewTextLocation(lx, ly)
	f := LF.TextFont(
		font,
		fontSize,
		lineHeight,
		stringToAlign(alignment),
		fontColor,
		opacity,
		wrap, anchor)
	return newText(msg, b, l, f)
}

func newText(msg string, b *TextBox, l *TextLocation, f *TextFont) *Text {
	return &Text{msg, b, l, f}
}

//
func (t *Text) String() string {
	return t.raw
}

//
func (t *Text) Scrive(d draw.Image) {
	bx := t.Box(d)
	switch {
	case t.wrap:
		t.DrawStringWrapped(bx)
	case t.anchor:
		t.DrawStringAnchored(bx)
	default:
		t.DrawString(bx)
	}
	mask := image.NewUniform(color.Alpha{uint8(255 * (t.opacity / 100))})
	draw.DrawMask(
		d,
		d.Bounds(),
		bx,
		t.Point(),
		mask,
		image.ZP,
		draw.Over,
	)
}

func drawer(i draw.Image, tf *TextFont, x, y float64) *font.Drawer {
	return &font.Drawer{
		Dst:  i,
		Src:  image.NewUniform(tf.color),
		Face: tf,
		Dot:  mth.Fixp(x, y),
	}
}

func drawString(i draw.Image, tf *TextFont, s string, x, y float64) {
	d := drawer(i, tf, x, y)
	d.DrawString(s)
}

//
func (t *Text) DrawString(i draw.Image) {
	drawString(i, t.TextFont, t.raw, 0, 0)
}

func drawStringAnchored(i draw.Image, tf *TextFont, s string, x, y, ax, ay float64) {
	w, h := tf.MeasureString(s)
	x -= ax * w
	y += ay * h
	drawString(i, tf, s, x, y)
}

//
func (t *Text) DrawStringAnchored(i draw.Image) {
	ax := 0.0
	ay := 0.0
	drawStringAnchored(i, t.TextFont, t.raw, 0, 0, ax, ay)
}

//
func (t *Text) DrawStringWrapped(i draw.Image) {
	ax := 0.0
	ay := 0.0
	width := float64(i.Bounds().Dx())
	drawStringWrapped(i, t.TextFont, t.raw, 0, 0, ax, ay, width, t.lineHeight, t.alignment)
}

// A type indicating text alignment direction: left, center, right.
type align int

const (
	aLeft align = iota
	aCenter
	aRight
)

func stringToAlign(s string) align {
	switch strings.ToLower(s) {
	case "left":
		return aLeft
	case "center":
		return aCenter
	case "right":
		return aRight
	}
	return aLeft
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

func drawStringWrapped(
	i draw.Image,
	tf *TextFont,
	s string,
	x, y, ax, ay, width, lineSpacing float64,
	a align,
) {
	lines := wordWrap(tf, s, width)
	h := float64(len(lines)) * tf.height * lineSpacing
	h -= (lineSpacing - 1) * tf.height
	x -= ax * width
	y -= ay * h
	switch a {
	case aLeft:
		ax = 0
	case aCenter:
		ax = 0.5
		x += width / 2
	case aRight:
		ax = 1
		x += width
	}
	ay = 1
	for _, line := range lines {
		drawStringAnchored(i, tf, line, x, y, ax, ay)
		y += tf.height * lineSpacing
	}
}

//
type TextBox struct {
	W, H int
}

//
func NewTextBox(bw, bh int) *TextBox {
	return &TextBox{bw, bh}
}

//
func (t *TextBox) Box(d draw.Image) draw.Image {
	if t.W == 0 {
		t.W = d.Bounds().Dx()
	}
	if t.H == 0 {
		t.H = d.Bounds().Dy()
	}
	return canvas.Scratch(d.ColorModel(), t.W, t.H)
}

//
type TextLocation struct {
	X, Y float64
}

//
func NewTextLocation(x, y float64) *TextLocation {
	return &TextLocation{x, y}
}

//
func (t *TextLocation) Locate() fixed.Point26_6 {
	return mth.Fixp(t.X, t.Y)
}

//
func (t *TextLocation) Point() image.Point {
	return image.Point{int(t.X), int(t.Y)}
}

//
type TextFont struct {
	font.Face
	height, lineHeight float64
	alignment          align
	color              color.Color
	opacity            float64
	wrap, anchor       bool
}

//
func (t *TextFont) MeasureString(s string) (w, h float64) {
	d := &font.Drawer{
		Face: t.Face,
	}
	a := d.MeasureString(s)
	return float64(a >> 6), t.height
}

//
type Fonts struct {
	has map[string]*truetype.Font
}

//
func NewFonts() *Fonts {
	return &Fonts{make(map[string]*truetype.Font)}
}

//
func (f *Fonts) Get(k string) *truetype.Font {
	if gf, exists := f.has[k]; exists {
		return gf
	}
	ft := f.has["default"]
	return ft
}

//
func (f *Fonts) TextFont(k string,
	height, lineHeight float64,
	a align,
	c color.Color,
	o float64,
	wrap, anchor bool) *TextFont {
	ft := f.Get(k)

	face := truetype.NewFace(ft, &truetype.Options{
		Size:    height,
		Hinting: font.HintingFull,
	})

	return &TextFont{face, height, lineHeight, a, c, o, wrap, anchor}
}

func parseFont(f []byte) (*truetype.Font, error) {
	ft, err := truetype.Parse(f)
	if err != nil {
		return nil, err
	}
	return ft, nil
}

//
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

//
func (f *Fonts) SetPath(name, path string) bool {
	if ft, err := loadFont(path); err == nil {
		return f.Set(name, ft)
	}
	return false
}

//
func (f *Fonts) SetByte(name string, fb []byte) bool {
	if ff, err := parseFont(fb); err == nil {
		return f.Set(name, ff)
	}
	return false
}

//
func (f *Fonts) List() []string {
	var ret AlphaS
	for k, _ := range f.has {
		ret = append(ret, k)
	}
	sort.Sort(ret)
	return ret
}

//
func (f *Fonts) SetDir(paths ...string) error {
	for _, path := range paths {
		dir, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		for _, v := range dir {
			switch {
			case v.IsDir():
				err = f.SetDir(filepath.Join(path, v.Name()))
				if err != nil {
					return err
				}
			default:
				fn := v.Name()
				rfn := strings.Split(fn, ".")[0]
				fnp, _ := filepath.Abs(filepath.Join(path, fn))
				f.SetPath(rfn, fnp)
			}
		}
	}
	return nil
}

// local fonts
var LF *Fonts

var (
	FontsHome  string
	FontsShare = "/usr/share/fonts"
)

func defaultFontDirs() []string {
	h := os.Getenv("HOME")
	homeF := fmt.Sprintf("%s/.fonts", h)
	var err error
	FontsHome, err = filepath.Abs(homeF)
	if err != nil {
		FontsHome = "./fonts"
	}
	dirs := []string{FontsHome, FontsShare}
	return dirs
}

func FontDirs(dirs string) []string {
	var ret = defaultFontDirs()
	spl := strings.Split(dirs, ",")
	for _, v := range spl {
		if v != "" {
			ret = append(ret, v)
		}
	}
	return ret
}

func initCoreFonts() {
	LF = NewFonts()
	ft, _ := Asset("ft/DroidSansMono.ttf")
	LF.SetByte("default", ft)
}
