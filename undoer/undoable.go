package undoer

type Undoable[L Log] interface {
	Rebuild([]L) error
}
