package conn

import (
	"errors"
	"fmt"
)

type Ticker interface {
	ID() string
	Tick() error
}

type ConnectionBlock interface {
	Connection(uint) (c Connection, ok bool)
}

type Simulation struct {
	nets    map[string]*net
	tickers map[string]Ticker
}

func NewSimulation() *Simulation {
	return &Simulation{
		nets:    make(map[string]*net),
		tickers: make(map[string]Ticker),
	}
}

func (s *Simulation) Tick() error {
	err := s.Solve()
	if err != nil {
		return fmt.Errorf("solve: %w", err)
	}
	for id, t := range s.tickers {
		err := t.Tick()
		if err != nil {
			return fmt.Errorf("ticker %s: %w", id, err)
		}
	}
	return nil
}

var ErrNetAlreadyExists = errors.New("net already exists")

func (s *Simulation) AddNet(id string) error {
	_, exists := s.nets[id]
	if exists {
		return ErrNetAlreadyExists
	}
	s.nets[id] = newNet()
	return nil
}

func (s *Simulation) RemoveNet(id string) error {
	_, exists := s.nets[id]
	if !exists {
		return ErrNetNotFound
	}
	delete(s.nets, id)
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

var ErrConnectionNotFoundInBlock = errors.New("connection not found in block")

func (s *Simulation) ConnectFromBlock(b ConnectionBlock, i uint, netID string) error {
	c, ok := b.Connection(i)
	if !ok {
		return ErrConnectionNotFoundInBlock
	}
	return s.Connect(c, netID)
}

func (s *Simulation) Disconnect(c Connection, netID string) error {
	net, ok := s.nets[netID]
	if !ok {
		return ErrNetNotFound
	}
	return net.disconnect(c)
}

func (s *Simulation) DisconnectFromBlock(b ConnectionBlock, i uint, netID string) error {
	c, ok := b.Connection(i)
	if !ok {
		return ErrConnectionNotFoundInBlock
	}
	return s.Disconnect(c, netID)
}

var ErrTickerAlreadyExists = errors.New("ticker already exists")

func (s *Simulation) AddTicker(t Ticker) error {
	_, exists := s.tickers[t.ID()]
	if exists {
		return ErrTickerAlreadyExists
	}
	s.tickers[t.ID()] = t
	return nil
}

var ErrTickerNotFound = errors.New("ticker not found")

func (s *Simulation) RemoveTicker(id string) error {
	_, exists := s.tickers[id]
	if !exists {
		return ErrTickerNotFound
	}
	delete(s.tickers, id)
	return nil
}

func (s *Simulation) Solve() error {
	for id, n := range s.nets {
		err := n.solve()
		if err != nil {
			return fmt.Errorf("net %s: %w", id, err)
		}
	}
	return nil
}

func (s *Simulation) ListNets() []string {
	ids := make([]string, 0, len(s.nets))
	for id := range s.nets {
		ids = append(ids, id)
	}
	return ids
}

func (s *Simulation) GetNetIsHigh(id string) (bool, error) {
	net, ok := s.nets[id]
	if !ok {
		return false, ErrNetNotFound
	}
	return net.isHigh, nil
}
