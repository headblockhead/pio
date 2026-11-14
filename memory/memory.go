package memory

import (
	"errors"
	"fmt"
	"slices"

	"git.edwardh.dev/pio/instr"
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
	Copy(source []uint16, destinationStartAddress uint) error
	Clear()

	AddProgram(*Program) (uint, error)
	AddProgramAtOffset(*Program, uint) error
	RemoveProgram(*Program, uint) error
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

func (m *Memory) Size() uint          { return uint(len(m.data)) }
func (m *Memory) Data() []uint16      { return m.data }
func (m *Memory) Initialized() []bool { return m.initialized }

func (m *Memory) Reader() Reader {
	return m
}

var ErrOutOfBounds = errors.New("out of bounds")
var ErrValueUninitialized = errors.New("value uninitialized")

func (m *Memory) Read(address uint) (uint16, error) {
	if address >= m.Size() {
		return 0, fmt.Errorf("%w: address %d", ErrOutOfBounds, address)
	}
	if !m.initialized[address] {
		return 0, fmt.Errorf("%w: address %d", ErrValueUninitialized, address)
	}
	return m.data[address], nil
}

func (m *Memory) Writer() Writer {
	return m
}

func (m *Memory) Write(address uint, value uint16) error {
	if address >= m.Size() {
		return fmt.Errorf("%w: address %d", ErrOutOfBounds, address)
	}
	m.data[address] = value
	m.initialized[address] = true
	return nil
}

var ErrExceedsMemorySize = errors.New("exceeds memory size")

func (m *Memory) Copy(source []uint16, toAddress uint) error {
	if toAddress+uint(len(source)) > m.Size() {
		return fmt.Errorf("%w %d: address %d + length %d", ErrExceedsMemorySize, len(m.data), toAddress, len(source))
	}
	for i, value := range source {
		m.data[toAddress+uint(i)] = value
		m.initialized[toAddress+uint(i)] = true
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

func (m *Memory) AddProgram(program *Program) (offset uint, err error) {
	if len(program.Instructions()) < 1 {
		return 0, ErrProgramEmpty
	}
	if len(program.Instructions()) > int(m.Size()) {
		return 0, fmt.Errorf("%w %d: program length %d", ErrExceedsMemorySize, m.Size(), len(program.Instructions()))
	}
	if program.HasOrigin() {
		offset := program.Origin()
		return offset, m.AddProgramAtOffset(program, offset)
	}
	for offset := 32 - len(program.Instructions()); offset >= 0; offset-- {
		err := m.AddProgramAtOffset(program, uint(offset))
		if !errors.Is(err, ErrProgramOverlapsExistingPrograms) {
			return uint(offset), err
		}
	}
	return 0, ErrProgramCannotFit
}

var ErrProgramOverlapsExistingPrograms = errors.New("program overlaps existing programs")

func (m *Memory) AddProgramAtOffset(program *Program, offset uint) error {
	if len(program.Instructions()) < 1 {
		return ErrProgramEmpty
	}
	if offset+uint(len(program.Instructions())) > m.Size() {
		return fmt.Errorf("%w %d: offset %d + program length %d", ErrExceedsMemorySize, m.Size(), offset, len(program.Instructions()))
	}
	var programMask uint32 = ((0b1 << len(program.Instructions())) - 1) << offset
	if programMask&m.programUsageMask != 0 {
		return ErrProgramOverlapsExistingPrograms
	}
	updatedInstructions := slices.Clone(program.Instructions())
	for i, instruction := range updatedInstructions {
		if instr.DecodeType(instruction) == instr.TypeJump {
			originalJump := updatedInstructions[i]
			// Increase the target of all jump instructions by the offset amount.
			updatedInstructions[i] = originalJump&0b1111111111100000 | (originalJump+uint16(offset))&0b11111
		}
	}
	err := m.Copy(updatedInstructions, offset)
	if err != nil {
		return fmt.Errorf("copying program: %w", err)
	}
	m.programUsageMask |= programMask
	return nil
}

func (m *Memory) RemoveProgram(program *Program, loadedOffset uint) error {
	if len(program.instructions) < 1 {
		return ErrProgramEmpty
	}
	if loadedOffset+uint(len(program.instructions)) > m.Size() {
		return fmt.Errorf("%w %d: offset %d + program length %d", ErrExceedsMemorySize, m.Size(), loadedOffset, len(program.instructions))
	}
	var programMask uint32 = ((0b1 << len(program.instructions)) - 1) << loadedOffset
	for i := 0; i < len(program.instructions); i++ {
		m.initialized[loadedOffset+uint(i)] = false
	}
	m.programUsageMask &= ^programMask
	return nil
}
