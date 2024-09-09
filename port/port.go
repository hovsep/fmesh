package port

import (
	"github.com/hovsep/fmesh/signal"
)

type Port struct {
	Signal signal.SignalInterface
	Pipes  Pipes //Refs to Pipes connected to that port (no in\out semantics)
}

type Ports map[string]*Port

func (p *Port) GetSignal() signal.SignalInterface {
	if p == nil {
		panic("invalid port")
	}

	if !p.HasSignal() {
		return nil
	}

	if p.Signal.IsAggregate() {
		return p.Signal.(*signal.Signals)
	}
	return p.Signal.(*signal.Signal)
}

func (p *Port) PutSignal(sig signal.SignalInterface) {
	if p.HasSignal() {
		//Aggregate SignalInterface
		var resValues []*signal.Signal

		//Extract existing SignalInterface(s)
		if p.Signal.IsSingle() {
			resValues = append(resValues, p.Signal.(*signal.Signal))
		} else if p.Signal.IsAggregate() {
			resValues = p.Signal.(*signal.Signals).Payload
		}

		//Add new SignalInterface(s)
		if sig.IsSingle() {
			resValues = append(resValues, sig.(*signal.Signal))
		} else if sig.IsAggregate() {
			resValues = append(resValues, sig.(*signal.Signals).Payload...)
		}

		p.Signal = &signal.Signals{
			Payload: resValues,
		}
		return
	}

	//Single SignalInterface
	p.Signal = sig
}

func (p *Port) ClearSignal() {
	p.Signal = nil
}

func (p *Port) HasSignal() bool {
	return p.Signal != nil
}

// Adds pipe reference to port, so all Pipes of the port are easily iterable (no in\out semantics)
func (p *Port) addPipeRef(pipe *Pipe) {
	p.Pipes = append(p.Pipes, pipe)
}

// PipeTo creates multiple pipes to other ports
func (p *Port) PipeTo(toPorts ...*Port) {
	for _, toPort := range toPorts {
		newPipe := &Pipe{
			From: p,
			To:   toPort,
		}
		p.addPipeRef(newPipe)
		toPort.addPipeRef(newPipe)
	}

}

// @TODO: this type must have good tooling for working with collection
// like adding new ports, filtering and so on

// @TODO: add error handling (e.g. when port does not exist)
func (ports Ports) ByName(name string) *Port {
	return ports[name]
}

// Deprecated, use ByName instead
func (ports Ports) ManyByName(names ...string) Ports {
	selectedPorts := make(Ports)

	for _, name := range names {
		if p, ok := ports[name]; ok {
			selectedPorts[name] = p
		}
	}

	return selectedPorts
}

func (ports Ports) AnyHasValue() bool {
	for _, p := range ports {
		if p.HasSignal() {
			return true
		}
	}

	return false
}

func (ports Ports) AllHaveValue() bool {
	for _, p := range ports {
		if !p.HasSignal() {
			return false
		}
	}

	return true
}

func (ports Ports) SetAll(val signal.SignalInterface) {
	for _, p := range ports {
		p.PutSignal(val)
	}
}

func (ports Ports) ClearAll() {
	for _, p := range ports {
		p.ClearSignal()
	}
}

func ForwardSignal(source *Port, dest *Port) {
	dest.PutSignal(source.GetSignal())
}
