package connection

type State uint

const (
	StateNone State = iota
	StateOutHigh
	StateOutLow
	StateBusKeeper
	StatePullUp
	StatePullDown
)

type Connection interface {
	ID() string
	GetState() State
	SetInput(bool)
}
