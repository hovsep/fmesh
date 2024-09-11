package port

import (
	"github.com/hovsep/fmesh/signal"
)

type Port struct {
	signal *signal.Signal
	Pipes  Pipes //Refs to Pipes connected to that port (no in\out semantics)
}

type Ports map[string]*Port

func (p *Port) Signal() *signal.Signal {
	return p.signal
}

func (p *Port) PutSignal(sig *signal.Signal) {
	p.signal = sig.Combine(p.Signal())
}

func (p *Port) ClearSignal() {
	p.signal = nil
}

func (p *Port) HasSignal() bool {
	return p.signal != nil
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

func (ports Ports) AnyHasSignal() bool {
	for _, p := range ports {
		if p.HasSignal() {
			return true
		}
	}

	return false
}

func (ports Ports) AllHaveSignal() bool {
	for _, p := range ports {
		if !p.HasSignal() {
			return false
		}
	}

	return true
}

func (ports Ports) PutSignal(sig *signal.Signal) {
	for _, p := range ports {
		p.PutSignal(sig)
	}
}

func (ports Ports) ClearAll() {
	for _, p := range ports {
		p.ClearSignal()
	}
}

func ForwardSignal(source *Port, dest *Port) {
	dest.PutSignal(source.Signal())
}
