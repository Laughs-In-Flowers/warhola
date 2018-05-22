package canvas

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"golang.org/x/image/math/fixed"
)

func init() {
	rr = rand.New(rand.NewSource(time.Now().UnixNano()))
}

var (
	rr           *rand.Rand
	tSizeMax     int = 400
	tPointNum    int = int(float64(tSizeMax) * 0.1)
	tPointErrExp int = int(float64(tPointNum) * 0.1)
)

func TestCanvas(t *testing.T) {
	setup(t)
	for _, m := range AvailableColorModels {
		for _, k := range AvailableKind {
			model := m.String()
			kind := k.String()
			name := fmt.Sprintf("cnv-%s.%s", model, kind)
			path := filepath.Join(testDir, name)
			dimX, dimY := rr.Intn(tSizeMax)+1, rr.Intn(tSizeMax)+1
			tc := newTestCanvas(model,
				path,
				kind,
				randPPU(),
				randPP(),
				dimX,
				dimY,
				t)
			probeCanvas(t, tc)
		}
	}
	teardown(t)
}

var testDir string = "/tmp/test-warhola-canvas"

func setup(t *testing.T) {
	err := os.MkdirAll(testDir, 0777)
	if err != nil {
		t.Errorf("test setup error: %s", err.Error())
	}
}

func probeCanvas(t *testing.T, c *testCanvas) {
	probeIdentity(t, c)
	probePxl(t, c)
	probeEqual(t, c)
	probeTransformer(t, c)
}

func probeIdentity(t *testing.T, c *testCanvas) {
	cnv, env, lnv := c.original.c, c.existing.c, c.clone.c
	tpt := c.path

	// original path
	cpt := cnv.Path()
	if cpt != tpt {
		failProbe(
			t,
			c.id,
			"Pather.Path",
			commonExpect,
			tpt, cpt,
		)
	}

	// new path
	cnv.SetPath("/new")
	npt := cnv.Path()
	if npt == tpt {
		failProbe(
			t,
			c.id,
			"new Pather.Path",
			commonExpect,
			"/new", npt,
		)
	}
	cnv.SetPath(c.path)

	// action
	ca := cnv.Action()
	ea := env.Action()
	la := lnv.Action()
	switch {
	case ca != "New", ea != "Open", la != "Clone":
		failProbe(
			t,
			c.id,
			"Identity.Action",
			commonExpect,
			"New, Open, Clone", fmt.Sprintf("%s, %s, %s", ca, ea, la),
		)
	}

	// kind
	ck := cnv.Kind()
	if ck != c.kind {
		failProbe(
			t,
			c.id,
			commonExpect,
			c.kind, ck,
		)
	}
}

func probePxl(t *testing.T, c *testCanvas) {
	probeImage(t, c)
	probePix(t, c)
	probeMeasure(t, c)
	probePaster(t, c)
}

func probeImage(t *testing.T, c *testCanvas) {
	i := c.original.c

	// ColorModel
	icm := i.ColorModel()
	if icm != c.color {
		failProbe(
			t,
			c.id,
			"Canvas.ColorModel",
			"expected %v, got %v",
			c.color, icm,
		)
	}

	// Bounds
	ib := i.Bounds()
	if ib != c.rect {
		failProbe(
			t,
			c.id,
			"Canvas.Bounds",
			"expected %v, got %v",
			c.rect, ib,
		)
	}

	// At
	probeAt(
		t,
		c.id,
		i,
		c.ps,
		defaultColorCompare,
	)

	// Opaque
	op := i.Opaque()
	v := reflect.ValueOf(op)
	if v.Kind() != reflect.Bool {
		failProbe(
			t,
			c.id,
			"Canvas.Opaque",
			"not a boolean type",
		)
	}

	sp := randPoint(c.color, c.rect.Max.X, c.rect.Max.Y, c.ps)
	i.Set(sp.x, sp.y, sp.c)
	gc := i.At(sp.x, sp.y)
	if gc != sp.c {
		failProbe(
			t,
			c.id,
			"Canvas.Set",
			"expected %v, got %v",
			sp.c, gc,
		)
	}

	// SubImage
	is := i.SubImage(c.sub)
	isb := is.Bounds()
	if isb != c.sub {
		failProbe(
			t,
			c.id,
			"Canvas.SubImage",
			"expected %v, got %v",
			isb, c.sub,
		)
	}
}

