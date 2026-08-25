package connection

import (
	"errors"
	"fmt"
)

type Simulation struct {
	nets map[string]net
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
	s.nets[id] = net{}
	return nil
}

var ErrNetNotFound = errors.New("net not found")
var ErrAlreadyConnected = errors.New("already connected")

func (s *Simulation) Connect(c Connection, netID string) error {
	net, ok := s.nets[netID]
	if !ok {
		return ErrNetNotFound
	}
	_, exists := net.connections[c.ID()]
	if exists {
		return ErrAlreadyConnected
	}
	net.connections[c.ID()] = c
	return nil
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
