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
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/warhola/lib/canvas"
	"github.com/Laughs-In-Flowers/warhola/lib/util"
	"github.com/Laughs-In-Flowers/warhola/lib/util/ctx"
	"github.com/fatih/structs"
	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	FontsHome          string
	FontsShare         = "/usr/share/fonts"
	defaultTextOptions *TextOptions
)

type TextOptions struct {
	Font, Msg, Color, ColorType, Align   string
	FontSize, LineHeight, Opacity        float64
	PadLeft, PadTop, PadRight, PadBottom float64
	Bx, By                               int
	Lx, Ly                               float64
	*TextDirs
}

func defaultTOptions() *TextOptions {
	return &TextOptions{
		Font:       "",
		Msg:        "no message",
		Color:      "FFF",
		ColorType:  "hex",
		Align:      "left",
		FontSize:   12,
		LineHeight: 1,
		Opacity:    100,
		TextDirs:   NewTextDirs(),
	}
}

func NewTextOptions(font, msg, color, align string, tbx, tby int, lx, ly float64) *TextOptions {
	o := defaultTOptions()
	o.Font = font
	o.Msg = msg
	o.Color = color
	o.Align = align
	o.Bx, o.By = tbx, tby
	o.Lx, o.Ly = lx, ly
	return o
}

type Color struct {
	R, G, B float64
}

func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R*65535.0 + 0.5)
	g = uint32(c.G*65535.0 + 0.5)
	b = uint32(c.B*65535.0 + 0.5)
	a = 0xFFFF
	return
}

// Hex parses a "html" hex color-string, either in the 3 "#f0c", 3 "f0c",
// 6 "ff1034", or the default 6 "#ff1034" digits form.
func hex(scol string) (Color, error) {
	var format string
	var factor float64
	var f3 string = "#%1x%1x%1x"
	var fa3 = 1.0 / 15.0
	var f6 string = "#%02x%02x%02x"
	var fa6 = 1.0 / 255.0
	switch len(scol) {
	case 3:
		scol = fmt.Sprintf("#%s", scol)
		format = f3
		factor = fa3
	case 4:
		format = f3
		factor = fa3
	case 6:
		scol = fmt.Sprintf("#%s", scol)
		format = f6
		factor = fa6
	default:
		format = f6
		factor = fa6
	}

	var r, g, b uint8
	n, err := fmt.Sscanf(scol, format, &r, &g, &b)
	if err != nil {
		return Color{}, err
	}
	if n != 3 {
		return Color{}, fmt.Errorf("color: %v is not a hex-color", scol)
	}

	return Color{float64(r) * factor, float64(g) * factor, float64(b) * factor}, nil
}

func (t *TextOptions) ColorV() color.Color {
	var col color.Color
	var err error
	switch t.ColorType {
	case "hex":
		col, err = hex(t.Color)
	default:
		col = color.White
	}
	if err != nil {
		return color.White
	}
	return col
}

type TextDirs struct {
	Dirs string
}

func NewTextDirs() *TextDirs {
	h := os.Getenv("HOME")
	homeF := fmt.Sprintf("%s/.fonts", h)
	var err error
	FontsHome, err = filepath.Abs(homeF)
	if err != nil {
		FontsHome = "./fonts"
	}

	dirs := strings.Join([]string{FontsHome, FontsShare}, ",")

	return &TextDirs{dirs}
}

func (t *TextDirs) Directories() []string {
	var ret []string
	spl := strings.Split(t.Dirs, ",")
	for _, v := range spl {
		if v != "" {
			ret = append(ret, v)
		}
	}
	return ret
}

func tFlags(fs *flip.FlagSet, o *TextOptions) {
	fs.StringVar(&o.Dirs, "fontDirs", o.Dirs, "comma delimited list of directories containing fonts")
	fs.StringVar(&o.Font, "font", o.Font, "the font to use for drawing text")
	fs.StringVar(&o.Msg, "message", o.Msg, "the text message to draw")
	fs.StringVar(&o.Color, "color", o.Color, "font color as a string, in either 3 or 6 characters with or without a preceding #")
	fs.StringVar(&o.Align, "alignment", o.Align, "text alignment in text box [left|center|right]")
	fs.Float64Var(&o.FontSize, "fontSize", o.FontSize, "the font size")
	fs.Float64Var(&o.LineHeight, "lineHeight", o.LineHeight, "line height")
	fs.IntVar(&o.Bx, "X", o.Bx, "The X dimension of the text box.")
	fs.IntVar(&o.By, "Y", o.By, "The Y dimension of the text box.")
	fs.Float64Var(&o.Lx, "LX", o.Lx, "The X location point of the text box on the image.")
	fs.Float64Var(&o.Ly, "LY", o.Ly, "The Y location point of the text box on the image.")
	fs.Float64Var(&o.Opacity, "opacity", o.Opacity, "Opacity of the drawn text 1-100")
	//fs.BoolVar(&o.wrap, "wrap", o.wrap, "Wrap text to multiple lines where possible.")
	//padLeft, padTop, padRight, padBottom              float64
}