type equalColor func(color.Color, color.Color) bool

func defaultColorCompare(a color.Color, b color.Color) bool {
	return a == b
}

func probeAt(
	t *testing.T,
	id string,
	c Canvas,
	ps []*testPoint,
	ecfn equalColor,
) {
	var errs []string
	for _, p := range ps {
		getColor := c.At(p.x, p.y)
		if !ecfn(getColor, p.c) {
			errs = append(errs, fmt.Sprintf("\ttest color retrieved %v not equal to expected color %v -- (%d,%d)",
				getColor,
				p.c,
				p.x, p.y,
			))
		}
		if len(errs) >= tPointErrExp {
			msg := fmt.Sprintf("too many errors (%d):\n", len(errs))
			var msgs []string
			for _, e := range errs {
				msgs = append(msgs, fmt.Sprintf("%s\n", e))
			}
			failProbeMulti(t, id, "Canvas.At", msg, msgs...)
		}
	}
}

func probePix(t *testing.T, c *testCanvas) {
	o := c.original.c

	// get/set one pix item
	op := o.Pix(1)
	np := op + 5
	o.SetPix(1, np)
	tp := o.Pix(1)
	if tp != np {
		failProbe(
			t,
			c.id,
			"Pix.Get/Pix.Set",
			"unable to properly set and get pix",
		)
	}
	o.SetPix(1, op)

	// stride
	str := o.Stride()
	v := reflect.ValueOf(str)
	if v.Kind() != reflect.Int {
		failProbe(
			t,
			c.id,
			"Pix.Stride",
			"not an integer",
		)
	}
}

func probeMeasure(t *testing.T, c *testCanvas) {
	o := c.original.c
	b := o.Bounds()

	// PP
	inch, cm, mm := o.PP("inch"), o.PP("cm"), o.PP("mm")
	mi := math.Round(inch / 2.54)
	mc := math.Round(cm)
	mmi := math.Round(mm * 10)
	switch {
	case mi != mc,
		mc != mmi,
		mmi != mi:
		failProbe(
			t,
			c.id,
			"Measure.PP",
			"inch, cm, and mm are not equal(%F, %F, %F)",
			mi, mc, mmi,
		)
	}

	// Height, Width
	w, h := float64(b.Dx()), float64(b.Dy())
	ow, oh := o.Width(), o.Height()
	if w != ow || h != oh {
		failProbe(
			t,
			c.id,
			"Measure.Height/Measure.Width",
			"not equal [%s],[%s]",
			fmt.Sprintf("%F != %F", w, ow), fmt.Sprintf("%F != %F", h, oh),
		)
	}

	// Radius
	radius := float64(b.Dy()) / 2
	oradius := o.Radius()
	if oradius != radius {
		failProbe(
			t,
			c.id,
			"Measure.Radius",
			commonExpect,
			radius, oradius,
		)
	}

	// Anchors
	A, B, C := o.Anchor("A"), o.Anchor("B"), o.Anchor("C")
	origin, center, bound := o.Anchor("origin"), o.Anchor("center"), o.Anchor("bound")
	ecenter := Moint{float64(c.rect.Max.X) / 2, float64(c.rect.Max.Y) / 2}
	ebound := Moint{float64(c.rect.Max.X), float64(c.rect.Max.Y)}
	switch {
	case A == ZM, B == ZM, C != ZM:
		failProbe(
			t,
			c.id,
			"Measure.Anchor",
			"retrieved anchor moint are zero or exist when they should not",
		)
	case origin != ZM,
		center != ecenter,
		bound != ebound:
		failProbe(
			t,
			c.id,
			"Measure.Anchor",
			"default anchors incorrect expected: (%v, %v, %v) got: (%v, %v,%v)",
			origin, center, bound, ZM, ecenter, ebound,
		)
	}

	// Distance
	d := o.Distance("", A, B)
	ed := o.PP("")
	if math.Round(d) != ed {
		failProbe(
			t,
			c.id,
			"Distance",
			commonExpect,
			ed, d,
		)
	}

	du := math.Round(o.Distance(c.ppu, A, B))
	if du != 1 {
		failProbe(
			t,
			c.id,
			"Distance of Unit",
			commonExpect,
			1, du,
		)
	}
}

