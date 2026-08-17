package pio

type ConnectionCircuit interface {
	GetState() (PadState, error)
	SetInput(input PadInput) error
}

type Connection struct {
	belongsToPads PadsSolver
	pad           uint
}

func NewConnection(belongsToPads PadsSolver, pad uint) *Connection {
	return &Connection{
		belongsToPads: belongsToPads,
		pad:           pad,
	}
}

func (c *Connection) GetState() (PadState, error) {
	return c.belongsToPads.GetState(c.pad)
}

func (c *Connection) SetInput(input PadInput) error {
	return c.belongsToPads.SetInput(c.pad, input)
}
