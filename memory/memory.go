package memory

import (
	"errors"
	"fmt"
)

type Observer interface {
	Size() uint
	Data() []uint16
	Initialized() []bool
}

type Reader interface {
	Read(address uint) (uint16, error)
}

type Writer interface {
	Write(address uint, value uint16) error
}

type Memory struct {
	data        []uint16
	initialized []bool
}

func New(size uint) *Memory {
	return &Memory{
		data:        make([]uint16, size),
		initialized: make([]bool, size),
	}
}

func (m *Memory) Observer() Observer {
	return m
}

func (m *Memory) Size() uint          { return (uint)(len(m.data)) }
func (m *Memory) Data() []uint16      { return m.data }
func (m *Memory) Initialized() []bool { return m.initialized }

func (m *Memory) Reader() Reader {
	return m
}

var ErrOutOfBounds = errors.New("out of bounds")
var ErrValueUninitialized = errors.New("value uninitialized")

func (m *Memory) Read(address uint) (uint16, error) {
	if address >= m.Size() {
		return 0, fmt.Errorf("address %d: %w", address, ErrOutOfBounds)
	}
	if !m.initialized[address] {
		return 0, fmt.Errorf("address %d: %w", address, ErrValueUninitialized)
	}
	return m.data[address], nil
}

func (m *Memory) Writer() Writer {
	return m
}

func (m *Memory) Write(address uint, value uint16) error {
	if address >= m.Size() {
		return fmt.Errorf("address %d: %w", address, ErrOutOfBounds)
	}
	m.data[address] = value
	m.initialized[address] = true
	return nil
}
