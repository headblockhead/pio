package pio

type Solver struct {
	circuits []CircuitSolver
}

func NewSolver() *Solver {
	return &Solver{}
}

func (s *Solver) AddCircuit(circuit CircuitSolver) error {
	s.circuits = append(s.circuits, circuit)
	return nil
}

func (s *Solver) Solve() error {
	for _, circuit := range s.circuits {
		err := circuit.Solve()
		if err != nil {
			return err
		}
	}
	return nil
}