func tExecute(o *TextOptions, c context.Context, a []string) (context.Context, flip.ExitStatus) {
	var status flip.ExitStatus
	for _, fn := range executing {
		c, status = fn(o, c)
		if status != flip.ExitNo {
			return c, status
		}
	}
	return c, status
}

type execution func(o *TextOptions, c context.Context) (context.Context, flip.ExitStatus)

var tl log.Logger

func setLog(o *TextOptions, c context.Context) (context.Context, flip.ExitStatus) {
	tl = ctx.Log(c)
	return c, flip.ExitNo
}

func setFonts(o *TextOptions, c context.Context) (context.Context, flip.ExitStatus) {
	err := LF.SetDir(o.Directories()...)
	if err != nil {
		util.Failure(tl, "text", err)
		return c, flip.ExitFailure
	}
	return c, flip.ExitNo
}

func writeTextCtx(o *TextOptions, c context.Context) (context.Context, flip.ExitStatus) {
	cv := ctx.Canvas(c)
	WriteText(cv, o)
	c = context.WithValue(c, 4, cv)
	util.Success(tl, "text", fmt.Sprintf("wrote message '%s'", o.Msg))
	return c, flip.ExitNo
}

func textBoxImage(cv canvas.Canvas, o *TextOptions) canvas.Image {
	if o.Bx == 0 {
		o.Bx = cv.Bounds().Dx()
	}
	if o.By == 0 {
		o.By = cv.Bounds().Dy()
	}
	return canvas.NewFrom(cv.ColorModel(), o.Bx, o.By)
}

func WriteText(cv canvas.Canvas, o *TextOptions) {
	tb := textBoxImage(cv, o)
	t := NewText(o.PadLeft, (o.PadTop - o.PadBottom), o.Font, o.FontSize, o.ColorV(), o.Msg)
	t.Draw(tb, o)
	cv.Overlay(tb, canvas.Pt(o.Lx, o.Ly), o.Opacity)
}

func debugTxt(o *TextOptions, c context.Context) (context.Context, flip.ExitStatus) {
	if d := ctx.DebugMap(c); d != nil {
		mm := structs.Map(o)
		for k, v := range mm {
			val := fmt.Sprintf("%v", v)
			d[k] = val
		}
		c = context.WithValue(c, 1, d)
	}
	return c, flip.ExitNo
}

func end(o *TextOptions, c context.Context) (context.Context, flip.ExitStatus) {
	return c, flip.ExitSuccess
}

var executing = []execution{
	setLog,
	setFonts,
	writeTextCtx,
	debugTxt,
	end,
}

func textCommand() flip.Command {
	to := defaultTextOptions
	fs := flip.NewFlagSet("text", flip.ContinueOnError)
	tFlags(fs, to)
	return flip.NewCommand(
		"",
		"text",
		"Apply text to an image",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return tExecute(to, c, a)
		},
		fs,
	)
}

type TextLocation struct {
	X, Y float64
}

func (t *TextLocation) Locate() fixed.Point26_6 {
	return util.Fixp(t.X, t.Y)
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
		Size:    z,
		Hinting: font.HintingFull,
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

func (f *Fonts) List() []string {
	var ret util.AlphaS
	for k, _ := range f.has {
		ret = append(ret, k)
	}
	sort.Sort(ret)
	return ret
}

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

type Text struct {
	*TextLocation
	*TextFont
	raw string
}

func NewText(x, y float64, f string, z float64, c color.Color, txt string) *Text {
	return &Text{
		TextLocation: NewTextLocation(x, y),
		TextFont:     LF.TextFont(f, z, c),
		raw:          txt,
	}
}

func (t *Text) Draw(i draw.Image, o *TextOptions) {
	//switch {
	//case o.wrap:
	w := float64(i.Bounds().Dx()) - o.PadRight
	t.DrawStringWrapped(i, 0, 0, w, o.LineHeight, StringToAlign(o.Align))
	//default:
	//t.DrawString(i)
	//t.DrawStringAnchored(i, 0.5, 0.5)
	//}
}

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

func (t *Text) DrawString(i draw.Image) {
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

func (t *Text) DrawStringWrapped(i draw.Image, ax, ay, width, lineSpacing float64, a Align) {
	drawStringWrapped(i, t.TextFont, t.raw, t.X, t.Y, ax, ay, width, lineSpacing, a)
}

type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

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

var LF *Fonts

func initLocalFonts() {
	LF = &Fonts{make(map[string]*truetype.Font)}
	ft, _ := Asset("ft/DroidSansMono.ttf")
	LF.SetByte("default", ft)
}

func init() {
	defaultTextOptions = defaultTOptions()
	initLocalFonts()
}
