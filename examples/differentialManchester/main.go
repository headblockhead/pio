package main

import (
	"fmt"

	"github.com/headblockhead/pio"
	"github.com/headblockhead/pio/gpio"
	"github.com/headblockhead/pio/simulation"
	"github.com/headblockhead/pio/statemachine"
)

// programTX and programRX are exact assembled versions of the differential_manchester example from commit 7fe60d6 of the Raspberry Pi pico-examples GitHub repository.
// See README.md for how this code deviates from the official example.

var programTX = []uint16{
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
}

const programTXWrapFrom = 9
const programTXWrapTo = 0

var programRX = []uint16{
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
}

const programRXWrapFrom = 9
const programRXWrapTo = 5

var aPin uint = 2
var bPin uint = 3

// Errors are ignored throughout this example to help readability.
// Where errors are ignored, an underscore is used to demonstrate that the return value is intentionally being ignored.
// Please note that you should not normally ignore errors when using this package.

func main() {
	// Create a new simulation environment.
	s := simulation.New()

	// Setup both RP2040s' initial states.
	rA := setupRP2040A()
	rB := setupRP2040B()

	// Attach the RP2040s to the simulation instance so that they can be ticked collectively easily.
	_ = s.AddTicker(rA)
	_ = s.AddTicker(rB)

	// Create a shared net for the RP2040s to communicate over, and add it to the simulation so it will be solved each tick.
	signalNet := simulation.NewNet("BMC Signal")
	_ = s.AddNet(signalNet)

	// Attach both RP2040s to the net.
	conA, _ := rA.Connection(aPin)
	_ = signalNet.Connect(conA)
	conB, _ := rB.Connection(bPin)
	_ = signalNet.Connect(conB)

	// The simulation is now ready!
	// To run the simulation, use the Tick function.
	// Each Tick is equivalent to one cycle of the rp2040's system clock.

	// In this simulation, RP2040 A will transmit three 32-bit words out from its TX FIFO and over signalNet to RP2040 B, which will receive them into its RX FIFO.
	// The value of the signal net connecting RP2040 A and B will be printed to the console each tick using "." for low and "#" for high.
	// When RP2040 B receives a word, it will be printed to the console along with the tick cycle in which it was received.

	aPIO0, _ := rA.PIO(0)
	aSM0, _ := aPIO0.StateMachine(0)

	tx := aSM0.FIFOTXWriter()
	_ = tx.Write(0x00000000) // pio_sm_put_blocking
	_ = tx.Write(0x0ff0a55a) // pio_sm_put_blocking
	_ = tx.Write(0x12345678) // pio_sm_put_blocking

	bPIO0, _ := rB.PIO(0)
	bSM0, _ := bPIO0.StateMachine(0)
	bRXReader := bSM0.FIFORXReader()

	ticksToRun := 2500
	for i := range ticksToRun {
		// Tick the simulation forward one cycle.
		err := s.Tick()
		if err != nil {
			panic(err)
		}

		// Every 100 ticks, print the current tick number and create a new line.
		if i%100 == 0 {
			if i != 0 {
				fmt.Print("\n")
			}
			fmt.Printf("Tick %04d: ", i)
		}

		// Print the current value of the signal net each tick.
		if signalNet.LogicLevel() {
			fmt.Print("#")
		} else {
			fmt.Print(".")
		}

		// If RP2040 B's RX FIFO is not empty, it has received a word.
		if !bRXReader.IsEmpty() {
			// Read the word from RP2040 B's RX FIFO and print it to the console.
			data, _ := bRXReader.Read() // pio_sm_get_blocking
			fmt.Printf("\n\nRP2040 B Received 0x%08X on tick %d\n\n           ", data, i)
			for range i%100 + 1 {
				fmt.Print(" ")
			}
		}
	}

	fmt.Printf("\n\n%d ticks executed.\n", ticksToRun)
}

func setupRP2040A() *pio.RP2040 {
	r := pio.NewRP2040("RP2040 A (TX)")

	pio0, _ := r.PIO(0)

	mw := pio0.MemoryWriter()
	_ = mw.Copy(programTX, 0) // pio_add_program

	_ = pio0.SetPinOutput(aPin, false)      // pio_sm_set_pin
	_ = pio0.SetPinOutputEnable(aPin, true) // pio_sm_set_pindirs

	txGPIO, _ := r.GPIO(aPin)
	txGPIO.SetFunction(gpio.FunctionPIO0) // pio_gpio_init

	sm0, _ := pio0.StateMachine(0)

	// sm_config_set_wrap
	_ = sm0.SetWrapFromAddress(programTXWrapFrom)
	_ = sm0.SetWrapToAddress(programTXWrapTo)

	// sm_config_set_sideset
	_ = sm0.SetBitCountSideset(2)
	sm0.SetSidesetIsOptional(true)
	sm0.SetSidesetControlsPinDirection(false)

	// sm_config_set_sideset_pins
	_ = sm0.SetBaseSidesetPin(aPin)

	// sm_config_set_out_shift
	sm0.SetOutShiftMovesRight(true)
	sm0.SetAutopullEnabled(true)
	_ = sm0.SetPullThreshold(32)

	_ = sm0.SetFIFOJoin(statemachine.FIFOJoinTX) // sm_config_set_fifo_join

	_ = sm0.SetClockDivisor(125.0 / (16 * 5)) // sm_config_set_clkdiv

	var blockingPullInstruction uint16 = 0b1000000010100000 // pio_encode_pull(false, true)
	sm0.ForceInstruction(blockingPullInstruction)           // pio_sm_exec

	sm0.SetEnabled(true) // pio_sm_set_enabled

	return r
}

func setupRP2040B() *pio.RP2040 {
	r := pio.NewRP2040("RP2040 B (RX)")

	pio0, _ := r.PIO(0)

	mw := pio0.MemoryWriter()
	_ = mw.Copy(programRX, 0) // pio_add_program

	_ = pio0.SetPinOutputEnable(bPin, false) // pio_sm_set_pindirs

	rxGPIO, _ := r.GPIO(bPin)
	rxGPIO.SetFunction(gpio.FunctionPIO0) // pio_gpio_init

	sm0, _ := pio0.StateMachine(0)

	// sm_config_set_wrap
	_ = sm0.SetWrapFromAddress(programRXWrapFrom)
	_ = sm0.SetWrapToAddress(programRXWrapTo)

	_ = sm0.SetBaseInPin(bPin) // sm_config_set_in_pins
	_ = sm0.SetJumpPin(bPin)   // sm_config_set_jmp_pin

	// sm_config_set_in_shift
	sm0.SetInShiftMovesRight(true)
	sm0.SetAutopushEnabled(true)
	_ = sm0.SetPushThreshold(32)

	_ = sm0.SetFIFOJoin(statemachine.FIFOJoinRX) // sm_config_set_fifo_join

	_ = sm0.SetClockDivisor(125.0 / (16 * 5)) // sm_config_set_clkdiv

	var setXInstruction uint16 = 0b1110000000100001 // pio_encode_set(pio_x, 1)
	var setYInstruction uint16 = 0b1110000001000000 // pio_encode_set(pio_y, 0)
	sm0.ForceInstruction(setXInstruction)           // pio_sm_exec
	// The state machine must be manually ticked forward once here, as unlike in hardware, it is not continuously running in the background.
	// If it was not ticked here, the second ForceInstruction call would correctly overwrite the first, before the first instruction had a chance to execute.
	_ = sm0.Tick()
	sm0.ForceInstruction(setYInstruction) // pio_sm_exec

	sm0.SetEnabled(true) // pio_sm_set_enabled

	return r
}
