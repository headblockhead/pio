package fifo

import (
	"errors"
	"fmt"
)

type Observer interface {
	Size() uint
	Level() uint
	IsEmpty() bool
	IsFull() bool
	Buffer() []uint32
}

type Reader interface {
	Observer() Observer
	Read() (uint32, error)
	Level() uint
	IsEmpty() bool
	Resize(size uint)
}

type Writer interface {
	Observer() Observer
	Write(uint32) error
	Level() uint
	IsFull() bool
	Resize(size uint)
}

type Undoer interface {
}

type FIFO struct {
	logs []*Log

	buf []uint32

	head  uint
	tail  uint
	level uint
}

func New(size uint) *FIFO {
	f := &FIFO{
		logs: []*Log{LogInit(size)},

		buf: make([]uint32, size),

		head:  0,
		tail:  0,
		level: 0,
	}
	return f
}

func (f *FIFO) Resize(size uint) {
	if size == f.Size() {
		return
	}
	f.buf = make([]uint32, size)
	f.head = 0
	f.tail = 0
	f.level = 0
	f.logs = append(f.logs, LogResize(size))
}

func (f *FIFO) Observer() Observer { return f }
func (f *FIFO) Reader() Reader     { return f }
func (f *FIFO) Writer() Writer     { return f }
func (f *FIFO) Size() uint         { return (uint)(len(f.buf)) }
func (f *FIFO) Level() uint        { return f.level }
func (f *FIFO) Buffer() []uint32   { return f.buf }
func (f *FIFO) IsEmpty() bool      { return f.Level() == 0 }

var ErrEmpty = errors.New("empty")

func (f *FIFO) Read() (uint32, error) {
	if f.IsEmpty() {
		return 0, fmt.Errorf("reading: %w", ErrEmpty)
	}
	f.level--
	value := f.buf[f.tail]
	f.tail = (f.tail + 1) % f.Size()
	f.logs = append(f.logs, LogRead())
	return value, nil
}

func (f *FIFO) IsFull() bool { return f.Level() >= f.Size() }

var ErrFull = errors.New("full")

func (f *FIFO) Write(value uint32) error {
	if f.IsFull() {
		return fmt.Errorf("writing 0x%08X: %w", value, ErrFull)
	}
	f.level++
	f.buf[f.head] = value
	f.head = (f.head + 1) % f.Size()
	f.logs = append(f.logs, LogWrite(value))
	return nil
}

func (f *FIFO) SaveLoader() SaveLoader {
	return f
}

func (f *FIFO) Save() []*Log {
	return f.logs
}

func (f *FIFO) Load(logs []*Log) {
	lastLog := logs[len(logs)-1]
	f.logs = logs
	f.buf = make([]uint32, len(lastLog.Buffer))
	copy(f.buf, lastLog.Buffer)
}
