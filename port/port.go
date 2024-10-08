package port

import (
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/signal"
)

// Port defines a connectivity point of a component
type Port struct {
	common.NamedEntity
	common.LabeledEntity
	signals signal.Group //Signal buffer
	pipes   Group        //Outbound pipes
}

// New creates a new port
func New(name string) *Port {
	return &Port{
		NamedEntity: common.NewNamedEntity(name),
		pipes:       NewGroup(),
		signals:     signal.NewGroup(),
	}

}

// Signals getter
func (p *Port) Signals() signal.Group {
	return p.signals
}

// Pipes getter
func (p *Port) Pipes() Group {
	return p.pipes
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

// Flush pushes signals to pipes and clears the port
// @TODO: hide this method from user
func (p *Port) Flush() {
	if !p.HasSignals() || !p.HasPipes() {
		return
	}

	for _, outboundPort := range p.pipes {
		//Fan-Out
		ForwardSignals(p, outboundPort)
	}
	p.Clear()
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
func (p *Port) PipeTo(destPorts ...*Port) {
	for _, destPort := range destPorts {
		if destPort == nil {
			continue
		}
		p.pipes = p.pipes.With(destPort)
	}
}

// withPipes adds pipes and returns the port
func (p *Port) withPipes(destPorts ...*Port) *Port {
	for _, destPort := range destPorts {
		p.PipeTo(destPort)
	}
	return p
}

// WithLabels sets labels and returns the port
func (p *Port) WithLabels(labels common.LabelsCollection) *Port {
	p.LabeledEntity.SetLabels(labels)
	return p
}

// ForwardSignals copies all signals from source port to destination port, without clearing the source port
func ForwardSignals(source *Port, dest *Port) {
	dest.PutSignals(source.Signals()...)
}
