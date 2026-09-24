package memory

import "errors"

type Memory struct {
	data        []uint16
	initialized []bool
}

func NewMemory(size uint) *Memory {
	return &Memory{
		data:        make([]uint16, size),
		initialized: make([]bool, size),
	}
}

func (m *Memory) Size() uint {
	return (uint)(len(m.data))
}

type MemoryReader interface {
	Read(address uint) (uint16, error)
}

func (m *Memory) Reader() MemoryReader {
	return m
}

var ErrMemoryOutOfBounds = errors.New("out of bounds")
var ErrMemoryUninitialized = errors.New("uninitialized memory")

func (m *Memory) Read(address uint) (uint16, error) {
	if address >= m.Size() {
		return 0, ErrMemoryOutOfBounds
	}
	if !m.initialized[address] {
		return 0, ErrMemoryUninitialized
	}
	return m.data[address], nil
}

type MemoryWriter interface {
	Write(address uint, value uint16) error
}

func (m *Memory) Writer() MemoryWriter {
	return m
}

func (m *Memory) Write(address uint, value uint16) error {
	if address >= m.Size() {
		return ErrMemoryOutOfBounds
	}
	m.data[address] = value
	m.initialized[address] = true
	return nil
}
