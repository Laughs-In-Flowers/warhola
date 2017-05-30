package tool

// A struct encompassing a 2D float64 X - Y coordinate
type Point struct {
	X, Y float64
}

// A struct embedding a Point and a string tag as formal name.
type Anchor struct {
	Tag string
	*Point
}

// Produces a new Anchor from the provided tag nad point.
func NewAnchor(tag string, point *Point) *Anchor {
	return &Anchor{tag, point}
}

// Sets an Anchor for use in a Tool instance.
func (t *Tool) SetAnchor(tag string, point *Point) {
	a := NewAnchor(tag, point)
	t.Anchors = append(t.Anchors, a)
}

// Gets the provided Anchor by tag from the tool instance.
func (t *Tool) GetAnchor(tag string) *Point {
	for _, a := range t.Anchors {
		if a.Tag == tag {
			return a.Point
		}
	}
	return nil
}