func probeEqual(t *testing.T, c *testCanvas) {
	action := func(s string) string {
		return fmt.Sprintf("equal Canvas -- %s", s)
	}
	o, e, l := c.original.c, c.existing.c, c.clone.c

	// ColorModel
	om := o.ColorModel()
	em := e.ColorModel()
	cm := l.ColorModel()
	switch {
	case om != c.color, em != c.color, cm != c.color:
		failProbe(
			t,
			c.id,
			action("ColorModel"),
			"expected %s, got (%v,%v,%v)",
			c.model, om, em, cm,
		)
	}

	// Bounds
	ob := o.Bounds()
	eb := e.Bounds()
	cb := l.Bounds()
	switch {
	case ob != c.rect, eb != c.rect, cb != c.rect:
		failProbe(
			t,
			c.id,
			action("Bounds"),
			"expected %v, got (%v,%v,%v)",
			c.rect, ob, eb, cb,
		)
	}

	// At
	// no adequate strategy for comparing a new image/pxl to one read out of
	// the file system for now; its not a (major) problem with png, but jpg
	// images create gross errors between zeroed and read in blank pix entries
	var aerrs []string
	var pcount int = 0
	var laerrs int = 0
	for x := 0; x <= c.rect.Max.X; x = x + 1 {
		for y := 0; y <= c.rect.Max.Y; y = y + 1 {
			pcount = pcount + 1
			oa := o.At(x, y)
			ea := e.At(x, y)
			ca := l.At(x, y)
			errsAppend := func() {
				aerrs = append(aerrs,
					fmt.Sprintf("\tat (%d, %d) unequal pixels: %v, %v, %v\n", x, y, oa, ea, ca),
				)
				laerrs = laerrs + 1
			}
			switch {
			case ea != ca:
				errsAppend()
			}
		}
	}
	acceptable := int(float64(pcount) * 0.05)
	if laerrs > acceptable {
		msg := "compare pixel errors"
		var msgs []string
		for _, m := range aerrs {
			msgs = append(msgs, m)
		}
		failProbeMulti(t, c.id, action("At"), msg, msgs...)
	}

	// Opaque
	// again new image not run through Kind encoding(eg jpg) behaves differently
	// so one to one comparison is a problem
	eo := e.Opaque()
	co := l.Opaque()
	if eo != co {
		failProbe(
			t,
			c.id,
			action("Opaque"),
			"unequal results",
		)
	}

	// PixOffset
	ox, oy := rr.Intn(c.rect.Max.X), rr.Intn(c.rect.Max.Y)
	oof := o.PixOffset(ox, oy)
	eof := e.PixOffset(ox, oy)
	cof := l.PixOffset(ox, oy)
	switch {
	case oof != eof, eof != cof, cof != oof:
		failProbe(
			t,
			c.id,
			action("PixOffset"),
			"expected equal (%d,%d,%d)",
			oof, eof, cof,
		)
	}

	// Set
	sp := randPoint(c.color, c.dX, c.dY, c.ps)
	o.Set(sp.x, sp.y, sp.c)
	e.Set(sp.x, sp.y, sp.c)
	l.Set(sp.x, sp.y, sp.c)

	oc, ec, cc := o.At(sp.x, sp.y), e.At(sp.x, sp.y), l.At(sp.x, sp.y)
	switch {
	case oc != ec, ec != cc, cc != oc:
		failProbe(
			t,
			c.id,
			action("Set"),
			"expected equal results got (%v,%v,%v)",
			oc, ec, cc,
		)
	}

	// SubImage
	sb := c.sub
	osb := o.SubImage(sb).Bounds()
	esb := e.SubImage(sb).Bounds()
	csb := l.SubImage(sb).Bounds()
	switch {
	case osb != sb, esb != sb, csb != sb:
		failProbe(
			t,
			c.id,
			action("SubImage"),
			"expected %v, got (%v,%v,%v)",
			sb, osb, esb, csb,
		)
	}
}

