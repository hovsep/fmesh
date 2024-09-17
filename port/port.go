package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Port defines a connectivity point of a component
type Port struct {
	name   string
	signal *signal.Signal //Current signal set on the port
	pipes  Collection     //Refs to all outbound pipes connected to this port
}

// NewPort creates a new port
func NewPort(name string) *Port {
	return &Port{
		name:  name,
		pipes: NewPortsCollection(),
	}
}

// Name getter
func (p *Port) Name() string {
	return p.name
}

// Signal getter
func (p *Port) Signal() *signal.Signal {
	return p.signal
}

// PutSignal adds a signal to current signal
func (p *Port) PutSignal(sig *signal.Signal) {
	p.signal = sig.Combine(p.Signal())
}

// ClearSignal removes current signal from the port
func (p *Port) ClearSignal() {
	p.signal = nil
}

// HasSignal says whether port signal is set or not
func (p *Port) HasSignal() bool {
	return p.signal != nil
}

// PipeTo creates one or multiple pipes to other port(s)
func (p *Port) PipeTo(toPorts ...*Port) {
	for _, toPort := range toPorts {
		if toPort == nil {
			continue
		}
		p.pipes = p.pipes.Add(toPort)
	}
}

// Flush pushed current signal to pipes and clears the port
func (p *Port) Flush() {
	if !p.HasSignal() || len(p.pipes) == 0 {
		return
	}

	for _, outboundPort := range p.pipes {
		//Fan-Out
		ForwardSignal(p, outboundPort)
	}
	p.ClearSignal()
}

// ForwardSignal puts a signal from source port to destination port, without removing it on source port
func ForwardSignal(source *Port, dest *Port) {
	dest.PutSignal(source.Signal())
}
