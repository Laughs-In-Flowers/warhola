package builtins

import (
	"context"
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
	"github.com/Laughs-In-Flowers/warhola/lib/util"
	"github.com/Laughs-In-Flowers/warhola/lib/util/ctx"
	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func textFlags(o *Options) *flip.FlagSet {
	v := o.Vector
	fs := flip.NewFlagSet("text", flip.ContinueOnError)
	fs.StringVectorVar(v, "fontDirs", "text.fonts.dirs", "", "A comma delimited list of directories containing fonts.")
	fs.StringVectorVar(v, "font", "text.font", "", "The font to use for drawing text.")
	fs.StringVectorVar(v, "message", "text.message", "no message", "The text message to draw")
	fs.StringVectorVar(v, "colorType", "text.color.type", "hex", "The color type specification to use. [hex]")
	fs.StringVectorVar(v, "color", "text.color.value", "FFF", "Font color as a string.")
	fs.StringVectorVar(v, "alignment", "text.align", "left", "Text alignment in text box. [left|center|right]")
	fs.Float64VectorVar(v, "fontSize", "text.font.size", 12, "The font size.")
	fs.Float64VectorVar(v, "lineHeight", "text.line.height", 1, "The line height.")
	fs.IntVector(v, "Width", "text.box.width", "The width dimension of the text box.")
	fs.IntVector(v, "Height", "text.box.height", "The height dimension of the text box.")
	fs.Float64Vector(v, "LX", "text.box.location.X", "The X location point of the text box on the image.")
	fs.Float64Vector(v, "LY", "text.box.location.Y", "The Y location point of the text box on the image.")
	fs.Float64VectorVar(v, "opacity", "text.opacity", 100, "Opacity of the drawn text 1-100")
	fs.Float64Vector(v, "padLeft", "text.box.pad.left", "pads left within text box")
	fs.Float64Vector(v, "padTop", "text.box.pad.top", "pads top within text box")
	fs.Float64Vector(v, "padRight", "text.box.pad.right", "pads right within text box(affects line width for wrapped text only)")
	//fs.Float64Vector(v, "padBottom", "text.box.pad.bottom", "does nothing")
	fs.BoolVectorVar(v, "wrap", "text.wrap", true, "")
	fs.BoolVectorVar(v, "anchor", "text.anchor", false, "")
	return fs
}

func TextDebugReset(o *Options, msg string) {
	o.SetString("text.font", "")
	o.SetString("text.message", msg)
	o.SetString("text.color.type", "hex")
	o.SetString("text.color.value", "FFF")
	o.SetFloat64("text.font.size", 12)
	o.SetFloat64("text.line.height", 1)
	o.SetString("text.align", "left")
	o.SetFloat64("text.opacity", 100)
	o.SetBool("text.wrap", true)
}

var texecuting = buildExecuting("text",
	execution{50, setFonts},
	execution{51, writeTextCtx},
)

var (
	FontsHome  string
	FontsShare = "/usr/share/fonts"
)

func defaultTextDirs() []string {
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

func FontDirectories(dirs string) []string {
	var ret = defaultTextDirs()
	spl := strings.Split(dirs, ",")
	for _, v := range spl {
		if v != "" {
			ret = append(ret, v)
		}
	}
	return ret
}

func setFonts(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	dirs := o.ToString("text.fonts.dirs")
	err := LF.SetDir(FontDirectories(dirs)...)
	if err != nil {
		util.Failure(o, "text", err)
		return c, flip.ExitFailure
	}
	return c, flip.ExitNo
}

func writeTextCtx(o *Options, c context.Context) (context.Context, flip.ExitStatus) {
	cv := ctx.Canvas(c)
	msg := WriteText(cv, o)
	c = context.WithValue(c, 4, cv)
	util.Success(o, "text", fmt.Sprintf("wrote message '%s'", msg))
	return c, flip.ExitNo
}

// Given a canvas and instance of Options, will draw text to the canvas.
func WriteText(cv canvas.Canvas, o *Options) string {
	t := OptionsToText(o)
	t.Draw(cv)
	return t.raw
}

func textCommand() flip.Command {
	return flip.NewCommand(
		"",
		"text",
		"Apply text to an image",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return execute(BuiltInsOptions, c, a, texecuting)
		},
		textFlags(BuiltInsOptions),
	)
}

//
type TextBox struct {
	W, H                     int
	Left, Top, Right, Bottom float64
}

//
func (t *TextBox) BoxImage(cv canvas.Canvas) draw.Image {
	if t.W == 0 {
		t.W = cv.Bounds().Dx()
	}
	if t.H == 0 {
		t.H = cv.Bounds().Dy()
	}
	return canvas.Scratch(cv.ColorModel(), t.W, t.H)
}

//
type TextLocation struct {
	X, Y float64
}

//
func (t *TextLocation) Locate() fixed.Point26_6 {
	return util.Fixp(t.X, t.Y)
}

//
func NewTextLocation(x, y float64) *TextLocation {
	return &TextLocation{x, y}
}

//
type TextFont struct {
	font.Face
	height, lineHeight float64
	alignment          Align
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
type Text struct {
	raw string
	*TextBox
	*TextLocation
	*TextFont
}

// Translates a set of Options to a Text instance.
func OptionsToText(o *Options) *Text {
	bw, bh := o.ToInt("text.box.width"), o.ToInt("text.box.height")
	lx, ly := o.ToFloat64("text.box.location.x"), o.ToFloat64("text.box.location.y")
	pl := o.ToFloat64("text.padLeft")
	pt := o.ToFloat64("text.padTop")
	pb := o.ToFloat64("text.padBottom")
	pr := o.ToFloat64("text.padRight")
	tf := o.ToString("text.font")
	tfz := o.ToFloat64("text.font.size")
	tct := o.ToString("text.color.type")
	tc := o.ToString("text.color.value")
	op := o.ToFloat64("text.opacity")
	lh := o.ToFloat64("text.line.height")
	ta := o.ToString("text.align")
	wr, an := o.ToBool("text.wrap"), o.ToBool("text.anchor")
	msg := o.ToString("text.message")
	return NewText(bw, bh,
		lx, ly,
		pl, pt, pb, pr,
		tf, tfz, tct, tc, op,
		lh, ta, wr, an,
		msg)
}

//
func NewText(
	bw, bh int,
	lx, ly float64,
	pl, pt, pr, pb float64,
	font string,
	fontSize float64,
	fontColorModel, fontColor string,
	opacity float64,
	lineHeight float64,
	alignment string,
	wrap, anchor bool,
	msg string) *Text {
	return &Text{
		TextBox:      &TextBox{bw, bh, pl, pt, pr, pb},
		TextLocation: NewTextLocation(lx, ly),
		TextFont: LF.TextFont(
			font,
			fontSize,
			lineHeight,
			StringToAlign(alignment),
			ToColor(fontColorModel, fontColor),
			opacity,
			wrap, anchor,
		),
		raw: msg,
	}
}

//
func (t *Text) Draw(cv canvas.Canvas) {
	tb := t.BoxImage(cv)
	switch {
	case t.wrap:
		t.DrawStringWrapped(tb)
	case t.anchor:
		t.DrawStringAnchored(tb)
	default:
		t.DrawString(tb)
	}
	cv.Overlay(tb, image.Point{int(t.X), int(t.Y)}, t.opacity)
}

//
func (t *Text) String() string {
	return t.raw
}

func drawer(i draw.Image, tf *TextFont, x, y float64) *font.Drawer {
	return &font.Drawer{
		Dst:  i,
		Src:  image.NewUniform(tf.color),
		Face: tf,
		Dot:  util.Fixp(x, y),
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
	drawStringAnchored(i, t.TextFont, t.raw, 0+t.Left, 0+t.Top, ax, ay)
}

//
func (t *Text) DrawStringWrapped(i draw.Image) {
	ax := 0.0
	ay := 0.0
	width := float64(i.Bounds().Dx()) - t.Right
	drawStringWrapped(i, t.TextFont, t.raw, 0+t.Left, 0+t.Top, ax, ay, width, t.lineHeight, t.alignment)
}

// A type indicating text alignment direction: left, center, right.
type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

// Provides Align base on string, defaulting to AlignLeft
func StringToAlign(s string) Align {
	switch strings.ToLower(s) {
	case "left":
		return AlignLeft
	case "center":
		return AlignCenter
	case "right":
		return AlignRight
	}
	return AlignLeft
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

//
type Fonts struct {
	has map[string]*truetype.Font
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
	a Align,
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
	var ret util.AlphaS
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

func initLocalFonts() {
	LF = &Fonts{make(map[string]*truetype.Font)}
	ft, _ := Asset("ft/DroidSansMono.ttf")
	LF.SetByte("default", ft)
}

func init() {
	initLocalFonts()
}
