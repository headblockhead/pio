package pio

import "errors"

type FIFOWriter interface {
	Write(value uint32) error
}

type FIFOReader interface {
	Read() (value uint32, err error)
}

type FIFO struct {
	buf [4]uint32

	head  uint
	tail  uint
	level uint
}

func NewFIFO() *FIFO {
	return &FIFO{
		buf: [4]uint32{},

		head:  0,
		tail:  0,
		level: 0,
	}
}

var FIFOFull = errors.New("FIFO is full")

func (f *FIFO) Write(value uint32) error {
	if f.level == 4 {
		return FIFOFull
	}
	f.level++
	f.buf[f.head] = value
	f.head = (f.head + 1) % 4
	return nil
}

var FIFOEmpty = errors.New("FIFO is empty")

func (f *FIFO) Read() (value uint32, err error) {
	if f.level == 0 {
		return 0, FIFOEmpty
	}
	f.level--
	f.tail = (f.tail + 1) % 4
	return f.buf[f.tail], nil
}

func (f *FIFO) Level() uint {
	return f.level
}
