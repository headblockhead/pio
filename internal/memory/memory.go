package memory

import "errors"

type Memory struct {
	data []uint16
}

func NewMemory(size uint) *Memory {
	return &Memory{}
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

func (m *Memory) Read(address uint) (uint16, error) {
	if address >= m.Size() {
		return 0, ErrMemoryOutOfBounds
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
	return nil
}
