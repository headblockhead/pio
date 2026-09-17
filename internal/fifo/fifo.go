package fifo

import (
	"errors"
)

type Observer interface {
	Size() uint
	Level() uint
	IsEmpty() bool
	IsFull() bool
	Buffer() []uint32
}

type Reader interface {
	Read() (uint32, error)
	Level() uint
	IsEmpty() bool
}

type Writer interface {
	Write(uint32) error
	Level() uint
	IsFull() bool
}

type FIFO struct {
	buf []uint32

	head  uint
	tail  uint
	level uint
}

func NewFIFO(size uint) *FIFO {
	buf := make([]uint32, size)
	return &FIFO{buf: buf}
}

func (f *FIFO) Observer() Observer {
	return f
}

func (f *FIFO) Reader() Reader {
	return f
}

func (f *FIFO) Writer() Writer {
	return f
}

func (f *FIFO) Size() uint {
	return (uint)(len(f.buf))
}

func (f *FIFO) Level() uint {
	return f.level
}

func (f *FIFO) Buffer() []uint32 {
	return f.buf
}

func (f *FIFO) IsEmpty() bool {
	return f.Level() == 0
}

var ErrFIFOEmpty = errors.New("FIFO is empty")

func (f *FIFO) Read() (uint32, error) {
	if f.IsEmpty() {
		return 0, ErrFIFOEmpty
	}
	f.level--
	value := f.buf[f.tail]
	f.tail = (f.tail + 1) % f.Size()
	return value, nil
}

func (f *FIFO) IsFull() bool {
	return f.Level() >= f.Size()
}

var ErrFIFOFull = errors.New("FIFO is full")

func (f *FIFO) Write(value uint32) error {
	if f.IsFull() {
		return ErrFIFOFull
	}
	f.level++
	f.buf[f.head] = value
	f.head = (f.head + 1) % f.Size()
	return nil
}
