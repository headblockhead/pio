# Differential Manchester

This is an example of a two rp2040s running the Differential Manchester example from the [pico-examples](https://github.com/raspberrypi/pico-examples) repository to communicate some data between them.
This is a slight deviation from the original example in order to further showcase the emulator.
Rather than a single rp2040 communicating by itself as in the original, this example program emulates two rp2040s communicating to eachother.
The setup for the state machines is replicated as closely as possible to the original `differential_manchester_tx_program_init` and `differential_manchester_rx_program_init` functions.
Where appropriate, lines are commented with the equivalent [pico-sdk](https://github.com/raspberrypi/pico-sdk) function calls.
