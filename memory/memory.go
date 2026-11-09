package memory

import (
	"errors"
	"fmt"
	"slices"
)

type Observer interface {
	Size() uint8
	Data() []uint16
	Initialized() []bool
}

type Reader interface {
	Read(address uint8) (uint16, error)
}

type Writer interface {
	Write(address uint8, value uint16) error
	Copy(source []uint16, destinationStartAddress uint8) error
	Clear()

	AddProgram(*Program) (uint8, error)
	AddProgramAtOffset(*Program, uint8) error
	RemoveProgram(*Program, uint8) error
}

type Memory struct {
	data        []uint16
	initialized []bool

	programUsageMask uint32
}

func New() *Memory {
	return &Memory{
		data:        make([]uint16, 32),
		initialized: make([]bool, 32),

		programUsageMask: 0,
	}
}

func (m *Memory) Observer() Observer {
	return m
}

func (m *Memory) Size() uint8         { return (uint8)(len(m.data)) }
func (m *Memory) Data() []uint16      { return m.data }
func (m *Memory) Initialized() []bool { return m.initialized }

func (m *Memory) Reader() Reader {
	return m
}

var ErrOutOfBounds = errors.New("out of bounds")
var ErrValueUninitialized = errors.New("value uninitialized")

func (m *Memory) Read(address uint8) (uint16, error) {
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

func (m *Memory) Write(address uint8, value uint16) error {
	if address >= m.Size() {
		return fmt.Errorf("address %d: %w", address, ErrOutOfBounds)
	}
	m.data[address] = value
	m.initialized[address] = true
	return nil
}

var ErrExceedsMemorySize = errors.New("exceeds memory size")

func (m *Memory) Copy(source []uint16, destinationStartAddress uint8) error {
	if destinationStartAddress+uint8(len(source)) > m.Size() {
		return fmt.Errorf("destination start address %d with length %d: %w", destinationStartAddress, len(source), ErrExceedsMemorySize)
	}
	for i, value := range source {
		m.data[destinationStartAddress+uint8(i)] = value
		m.initialized[destinationStartAddress+uint8(i)] = true
	}
	return nil
}

func (m *Memory) Clear() {
	for i := range m.data {
		m.data[i] = uint16(i & 0b11111) // jump to self
		m.initialized[i] = false
	}
	m.programUsageMask = 0
}

var ErrProgramEmpty = errors.New("program is empty")
var ErrProgramCannotFit = errors.New("program cannot fit")

func (m *Memory) AddProgram(program *Program) (offset uint8, err error) {
	if len(program.Instructions()) < 1 {
		return 0, ErrProgramEmpty
	}
	if len(program.Instructions()) > int(m.Size()) {
		return 0, fmt.Errorf("program length %d: %w", len(program.Instructions()), ErrExceedsMemorySize)
	}
	if program.HasOrigin() {
		offset := program.Origin()
		return offset, m.AddProgramAtOffset(program, offset)
	}
	for offset := 32 - len(program.Instructions()); offset >= 0; offset-- {
		err := m.AddProgramAtOffset(program, uint8(offset))
		if !errors.Is(err, ErrProgramOverlapsExistingPrograms) {
			return uint8(offset), err
		}
	}
	return 0, ErrProgramCannotFit
}

var ErrProgramOverlapsExistingPrograms = errors.New("program overlaps existing programs")

func (m *Memory) AddProgramAtOffset(program *Program, offset uint8) error {
	if len(program.Instructions()) < 1 {
		return ErrProgramEmpty
	}
	if offset+uint8(len(program.Instructions())) > m.Size() {
		return fmt.Errorf("offset %d with length %d: %w", offset, len(program.Instructions()), ErrExceedsMemorySize)
	}
	var programMask uint32 = ((0b1 << len(program.Instructions())) - 1) << offset
	if programMask&m.programUsageMask != 0 {
		return ErrProgramOverlapsExistingPrograms
	}
	updatedInstructions := slices.Clone(program.Instructions())
	for i, instr := range updatedInstructions {
		instructionType := (instr >> 13) & 0b111
		// update jump instructions to point to the correct offset
		if instructionType == 0b000 {
			var originalAddress = instr & 0b11111
			var offsetAddress = originalAddress + uint16(offset)
			var newInstruction = (instr & 0b1111111111100000) | (offsetAddress & 0b11111)
			updatedInstructions[i] = newInstruction
		}
	}
	err := m.Copy(updatedInstructions, offset)
	if err != nil {
		return fmt.Errorf("copying program: %w", err)
	}
	m.programUsageMask |= programMask
	return nil
}

func (m *Memory) RemoveProgram(program *Program, loadedOffset uint8) error {
	if len(program.instructions) < 1 {
		return ErrProgramEmpty
	}
	var programMask uint32 = ((0b1 << len(program.instructions)) - 1) << loadedOffset
	for i := 0; i < len(program.instructions); i++ {
		m.initialized[loadedOffset+uint8(i)] = false
	}
	m.programUsageMask &= ^programMask
	return nil
}
