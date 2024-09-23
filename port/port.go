package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Port defines a connectivity point of a component
type Port struct {
	name    string
	signals signal.Group //Signal buffer
	pipes   Group        //Outbound pipes
}

// New creates a new port
func New(name string) *Port {
	return &Port{
		name:    name,
		pipes:   NewGroup(),
		signals: signal.NewGroup(),
	}
}

// Name getter
func (p *Port) Name() string {
	return p.name
}

// Signals getter
func (p *Port) Signals() signal.Group {
	return p.signals
}

// setSignals sets signals field
func (p *Port) setSignals(signals signal.Group) {
	p.signals = signals
}

// PutSignals adds signals
// @TODO: rename
func (p *Port) PutSignals(signals ...*signal.Signal) {
	p.setSignals(p.Signals().With(signals...))
}

// WithSignals adds signals and returns the port
func (p *Port) WithSignals(signals ...*signal.Signal) *Port {
	p.PutSignals(signals...)
	return p
}

// Clear removes all signals from the port
func (p *Port) Clear() {
	p.setSignals(signal.NewGroup())
}

// DisposeSignals removes n signals from the beginning of signal buffer
func (p *Port) DisposeSignals(n int) {
	p.setSignals(p.Signals()[n:])
}

// FlushAndDispose flushes n signals and then disposes them
func (p *Port) FlushAndDispose(n int) {
	if n > len(p.Signals()) {
		//Flush all signals and clear
		p.Flush()
		p.Clear()
	}

	if !p.HasSignals() || !p.HasPipes() {
		return
	}

	for _, outboundPort := range p.pipes {
		//Fan-Out
		ForwardNSignals(p, outboundPort, n)
	}

	p.DisposeSignals(n)
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
		p.pipes = p.pipes.With(toPort)
	}
}

// Flush pushes signals to pipes, clears the port if needed and returns true when flushed
// @TODO: hide this method from user
func (p *Port) Flush() bool {
	if !p.HasSignals() || !p.HasPipes() {
		return false
	}

	for _, outboundPort := range p.pipes {
		//Fan-Out
		ForwardSignals(p, outboundPort)
	}
	return true
}

// ForwardSignals copies all signals from source port to destination port, without clearing the source port
func ForwardSignals(source *Port, dest *Port) {
	dest.PutSignals(source.Signals()...)
}

// ForwardNSignals forwards n signals
func ForwardNSignals(source *Port, dest *Port, n int) {
	dest.PutSignals(source.Signals()[:n]...)
}
