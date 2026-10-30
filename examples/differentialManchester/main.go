package main

import (
	"math/rand/v2"
	"time"

	"git.edwardh.dev/pio"
	"git.edwardh.dev/pio/gpio"
	"git.edwardh.dev/pio/memory"
	"git.edwardh.dev/pio/simulation"
	"git.edwardh.dev/pio/statemachine"
)

// programTX and programRX are exact assembled versions of the differential_manchester example from commit 7fe60d6 of the Raspberry Pi pico-examples GitHub repository.

var programTX = memory.NewProgramWithOrigin([]uint16{
	//     .wrap_target
	0x6021, //  0: out    x, 1
	0x1e24, //  1: jmp    !x, 4           side 1 [6]
	0xa042, //  2: nop
	0x1600, //  3: jmp    0               side 0 [6]
	0x0705, //  4: jmp    5                      [7]
	0x6021, //  5: out    x, 1
	0x1629, //  6: jmp    !x, 9           side 0 [6]
	0xa042, //  7: nop
	0x1e05, //  8: jmp    5               side 1 [6]
	0x0700, //  9: jmp    0                      [7]
	//     .wrap
}, 0)

const programTXWrapFrom = 9
const programTXWrapTo = 0

var programRX = memory.NewProgramWithOrigin([]uint16{
	0x2ba0, //  0: wait   1 pin, 0               [11]
	0x00c4, //  1: jmp    pin, 4
	0x4021, //  2: in     x, 1
	0x0000, //  3: jmp    0
	0x4141, //  4: in     y, 1                   [1]
	//     .wrap_target
	0x2b20, //  5: wait   0 pin, 0               [11]
	0x00c9, //  6: jmp    pin, 9
	0x4041, //  7: in     y, 1
	0x0000, //  8: jmp    0
	0x4121, //  9: in     x, 1                   [1]
	//     .wrap
}, 0)

const programRXWrapFrom = 9
const programRXWrapTo = 5

// Errors are ignored throughout this example to help readability.
// Where errors are ignored, an underscore is used to demonstrate that the return value is intentionally not used.
// Please note that you should not normally ignore errors when using this package.

func main() {
	// Create a new simulation environment.
	// A simulation may contain many RP2040s, and many nets to electrically connect those RP2040s.
	// A simulation can be 'ticked', where each tick is equivalent to one CPU cycle of all RP2040s.
	// (Note that you may choose not to use the simulation structure whatsoever, RP2040s and nets can be ticked and solved without it, it just provides a convenient way to 'tick' many things at the same time.)
	s := simulation.New()

	// Create both RP2040s, and setup their configurations.
	// See the functions setupA/B at the end of this file for details.
	rA := pio.NewRP2040("RP2040 A (TX)")
	setupA(rA)
	rB := pio.NewRP2040("RP2040 B (RX)")
	setupB(rB)

	// Attach the RP2040s to the simulation instance so that they can be ticked by/with the simulation.
	// Note that anything which implements the simulation.Ticker interface can be ticked using a simulation, meaning you man implement your own tickable structures and add them, if you desire.
	_ = s.AddTicker(rA)
	_ = s.AddTicker(rB)

	// Create a shared net for the RP2040s to communicate over, and add it to the simulation.
	// A net attaches multiple connections (eg RP2040 GPIOs) together, like wiring things together on a breadboard.
	// A net combines the output signal of many connections together to form a single shared state, which is used as the input of those connections.
	// The process of finding the shared state is called 'solving' the net.
	// Adding a net to the simulation means that it will be solved whenever the simulation is ticked.
	// Note that anything which implements the simulation.Connection interface can be added to a net, meaning you may implement your own connections and add them, if you desire.
	signalNet := simulation.NewNet("BMC Signal")
	_ = s.AddNet(signalNet)

	// Attach pin 2 on RP2040 A to the signal net.
	conA, _ := rA.Connection(2)
	_ = signalNet.Connect(conA)

	// Attach pin 5 on RP2040 B to the signal net.
	conB, _ := rB.Connection(5)
	_ = signalNet.Connect(conB)

	// The simulation is now ready!
	// Both RP2040s have been setup, and are electrically connected.

	// In this example, RP2040 A will transmit many 32-bit words out from its TX FIFO and over the signalNet to RP2040 B, which will receive them into its RX FIFO.
	// The internal state of RP2040 A and RP2040 B's relevant state machines will be displayed in the terminal.
	// Additionally, the value of the signal net connecting RP2040 A and B will be displayed beneath, so that the transmission is clearly visible.

	aPIO0, _ := rA.PIO(0)
	aSM0, _ := aPIO0.StateMachine(0)
	aTXWriter := aSM0.FIFOTXWriter()

	bPIO0, _ := rB.PIO(0)
	bSM0, _ := bPIO0.StateMachine(0)
	bRXReader := bSM0.FIFORXReader()

	// The simulation will be run forever until it is manually stopped.
	for {
		// Draw the state to the terminal. (see tui.go)
		programStateDraw(programStateRender(aSM0, bSM0, signalNet.LogicLevel()))

		// If RP2040 A's TX FIFO is not full, it will be given another word to transmit.
		if !aTXWriter.IsFull() {
			_ = aTXWriter.Write(rand.Uint32())
		}

		// Tick the whole simulation forward one cycle.
		err := s.Tick()
		if err != nil {
			panic(err)
		}

		// If RP2040 B's RX FIFO is not empty, it has received a word.
		if !bRXReader.IsEmpty() {
			// Read the word from RP2040 B's RX FIFO.
			incomingWord, _ := bRXReader.Read() // pio_sm_get_blocking
			// The word is discarded here in this example.
			_ = incomingWord
		}

		// Wait 16ms inbetween each tick so the terminal output is visible.
		time.Sleep(16 * time.Millisecond)
	}
}

