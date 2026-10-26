package main

import (
	"fmt"
	"git.edwardh.dev/pio/statemachine"
	"slices"
)

// This file deals with drawing the state of the RP2040s to the terminal, and is a bit messy.
// For the usage example, see main.go.

var programTXStrings = []string{
	"0: out   x, 1             ",
	"1: jmp   !x, 4 side 1 [6] ",
	"2: nop                    ",
	"3: jmp   0     side 0 [6] ",
	"4: jmp   5            [7] ",
	"5: out   x, 1             ",
	"6: jmp   !x, 9 side 0 [6] ",
	"7: nop                    ",
	"8: jmp   5     side 1 [6] ",
	"9: jmp   0            [7] ",
}

var programRXStrings = []string{
	"0: wait  1 pin, 0 [11]",
	"1: jmp   pin, 4       ",
	"2: in    x, 1         ",
	"3: jmp   0            ",
	"4: in    y, 1     [1] ",
	"5: wait  0 pin, 0 [11]",
	"6: jmp   pin, 9       ",
	"7: in    y, 1         ",
	"8: jmp   0            ",
	"9: in    x, 1     [1] ",
}

var logicHistory []bool = []bool{}
var line string = "--------------------------------------------------------------------------------"

func programStateRender(smA *statemachine.StateMachine, smB *statemachine.StateMachine, currentLogic bool) []string {
	logicHistory = append(logicHistory, currentLogic)
	if len(logicHistory) > len(line) {
		logicHistory = logicHistory[1:]
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
		finalProgs = append(finalProgs, string(progTXNew[i]+"              "+progRXNew[i]))
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
