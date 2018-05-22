package canvas

import (
	"strings"

	"github.com/Laughs-In-Flowers/warhola/lib/util/xrr"
)

type Identity interface {
	Pather
	Action() string
	Kind() string
}

type identity struct {
	*pather
	action Action
	kind   Kind
}

func newIdentity() *identity {
	return &identity{
		&pather{""}, ACTIONNOOP, KINDNOOP,
	}
}

// An interface for encapsulating a string path.
type Pather interface {
	Path() string
	SetPath(string)
}

type pather struct {
	path string
}

// A default path.
const PATHNOOP = "PathNoop"

// Returns a string for this *pathers path.
func (p *pather) Path() string {
	return p.path
}

// Sets the provided string as this *pathers path.
func (p *pather) SetPath(as string) {
	p.path = as
}

func (p *pather) clone() *pather {
	np := *p
	return &np
}

// A type indicating the action identity of the canvas: noop, new, open, clone.
type Action int

func (i *identity) Action() string {
	return i.action.String()
}

func (a Action) String() string {
	switch a {
	case ACTIONNEW:
		return "New"
	case ACTIONOPEN:
		return "Open"
	case ACTIONCLONE:
		return "Clone"
	}
	return "Noop"
}

const (
	ACTIONNOOP Action = iota
	ACTIONNEW
	ACTIONOPEN
	ACTIONCLONE
)

// Indicates the type of file the canvas is.
type Kind int

func (i *identity) Kind() string {
	return i.kind.String()
}

const (
	KINDNOOP Kind = iota
	PNG
	JPG
)

//A variable containing a listing of available and fully functional Kind.
var AvailableKind = []Kind{
	PNG,
	JPG,
}

func stringToKind(s string) Kind {
	switch strings.ToUpper(s) {
	case "PNG":
		return PNG
	case "JPEG", "JPG":
		return JPG
	}
	return KINDNOOP
}

// Provides a string of this Kind.
func (k Kind) String() string {
	switch k {
	case PNG:
		return "png"
	case JPG:
		return "jpg"
	}
	return "KindNoop"
}

var kindError = xrr.Xrror("%s is not a recognized image kind").Out

func (i *identity) clone() *identity {
	return &identity{
		i.pather.clone(), ACTIONCLONE, i.kind,
	}
}