func setupA(r *pio.RP2040) {
	var aPin uint = 2
	pio0, _ := r.PIO(0)

	mw := pio0.MemoryWriter()
	offset, _ := mw.AddProgram(programTX) // pio_add_program

	_ = pio0.SetPinOutput(aPin, false)      // pio_sm_set_pin
	_ = pio0.SetPinOutputEnable(aPin, true) // pio_sm_set_pindirs

	txGPIO, _ := r.GPIO(aPin)
	txGPIO.SetFunction(gpio.FunctionPIO0) // pio_gpio_init

	sm0, _ := pio0.StateMachine(0)

	_ = sm0.SetWrap(programTXWrapTo+offset, programTXWrapFrom+offset) // sm_config_set_wrap
	_ = sm0.SetSideset(2, true, false)                                // sm_config_set_sideset
	_ = sm0.SetBaseSidesetPin(aPin)                                   // sm_config_set_sideset_pins
	_ = sm0.SetOutShift(true, true, 32)                               // sm_config_set_out_shift
	_ = sm0.SetFIFOJoin(statemachine.FIFOJoinTX)                      // sm_config_set_fifo_join
	_ = sm0.SetClockDivisor(125.0 / (16 * 5))                         // sm_config_set_clkdiv

	sm0.Init(offset) // pio_sm_init

	var blockingPullInstruction uint16 = 0b1000000010100000 // pio_encode_pull(false, true)
	sm0.ForceInstruction(blockingPullInstruction)           // pio_sm_exec

	sm0.SetEnabled(true) // pio_sm_set_enabled
}

func setupB(r *pio.RP2040) {
	var bPin uint = 5
	pio0, _ := r.PIO(0)

	mw := pio0.MemoryWriter()
	offset, _ := mw.AddProgram(programRX) // pio_add_program

	_ = pio0.SetPinOutputEnable(bPin, false) // pio_sm_set_pindirs

	rxGPIO, _ := r.GPIO(bPin)
	rxGPIO.SetFunction(gpio.FunctionPIO0) // pio_gpio_init

	sm0, _ := pio0.StateMachine(0)

	_ = sm0.SetWrap(programRXWrapTo+offset, programRXWrapFrom+offset) // sm_config_set_wrap
	_ = sm0.SetBaseInPin(bPin)                                        // sm_config_set_in_pins
	_ = sm0.SetJumpPin(bPin)                                          // sm_config_set_jmp_pin
	_ = sm0.SetInShift(true, true, 32)                                // sm_config_set_in_shift
	_ = sm0.SetFIFOJoin(statemachine.FIFOJoinRX)                      // sm_config_set_fifo_join
	_ = sm0.SetClockDivisor(125.0 / (16 * 5))                         // sm_config_set_clkdiv

	sm0.Init(offset) // pio_sm_init

	var setXInstruction uint16 = 0b1110000000100001 // pio_encode_set(pio_x, 1)
	var setYInstruction uint16 = 0b1110000001000000 // pio_encode_set(pio_y, 0)
	sm0.ForceInstruction(setXInstruction)           // pio_sm_exec
	// The state machine must be manually ticked forward once here, as unlike in hardware, it is not continuously running in the background.
	// If it was not ticked here, the second ForceInstruction call would correctly overwrite the first, before the first instruction had a chance to execute.
	_ = sm0.Tick()
	sm0.ForceInstruction(setYInstruction) // pio_sm_exec

	sm0.SetEnabled(true) // pio_sm_set_enabled
}
