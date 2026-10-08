# pio

A Go package that emulates the RP2040's PIO hardware.

See the [examples](examples) for usage.
See [pio-gtk](https://github.com/headblockhead/pio-gtk) for a GUI version.

## Features, Accuracy, and Limitations

### Features

- Simultaneous emulation of multiple electrically-interconnected RP2040s.
- Complete emulation of the RP2040's PIO block, including all state machines.
- Full GPIO feature simulation (pull-ups, pull-downs, overrides).
- Cycle-accurate emulation, even when using clock-dividers.
- Usable inside of test suites or as part of another program.
- Provides errors in potentially unintended scenarios (such as performing an undefined operation, or using the value of an uninitialised register).

### Accuracy

This emulator has been made as accurate as possible to the real RP2040's hardware.
However, it does not simulate any analogue components of the RP2040, such as the output drive strength, output slew rate, or input hysteresis of its pads.
Instead, this emulator opts to use a simplified digital model, which is enough to simulate regular driving of pins, pull-ups, pull-downs, and the delay caused by the pad's input/output synchronisers.

### Limitations

Some areas of the PIO's logic are not fully publicly documented in as thorough detail as the rest, and while most of the time I was able to confirm using actual hardware, some educated assumptions have been used in areas where documentation is particularly lacking, and the functionality is inconvenient to create a hardware test for (I'm looking at you, `OUT_STICKY` and `INLINE_OUT_EN`). This should not affect functionality in the vast majority of cases.

Finally, this package does not (yet) include an assembler. This means that instructions must be converted to bytecode before they are run using this emulator. This may change in the future.
