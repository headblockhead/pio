package gpio

type Function uint

const (
	FunctionNone Function = iota
	FunctionPIO0
	FunctionPIO1
)

type Setter interface {
	SetOutputEnableOverride(Override)
	SetOutputOverride(Override)
	SetInputOverride(Override)
	SetFunction(Function)
}

type Getter interface {
	GetOutputEnableOverride() Override
	GetOutputOverride() Override
	GetInputOverride() Override
	GetFunction() Function
}

type GPIO struct {
	outputEnableOverride Override
	outputOverride       Override
	inputOverride        Override
	function             Function
}

func NewGPIO() *GPIO {
	return &GPIO{}
}

func (g *GPIO) Setter() Setter {
	return g
}

func (g *GPIO) SetOutputEnableOverride(o Override) { g.outputEnableOverride = o }
func (g *GPIO) SetOutputOverride(o Override)       { g.outputOverride = o }
func (g *GPIO) SetInputOverride(o Override)        { g.inputOverride = o }
func (g *GPIO) SetFunction(f Function)             { g.function = f }

func (g *GPIO) Getter() Getter {
	return g
}

func (g *GPIO) GetOutputEnableOverride() Override { return g.outputEnableOverride }
func (g *GPIO) GetOutputOverride() Override       { return g.outputOverride }
func (g *GPIO) GetInputOverride() Override        { return g.inputOverride }
func (g *GPIO) GetFunction() Function             { return g.function }
