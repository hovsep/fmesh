package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Port defines a connectivity point of a component
type Port struct {
	name    string
	signals signal.Group //Current signals set on the port
	pipes   Group        //Refs to all outbound pipes connected to this port
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

// PutSignals adds a signals to current signals
func (p *Port) PutSignals(signals ...*signal.Signal) {
	for _, s := range signals {
		p.signals = append(p.signals, s)
	}
}

// ClearSignals removes current signals from the port
func (p *Port) ClearSignals() {
	p.signals = signal.NewGroup()
}

// HasSignals says whether port signals is set or not
func (p *Port) HasSignals() bool {
	return len(p.signals) > 0
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

// Flush pushed current signals to pipes and clears the port
func (p *Port) Flush() {
	if !p.HasSignals() || len(p.pipes) == 0 {
		return
	}

	for _, outboundPort := range p.pipes {
		//Fan-Out
		ForwardSignals(p, outboundPort)
	}
	p.ClearSignals()
}

// ForwardSignals puts signals from source port to destination port, without clearing the source port
func ForwardSignals(source *Port, dest *Port) {
	dest.PutSignals(source.Signals()...)
}
