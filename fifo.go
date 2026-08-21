package pio

import "errors"

type fifo struct {
	buf []uint32

	head  uint
	tail  uint
	level uint
}

func newFIFO(size uint) *fifo {
	buf := make([]uint32, size)
	return &fifo{buf: buf}
}

var ErrFIFOFull = errors.New("FIFO is full")

func (f *fifo) write(value uint32) error {
	if f.isFull() {
		return ErrFIFOFull
	}
	f.level++
	f.buf[f.head] = value
	f.head = (f.head + 1) % 4
	return nil
}

var ErrFIFOEmpty = errors.New("FIFO is empty")

func (f *fifo) read() (value uint32, err error) {
	if f.isEmpty() {
		return 0, ErrFIFOEmpty
	}
	f.level--
	f.tail = (f.tail + 1) % 4
	return f.buf[f.tail], nil
}

func (f *fifo) size() uint {
	return (uint)(len(f.buf))
}

func (f *fifo) isFull() bool {
	return f.level >= f.size()
}

func (f *fifo) isEmpty() bool {
	return f.level == 0
}
