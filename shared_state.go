package pio

type SharedState struct {
	// assumes that every GPIO of every machine is directly interconnected
	gpioSyncers []GPIOStateSyncer
	gpioStates  [32]bool
}

// for now, assumes every pin has a pullup
func (ss *SharedState) ReconcileGPIOs() error {
	newGPIOStates := [32]bool{true}
	for _, gss := range ss.gpioSyncers {
		//TODO
	}
}
