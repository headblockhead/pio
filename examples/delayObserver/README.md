# Delay Observer

This is a simple example of a single rp2040 running a custom 'delay observer' program.
This program measures the amount of delay (in CPU cycles) between writing to a GPIO pin, and reading the updated state of that pin back.

An example of the output of this program is provided below:

```console
pio/examples/delayObserver$ go run .
Result: 0b00011111111111111111111111111111 (4 total cycles of delay)
```
