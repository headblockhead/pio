package pio

import "fmt"

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
	for i, circuit := range s.circuits {
		err := circuit.Solve()
		if err != nil {
			return fmt.Errorf("error solving circuit with index %d: %w", i, err)
		}
	}
	return nil
}
