package pio

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

var ErrFIFOFull = errors.New("FIFO is full")

func (f *FIFO) Write(value uint32) error {
	if f.IsFull() {
		return ErrFIFOFull
	}
	f.level++
	f.buf[f.head] = value
	f.head = (f.head + 1) % 4
	return nil
}

var ErrFIFOEmpty = errors.New("FIFO is empty")

func (f *FIFO) Read() (value uint32, err error) {
	if f.IsEmpty() {
		return 0, ErrFIFOEmpty
	}
	f.level--
	f.tail = (f.tail + 1) % 4
	return f.buf[f.tail], nil
}

func (f *FIFO) Size() uint {
	return (uint)(len(f.buf))
}

func (f *FIFO) IsFull() bool {
	return f.level == f.Size()
}

func (f *FIFO) IsEmpty() bool {
	return f.level == 0
}
