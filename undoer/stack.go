package undoer

import (
	"errors"
	"slices"
)

type Undoer[L Log] struct {
	undoable  Undoable[L]
	undoStack []L
	redoStack []L
}

func New[L Log](undoable Undoable[L]) *Undoer[L] {
	return &Undoer[L]{
		undoable:  undoable,
		undoStack: []L{},
		redoStack: []L{},
	}
}

func (u *Undoer[L]) UndoStack() []L { return u.undoStack }
func (u *Undoer[L]) RedoStack() []L { return u.redoStack }

func (u *Undoer[L]) AddHistory(l L) { u.undoStack = append(u.undoStack, l) }

var ErrUndoOutOfRange = errors.New("undo out of range")

func (u *Undoer[L]) Undo(i uint) error {
	if i == 0 {
		return nil
	}
	if int(i) > len(u.undoStack) {
		return ErrUndoOutOfRange
	}

	newTip := len(u.undoStack) - int(i)
	undone := slices.Clone(u.undoStack[newTip:])
	u.undoStack = u.undoStack[:newTip]

	slices.Reverse(undone)
	u.redoStack = append(u.redoStack, undone...)

	u.undoable.Rebuild(u.undoStack)

	return nil
}

var ErrRedoOutOfRange = errors.New("redo out of range")

func (u *Undoer[L]) Redo(i uint) error {
	if i == 0 {
		return nil
	}
	if int(i) > len(u.redoStack) {
		return ErrRedoOutOfRange
	}

	newTip := len(u.redoStack) - int(i)
	redone := slices.Clone(u.redoStack[newTip:])
	u.redoStack = u.redoStack[:newTip]

	slices.Reverse(redone)
	u.undoStack = append(u.undoStack, redone...)

	u.undoable.Rebuild(u.undoStack)

	return nil
}
