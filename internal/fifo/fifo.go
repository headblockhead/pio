package fifo

import "errors"

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

func (f *FIFO) Size() uint {
	return (uint)(len(f.buf))
}

type FIFOReader interface {
	Read() (uint32, error)
	IsEmpty() bool
}

func (f *FIFO) Reader() FIFOReader {
	return f
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
func (f *FIFO) IsEmpty() bool {
	return f.level == 0
}

type FIFOWriter interface {
	Write(uint32) error
	IsFull() bool
}

func (f *FIFO) Writer() FIFOWriter {
	return f
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
func (f *FIFO) IsFull() bool {
	return f.level >= f.Size()
}
