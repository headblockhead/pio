package main

import (
	"fmt"
	"slices"

	"git.edwardh.dev/pio/memory"
	"git.edwardh.dev/pio/statemachine"
)

// This file deals with drawing the state of the RP2040s to the terminal, and is a bit messy.
// For the usage example, see main.go.

var logicHistory []bool = []bool{}
var line string = "--------------------------------------------------------------------------------"

func programStateRender(programTX *memory.Program, programRX *memory.Program, smA *statemachine.StateMachine, smB *statemachine.StateMachine, currentLogic bool) []string {
	logicHistory = append(logicHistory, currentLogic)
	if len(logicHistory) > len(line) {
		logicHistory = logicHistory[1:]
	}

	txPinCount := smA.BitCountSideset()
	if smA.SidesetIsOptional() {
		txPinCount--
	}
	programTXStrings, err := programTX.Strings(txPinCount, smA.SidesetIsOptional())
	if err != nil {
		panic(err)
	}

	rxPinCount := smB.BitCountSideset()
	if smB.SidesetIsOptional() {
		rxPinCount--
	}
	programRXStrings, err := programRX.Strings(rxPinCount, smB.SidesetIsOptional())
	if err != nil {
		panic(err)
	}

	pcA := smA.ProgramCounter()
	pcB := smB.ProgramCounter()

	line0 := "         \x1b[4mTX Program (RP2040A)\033[0m                    \x1b[4mRX Program (RP2040B)\033[0m"
	line1 := fmt.Sprintf("PC:   %d                                 PC:   %d       ", pcA, pcB)
	line2 := fmt.Sprintf("X:    %d                                 X:    %d       ", smA.XRegister(), smB.XRegister())
	line3 := fmt.Sprintf("Y:    %d                                 Y:    %d       ", smA.YRegister(), smB.YRegister())
	line4 := fmt.Sprintf("dlay: %d                                 dlay: %d       ", smA.DelaysRemaining(), smB.DelaysRemaining())
	line5 := fmt.Sprintf("OSR:  %032b  ISR:  0b%032b", smA.OutputShiftRegister(), smB.InputShiftRegister())

	progTXNew := slices.Clone(programTXStrings)
	currentLine := []rune(progTXNew[pcA])
	currentLine[0] = '>'
	currentLine[1] = '>'
	progTXNew[pcA] = string(append([]rune("\033[35m"), append(currentLine, []rune("\033[0m")...)...))

	progRXNew := slices.Clone(programRXStrings)
	currentLine = []rune(progRXNew[pcB])
	currentLine[0] = '>'
	currentLine[1] = '>'
	progRXNew[pcB] = string(append([]rune("\033[35m"), append(currentLine, []rune("\033[0m")...)...))

	finalProgs := []string{line0, line1, line2, line3, line4, line5, ""}
	for i := range max(len(programTXStrings), len(programRXStrings)) {
		finalProgs = append(finalProgs, string(progTXNew[i]+progRXNew[i]))
	}

	finalProgs = append(finalProgs, "")
	finalProgs = append(finalProgs, line)
	dataLine1 := []rune{}
	dataLine2 := []rune{}
	for _, v := range logicHistory {
		if v {
			dataLine1 = append(dataLine1, '#')
			dataLine2 = append(dataLine2, ' ')
		} else {
			dataLine1 = append(dataLine1, ' ')
			dataLine2 = append(dataLine2, '.')
		}
	}
	finalProgs = append(finalProgs, string(dataLine1))
	finalProgs = append(finalProgs, string(dataLine2))
	finalProgs = append(finalProgs, line)

	return finalProgs
}

var alreadyDrawn bool = false

func programStateDraw(renderedState []string) {
	if !alreadyDrawn {
		for range len(renderedState) {
			fmt.Println()
		}
		alreadyDrawn = true
	}

	fmt.Print("\033[60D")
	for range len(renderedState) {
		fmt.Print("\033[1A")
		fmt.Print("\033[K")
	}
	for _, v := range renderedState {
		fmt.Print(v)
		fmt.Println()
	}
}
