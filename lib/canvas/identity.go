package canvas

type Identity interface {
	Pather
	Action() string
	FileType() string
}

type identity struct {
	*pather
	action   Action
	fileType FileType
}

func newIdentity() *identity {
	return &identity{
		&pather{""}, ACTIONNOOP, FILETYPENOOP,
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

func (i *identity) FileType() string {
	return i.fileType.String()
}

func (i *identity) clone() *identity {
	return &identity{
		i.pather.clone(), ACTIONCLONE, i.fileType,
	}
}
