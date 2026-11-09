package main

import (
	"fmt"
	"math/bits"

	"git.edwardh.dev/pio"
	"git.edwardh.dev/pio/gpio"
	"git.edwardh.dev/pio/instr"
	"git.edwardh.dev/pio/memory"
	"git.edwardh.dev/pio/simulation"
)

// programOutputObserver is a simple PIO program that writes a value to a pin, then reads back the value of the pin over and over again every cycle.
// This allows the number of cycles between writing a value to a pin, and reading that new value back, to be measured using the returned data.

var programOutputObserver = memory.NewProgram([]uint16{
	instr.NewSet(instr.SetDestinationPinDirections, 1).Encode(), //  0: set    pindirs, 1
	instr.NewSet(instr.SetDestinationPins, 0).Encode(),          //  1: set    pins, 0
	instr.EncodeNOP() | instr.NewDelay(31).Encode(),             //  2: nop                           [31]
	instr.NewSet(instr.SetDestinationPins, 1).Encode(),          //  3: set    pins, 1
	//     .wrap_target
	instr.NewIn(instr.InSourcePins, 1).Encode(), //  4: in     pins, 1
	//     .wrap
})

const programOutputObserverWrapFrom = 4
const programOutputObserverWrapTo = 4

// In real RP2040 hardware, inputOutputDelayCycles will only ever be 1 or 2, depending if the input/output synchronisers are disabled or enabled.
// However, since this is a simulation, any value >0 can be used, to simulate an arbitrary artificial delay.
const inputOutputDelayCycles uint = 2

func main() {
	s := simulation.New()

	r := pio.NewRP2040("RP2040")
	_ = s.AddTicker(r)

	pio0, _ := r.PIO(0)
	offset, _ := pio0.MemoryWriter().AddProgram(programOutputObserver)
	g, _ := r.GPIO(3)
	g.SetFunction(gpio.FunctionPIO0)
	sm0, _ := pio0.StateMachine(0)
	_ = sm0.SetBaseSetPin(3)
	_ = sm0.SetPinCountSet(1)
	_ = sm0.SetBaseInPin(3)
	_ = sm0.SetInShift(false, true, 32)
	_ = sm0.SetWrap(programOutputObserverWrapTo+offset, programOutputObserverWrapFrom+offset)
	_ = sm0.Init(offset)
	sm0.SetEnabled(true)

	p, _ := r.Pad(3)
	_ = p.SetInputDelayCycles(inputOutputDelayCycles)
	_ = p.SetOutputDelayCycles(inputOutputDelayCycles)
	n := simulation.NewNet("pin 3")
	_ = n.Connect(p)
	_ = s.AddNet(n)

	fifoRX := sm0.FIFORXReader()

	for fifoRX.IsEmpty() {
		_ = s.Tick()
	}

	result, _ := fifoRX.Read()
	leadingZerosCount := bits.LeadingZeros32(result)
	fmt.Printf("Result: 0b%032b (%d total cycles of delay)\n", result, leadingZerosCount+1)
}