func probePaster(t *testing.T, c *testCanvas) {
	// paste
	probeAt(t, c.id, c.paste.c, c.pastePs, defaultColorCompare)

	// overlay
	// calculating expected opacity color and compare to have would be better
	// here, but this suffices for now
	probeAt(t, c.id, c.overlay.c, c.overPs, defaultColorCompare)
}

func probeTransformer(t *testing.T, c *testCanvas) {
	probeCropper(t, c)
	probeResizer(t, c)
	probeTranslater(t, c)
}

func probeCropper(t *testing.T, c *testCanvas) {
	cropped := c.crop.c.Bounds()
	if c.sub.Size() != cropped.Size() {
		failProbe(
			t,
			c.id,
			"Crop",
			commonExpect,
			c.sub, cropped,
		)
	}
}

func probeResizer(t *testing.T, c *testCanvas) {}

func probeTranslater(t *testing.T, c *testCanvas) {}

var commonExpect string = "expected %v, got %v"

func failProbe(t *testing.T, id, action, msg string, with ...interface{}) {
	t.Logf("%s %s %s", id, action, fmt.Sprintf(msg, with...))
	teardown(t)
	t.FailNow()
}

func failProbeMulti(t *testing.T, id, action, msg string, msgs ...string) {
	t.Logf("%s %s %s", id, action, msg)
	for _, m := range msgs {
		t.Log(m)
	}
	teardown(t)
	t.FailNow()
}

func teardown(t *testing.T) {
	err := os.RemoveAll(testDir)
	if err != nil {
		t.Errorf("test teardown error: %s", err.Error())
	}
}

type testCanvas struct {
	id                        string
	model                     string
	color                     color.Model
	path, kind, ppu           string
	pp, opacity               float64
	dX, dY                    int
	rect, sub                 image.Rectangle
	ps, pastePs, overPs       []*testPoint
	original, existing, clone *tcInstance
	paste, overlay, crop      *tcInstance
}

type tcInstance struct {
	label string
	c     Canvas
}

type testPoint struct {
	c    color.Color
	x, y int
}

func randPointSet(r, inX, inY int, l color.Model) []*testPoint {
	ret := make([]*testPoint, 0)
	for i := 0; i <= r; i = i + 1 {
		ret = append(ret, randPoint(l, inX, inY, ret))
	}
	return ret
}

func randPoint(l color.Model, inX, inY int, ps []*testPoint) *testPoint {
	var x, y int
	var exist bool
	for {
		x, y = rr.Intn(inX), rr.Intn(inY)
		exist = rpExists(x, y, ps)
		if !exist {
			break
		}
	}
	c := randColor(l)
	return &testPoint{c, x, y}
}

func rpExists(x, y int, ps []*testPoint) bool {
	for _, p := range ps {
		if x == p.x && y == p.y {
			return true
		}
	}
	return false
}

func randColor(l color.Model) color.Color {
	initial := color.RGBA{
		uint8(rr.Intn(255)),
		uint8(rr.Intn(255)),
		uint8(rr.Intn(255)),
		uint8(rr.Intn(255)),
	}
	switch l {
	case color.RGBAModel:
		return initial
	default:
		return l.Convert(initial)
	}
	return nil
}

func randPPU() string {
	l := len(AvailablePPU)
	i := rr.Intn(l)
	return AvailablePPU[i]
}

func randPP() float64 {
	i := rr.Intn(400) + 1
	return float64(i + 50)
}

func randIntersect(of image.Rectangle) image.Rectangle {
	var ret image.Rectangle
	end := 0
	for end == 0 {
		ret = image.Rect(rr.Intn(of.Max.X), rr.Intn(of.Max.Y), rr.Intn(of.Max.X), rr.Intn(of.Max.Y))
		inter := of.Intersect(ret)
		if inter != image.ZR {
			end = 1
		}
	}
	return ret
}

