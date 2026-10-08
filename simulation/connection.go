package simulation

type ConnectionState uint

const (
	ConnectionStateNone ConnectionState = iota
	ConnectionStateOutHigh
	ConnectionStateOutLow
	ConnectionStateBusKeeper
	ConnectionStatePullUp
	ConnectionStatePullDown
)

type Connection interface {
	Component

	State() ConnectionState
	SetInput(logicLevel bool)
}
