# Differential Manchester

This is an example of a two rp2040s running the Differential Manchester example from the [pico-examples](https://github.com/raspberrypi/pico-examples) repository to communicate some data between them.
This is a slight deviation from the original example in order to further showcase the emulator.
Rather than a single rp2040 communicating by itself as in the original, this example program emulates two rp2040s communicating to eachother.
The setup for the state machines is replicated as closely as possible to the original `differential_manchester_tx_program_init` and `differential_manchester_rx_program_init` functions.
Where appropriate, lines are commented with the equivalent [pico-sdk](https://github.com/raspberrypi/pico-sdk) function calls.
Additionally, this version of the example is designed to run continuously using random data until stopped, rather than transmitting a specific test sequence once.

The program produces a visual depiction of some parts from the state machines of both rp2040s, shown as two columns of information.
Useful register values are shown at the top, and a drawing of the program being executed is provided beneath, with an arrow showing the instruction being executed.
Finally, at the bottom, a trace of the logic level of the net connecting the two RP2040s is depicted, scrolling right to left.

The important functionality of this example can be found in the `main.go` file.
The terminal graphics are separated into the `tui.go` file, such that they do not obscure the useful elements of the `main.go` file to a potential reader.

An example of the output of this program is provided below:

```console
pio/examples/differentialManchester$ go run .
         TX Program (RP2040A)                    RX Program (RP2040B)
PC:   2                                 PC:   8       
X:    1                                 X:    1       
Y:    0                                 Y:    0       
dlay: 5                                 dlay: 0       
OSR:  00000000000101100010101000111101  ISR:  0b01000110010000000000000000000000

0: out   x, 1                           0: wait  1 pin, 0 [11]
1: jmp   !x, 4 side 1 [6]               1: jmp   pin, 4       
>> nop                                  2: in    x, 1         
3: jmp   0     side 0 [6]               3: jmp   0            
4: jmp   5            [7]               4: in    y, 1     [1] 
5: out   x, 1                           5: wait  0 pin, 0 [11]
6: jmp   !x, 9 side 0 [6]               6: jmp   pin, 9       
7: nop                                  7: in    y, 1         
8: jmp   5     side 1 [6]               >> jmp   0            
9: jmp   0            [7]               9: in    x, 1     [1] 

--------------------------------------------------------------------------------
      #########################            #############                        
......                         ............             ........................
--------------------------------------------------------------------------------
```