func randMointsOneUnitApart(r image.Rectangle, d float64) (MointFunc, MointFunc) {
	bd := r.Bounds()
	a := Moint{
		float64(rr.Intn(bd.Dx() / 2)),
		float64(rr.Intn(bd.Dy() / 2)),
	}
	af := func(Measure) Moint {
		return a
	}
	g := (math.Pi * 2) / 8
	bx := a.X + (d * math.Cos(g))
	by := a.Y + (d * math.Sin(g))
	b := Moint{bx, by}
	bf := func(Measure) Moint {
		return b
	}
	return af, bf
}

func newTestCanvas(model, path, kind, ppu string, pp float64, dX, dY int, t *testing.T) *testCanvas {
	o, nErr := New(
		SetDebug(true),
		SetColorModel(model),
		SetPath(path),
		SetKind(kind),
		SetMeasure(pp, ppu),
		SetRect(dX, dY),
	)
	if nErr != nil {
		t.Errorf("new %s canvas error: %s", model, nErr.Error())
	}

	//blank for pasting later
	paster := o.Clone()
	overlayer := o.Clone()

	ps := randPointSet(tPointNum, dX, dY, o.ColorModel())
	for _, p := range ps {
		o.Set(p.x, p.y, p.c)
	}

	sErr := o.Save()
	if sErr != nil {
		t.Errorf("new canvas %s save error: %s", model, sErr.Error())
	}

	orect := image.Rect(0, 0, dX, dY)
	subrect := randIntersect(orect)

	a, b := randMointsOneUnitApart(orect, pp)
	o.SetAnchor("A", a)
	o.SetAnchor("B", b)

	original := &tcInstance{"original", o}

	tc := &testCanvas{
		path,
		model,
		o.ColorModel(),
		path,
		kind,
		ppu,
		pp,
		rr.Float64() * 100,
		dX,
		dY,
		orect,
		subrect,
		ps, nil, nil,
		original,
		nil,
		nil,
		nil,
		nil,
		nil,
	}

	x, xErr := New(
		SetDebug(true),
		SetColorModel(tc.model),
		SetPath(tc.path),
		SetKind(tc.kind),
		SetMeasure(tc.pp, tc.ppu),
		SetRect(tc.dX, tc.dY),
	)
	if xErr != nil {
		t.Errorf("existing canvas %s create error: %s", model, xErr.Error())
	}
	tc.existing = &tcInstance{"existing", x}

	cl := x.Clone()
	tc.clone = &tcInstance{"clone", cl}

	// paster paste
	pasteSet := randPointSet(tPointNum, dX, dY, o.ColorModel())
	for _, p := range pasteSet {
		paster.Set(p.x, p.y, p.c)
	}
	tc.pastePs = pasteSet
	toPaste := o.Clone()
	toPaste.Paste(paster, image.Point{0, 0})
	tc.paste = &tcInstance{"paste", toPaste}

	// paster overlay
	toOverlay := o.Clone()
	for _, p := range pasteSet {
		overlayer.Set(p.x, p.y, p.c)
	}
	toOverlay.Overlay(overlayer, image.Point{0, 0}, tc.opacity)
	var ovrSet []*testPoint
	for _, v := range pasteSet {
		olc := toOverlay.At(v.x, v.y)
		ovrSet = append(ovrSet, &testPoint{olc, v.x, v.y})
	}
	tc.overlay = &tcInstance{"overlay", toOverlay}
	tc.overPs = ovrSet

	// cropper
	cc, _ := o.CropTo(subrect)
	tc.crop = &tcInstance{"crop", cc}

	//create paletted and uniform to open

	return tc
}

