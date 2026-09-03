package connection

import (
	"errors"
	"fmt"
)

type Simulation struct {
	nets map[string]*net
}

func NewSimulation() *Simulation {
	return &Simulation{}
}

var ErrNetAlreadyExists = errors.New("net already exists")

func (s *Simulation) CreateNet(id string) error {
	_, exists := s.nets[id]
	if exists {
		return ErrNetAlreadyExists
	}
	s.nets[id] = NewNet()
	return nil
}

var ErrNetNotFound = errors.New("net not found")

func (s *Simulation) Connect(c Connection, netID string) error {
	net, ok := s.nets[netID]
	if !ok {
		return ErrNetNotFound
	}
	return net.connect(c)
}

func (s *Simulation) Solve() error {
	for id, n := range s.nets {
		err := n.solve()
		if err != nil {
			return fmt.Errorf("error solving net with id %s: %w", id, err)
		}
	}
	return nil
}
