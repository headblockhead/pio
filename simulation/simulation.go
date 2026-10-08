package simulation

import (
	"errors"
	"fmt"
)

type Ticker interface {
	Component

	Tick() error
}

type SimulationObserver interface {
	Nets() map[ComponentIdentifier]*Net
	Tickers() map[ComponentIdentifier]Ticker
}

type SimulationConfigurator interface {
	AddNet(*Net) error
	RemoveNet(*Net) error
	AddTicker(Ticker) error
	RemoveTicker(Ticker) error
}

type SimulationOperator interface {
	Tick() error
}

type Simulation struct {
	nets    map[ComponentIdentifier]*Net
	tickers map[ComponentIdentifier]Ticker
}

func New() *Simulation {
	return &Simulation{
		nets:    make(map[ComponentIdentifier]*Net),
		tickers: make(map[ComponentIdentifier]Ticker),
	}
}

func (s *Simulation) Observer() SimulationObserver {
	return s
}

func (s *Simulation) Nets() map[ComponentIdentifier]*Net      { return s.nets }
func (s *Simulation) Tickers() map[ComponentIdentifier]Ticker { return s.tickers }

func (s *Simulation) Configurator() SimulationConfigurator {
	return s
}

var ErrSimulationNetAlreadyAdded = errors.New("already added")

func (s *Simulation) AddNet(net *Net) error {
	if _, exists := s.nets[net.ID()]; exists {
		return fmt.Errorf("net [%v]: %w", net.Label(), ErrSimulationNetAlreadyAdded)
	}
	s.nets[net.ID()] = net
	return nil
}

var ErrSimulationNetNotCurrentlyAdded = errors.New("not currently added")

func (s *Simulation) RemoveNet(net *Net) error {
	if _, exists := s.nets[net.ID()]; !exists {
		return fmt.Errorf("net [%v]: %w", net.Label(), ErrSimulationNetNotCurrentlyAdded)
	}
	delete(s.nets, net.ID())
	return nil
}

var ErrSimulationTickerAlreadyAdded = errors.New("already added")

func (s *Simulation) AddTicker(ticker Ticker) error {
	if _, exists := s.tickers[ticker.ID()]; exists {
		return fmt.Errorf("ticker [%v]: %w", ticker.Label(), ErrSimulationTickerAlreadyAdded)
	}
	s.tickers[ticker.ID()] = ticker
	return nil
}

var ErrSimulationTickerNotCurrentlyAdded = errors.New("not currently added")

func (s *Simulation) RemoveTicker(ticker Ticker) error {
	if _, exists := s.tickers[ticker.ID()]; !exists {
		return fmt.Errorf("ticker [%v]: %w", ticker.Label(), ErrSimulationTickerNotCurrentlyAdded)
	}
	delete(s.tickers, ticker.ID())
	return nil
}

func (s *Simulation) Operator() SimulationOperator {
	return s
}

func (s *Simulation) Tick() error {
	for _, net := range s.nets {
		if err := net.Solve(); err != nil {
			return fmt.Errorf("net [%v]: %w", net.Label(), err)
		}
	}
	for _, ticker := range s.tickers {
		if err := ticker.Tick(); err != nil {
			return fmt.Errorf("ticker [%v]: %w", ticker.Label(), err)
		}
	}
	return nil
}
