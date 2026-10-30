package gpio

type Function uint

const (
	FunctionNone Function = iota
	FunctionPIO0
	FunctionPIO1
)

type Observer interface {
	OutputEnableOverride() Override
	OutputOverride() Override
	InputOverride() Override
	Function() Function
}

type Configurator interface {
	SetOutputEnableOverride(Override)
	SetOutputOverride(Override)
	SetInputOverride(Override)
	SetFunction(Function)
}

type Operator interface {
	Observer
}

type GPIO struct {
	outputEnableOverride Override
	outputOverride       Override
	inputOverride        Override
	function             Function
}

func New() *GPIO {
	return &GPIO{
		outputEnableOverride: OverrideNone,
		outputOverride:       OverrideNone,
		inputOverride:        OverrideNone,
		function:             FunctionNone,
	}
}

func (g *GPIO) Observer() Observer {
	return g
}

func (g *GPIO) OutputEnableOverride() Override { return g.outputEnableOverride }
func (g *GPIO) OutputOverride() Override       { return g.outputOverride }
func (g *GPIO) InputOverride() Override        { return g.inputOverride }
func (g *GPIO) Function() Function             { return g.function }

func (g *GPIO) Configurator() Configurator {
	return g
}

func (g *GPIO) Operator() Operator {
	return g
}

func (g *GPIO) SetOutputEnableOverride(o Override) { g.outputEnableOverride = o }
func (g *GPIO) SetOutputOverride(o Override)       { g.outputOverride = o }
func (g *GPIO) SetInputOverride(o Override)        { g.inputOverride = o }
func (g *GPIO) SetFunction(f Function)             { g.function = f }
