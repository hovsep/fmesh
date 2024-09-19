package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Port defines a connectivity point of a component
type Port struct {
	name    string
	signals signal.Collection //Current signals set on the port
	pipes   Group             //Refs to all outbound pipes connected to this port
}

// New creates a new port
func New(name string) *Port {
	return &Port{
		name:    name,
		pipes:   NewGroup(),
		signals: signal.NewCollection(),
	}
}

// Name getter
func (p *Port) Name() string {
	return p.name
}

// Signals getter
func (p *Port) Signals() signal.Collection {
	return p.signals
}

// PutSignals adds signals
// @TODO: rename
func (p *Port) PutSignals(signals ...*signal.Signal) {
	p.Signals().Add(signals...)
}

// WithSignals adds signals and returns the port
func (p *Port) WithSignals(signals ...*signal.Signal) *Port {
	p.PutSignals(signals...)
	return p
}

// ClearSignals removes all signals from the port
func (p *Port) ClearSignals() {
	p.signals = signal.NewCollection()
}

// HasSignals says whether port signals is set or not
func (p *Port) HasSignals() bool {
	return len(p.Signals()) > 0
}

// HasPipes says whether port has outbound pipes
func (p *Port) HasPipes() bool {
	return len(p.pipes) > 0
}

// PipeTo creates one or multiple pipes to other port(s)
// @TODO: hide this method from AF
func (p *Port) PipeTo(toPorts ...*Port) {
	for _, toPort := range toPorts {
		if toPort == nil {
			continue
		}
		p.pipes = p.pipes.Add(toPort)
	}
}

// Flush pushes signals to pipes, clears the port if needed and returns true when flushed
// @TODO: hide this method from user
func (p *Port) Flush(clearFlushed bool) bool {
	if !p.HasSignals() || !p.HasPipes() {
		return false
	}

	for _, outboundPort := range p.pipes {
		//Fan-Out
		ForwardSignals(p, outboundPort)
	}
	if clearFlushed {
		p.ClearSignals()
	}
	return true
}

// ForwardSignals puts signals from source port to destination port, without clearing the source port
func ForwardSignals(source *Port, dest *Port) {
	dest.PutSignals(source.Signals().AsGroup()...)
}