func TestMoint(t *testing.T) {
	id := "Moint"
	a := Moint{0, 0}

	// Fixed
	f := fixed.Point26_6{0, 0}
	if a.Fixed() != f {
		failProbe(
			t,
			id,
			"Fixed",
			commonExpect,
			a, f,
		)
	}

	b := Moint{50, 100}

	// Distance
	expD := math.Hypot(a.X-b.X, a.Y-b.Y)
	if dis := a.Distance(b); dis != expD {
		failProbe(
			t,
			id,
			"Distance",
			commonExpect,
			dis, expD,
		)
	}

	// Interpolate
	intp := float64(rr.Intn(100))
	interpExp := Moint{
		a.X + (b.X-a.X)*intp,
		a.Y + (b.Y-a.Y)*intp,
	}
	if interpHave := a.Interpolate(b, intp); interpHave != interpExp {
		failProbe(
			t,
			id,
			"Interpolate",
			commonExpect,
			interpHave, interpExp,
		)
	}

	iz := image.Point{0, 0}

	// IPoint() image.Point {
	pointExp := iz
	if pointHave := a.IPoint(); pointHave != pointExp {
		failProbe(
			t,
			id,
			"IPoint",
			commonExpect,
			pointHave, pointExp,
		)
	}

	// ToMoint(ip image.Point) Moint {
	if toMointHave := ToMoint(iz); toMointHave != a {
		failProbe(
			t,
			"-",
			"ToMoint",
			commonExpect,
			toMointHave, a,
		)
	}

	// Add
	am := Moint{float64(rr.Intn(500)), float64(rr.Intn(500))}
	addExp := Moint{a.X + am.X, a.Y + am.Y}
	if addHave := a.Add(am); addHave != addExp {
		failProbe(
			t,
			id,
			"Add",
			commonExpect,
			addHave, addExp,
		)
	}

	// Div
	divV := float64(rr.Intn(10))
	divExp := Moint{b.X / divV, b.Y / divV}
	if divHave := b.Div(divV); divHave != divExp {
		failProbe(
			t,
			id,
			"Div",
			commonExpect,
			divHave, divExp,
		)
	}

	// Eq(o Moint) bool {
	eqM := Moint{0, 0}
	if !a.Eq(eqM) {
		failProbe(
			t,
			id,
			"Eq",
			commonExpect,
			eqM, a,
		)
	}

	irect := image.Rect(0, 0, 51, 101)

	// In
	rmX, rmY := float64(irect.Min.X), float64(irect.Min.Y)
	rxX, rxY := float64(irect.Max.X), float64(irect.Max.Y)
	isIn := rmX <= b.X && b.X < rxX &&
		rmY <= b.Y && b.Y < rxY
	if isIn != b.In(irect) {
		failProbe(
			t,
			id,
			"In",
			"%s is not in %s",
			b, irect,
		)
	}

	// Mod
	// warning label on this section; still a ?
	rMin := ToMoint(irect.Min)
	w, h := float64(irect.Dx()), float64(irect.Dy())
	p := b.Sub(rMin)
	p.X = math.Mod(p.X, w)
	if p.X < 0 {
		p.X += w
	}
	p.Y = math.Mod(p.Y, h)
	if p.Y < 0 {
		p.Y += h
	}
	modExp := p.Add(rMin)
	if modHave := b.Mod(irect); modHave != modExp {
		failProbe(
			t,
			id,
			"Mod",
			commonExpect,
			modHave, modExp,
		)
	}

	// Mul
	mulV := float64(rr.Intn(100) + 1)
	mulExp := Moint{b.X * mulV, b.Y * mulV}
	if mulHave := b.Mul(mulV); mulHave != mulExp {
		failProbe(
			t,
			id,
			"Mul",
			commonExpect,
			mulHave, mulExp,
		)
	}

	// String()
	stringExp := "(0.0000,0.0000)"
	if stringHave := a.String(); stringHave != stringExp {
		failProbe(
			t,
			id,
			"String",
			commonExpect,
			stringHave, stringExp,
		)
	}

	// Sub(o Moint) Moint {
	subMoint := Moint{float64(rr.Intn(10)), float64(rr.Intn(10))}
	subExp := Moint{b.X - subMoint.X, b.Y - subMoint.Y}
	if subHave := b.Sub(subMoint); subHave != subExp {
		failProbe(
			t,
			id,
			"Sub",
			commonExpect,
			subHave, subExp,
		)
	}
}
