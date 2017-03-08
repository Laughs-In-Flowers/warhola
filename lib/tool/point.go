package tool

type Point struct {
	X, Y float64
}

type Anchor struct {
	Tag string
	*Point
}

func NewAnchor(tag string, point *Point) *Anchor {
	return &Anchor{tag, point}
}

func (t *Tool) SetAnchor(tag string, point *Point) {
	a := NewAnchor(tag, point)
	t.Anchors = append(t.Anchors, a)
}

func (t *Tool) GetAnchor(tag string) *Point {
	for _, a := range t.Anchors {
		if a.Tag == tag {
			return a.Point
		}
	}
	return nil
}
