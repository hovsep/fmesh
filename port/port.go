package port

import (
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/signal"
)

// Port defines a connectivity point of a component
type Port struct {
	common.NamedEntity
	common.LabeledEntity
	*common.Chainable
	buffer *signal.Group
	pipes  *Group //Outbound pipes
}

// New creates a new port
func New(name string) *Port {
	return &Port{
		NamedEntity:   common.NewNamedEntity(name),
		LabeledEntity: common.NewLabeledEntity(nil),
		Chainable:     common.NewChainable(),
		pipes:         NewGroup(),
		buffer:        signal.NewGroup(),
	}

}

// Buffer getter
// @TODO: maybe we can hide this and return signals to user code
func (p *Port) Buffer() *signal.Group {
	if p.HasChainError() {
		return p.buffer.WithChainError(p.ChainError())
	}
	return p.buffer
}

// Pipes getter
// @TODO maybe better to return []*Port directly
func (p *Port) Pipes() *Group {
	if p.HasChainError() {
		return p.pipes.WithChainError(p.ChainError())
	}
	return p.pipes
}

// withBuffer sets buffer field
func (p *Port) withBuffer(buffer *signal.Group) *Port {
	if buffer.HasChainError() {
		return p.WithChainError(buffer.ChainError())
	}
	p.buffer = buffer
	return p
}

// PutSignals adds signals to buffer
// @TODO: rename
func (p *Port) PutSignals(signals ...*signal.Signal) *Port {
	if p.HasChainError() {
		return p
	}
	return p.withBuffer(p.Buffer().With(signals...))
}

// WithSignals puts buffer and returns the port
func (p *Port) WithSignals(signals ...*signal.Signal) *Port {
	if p.HasChainError() {
		return p
	}

	return p.PutSignals(signals...)
}

// WithSignalGroups puts groups of buffer and returns the port
func (p *Port) WithSignalGroups(signalGroups ...*signal.Group) *Port {
	if p.HasChainError() {
		return p
	}
	for _, group := range signalGroups {
		signals, err := group.Signals()
		if err != nil {
			return p.WithChainError(err)
		}
		p.PutSignals(signals...)
		if p.HasChainError() {
			return p
		}
	}

	return p
}

// Clear removes all signals from the port buffer
func (p *Port) Clear() *Port {
	if p.HasChainError() {
		return p
	}
	return p.withBuffer(signal.NewGroup())
}

// Flush pushes buffer to pipes and clears the port
// @TODO: hide this method from user
func (p *Port) Flush() *Port {
	if p.HasChainError() {
		return p
	}

	if !p.HasSignals() || !p.HasPipes() {
		//@TODO maybe better to return explicit errors
		return New("").WithChainError(ErrPortNotReadyForFlush)
	}

	pipes, err := p.pipes.Ports()
	if err != nil {
		return p.WithChainError(err)
	}

	for _, outboundPort := range pipes {
		//Fan-Out
		err = ForwardSignals(p, outboundPort)
		if err != nil {
			return p.WithChainError(err)
		}
	}
	return p.Clear()
}

// HasSignals says whether port buffer is set or not
func (p *Port) HasSignals() bool {
	if p.HasChainError() {
		//@TODO: add logging here
		return false
	}
	signals, err := p.AllSignals()
	if err != nil {
		//@TODO: add logging here
		return false
	}
	return len(signals) > 0
}

// HasPipes says whether port has outbound pipes
func (p *Port) HasPipes() bool {
	if p.HasChainError() {
		//@TODO: add logging here
		return false
	}
	pipes, err := p.pipes.Ports()
	if err != nil {
		//@TODO: add logging here
		return false
	}

	return len(pipes) > 0
}

// PipeTo creates one or multiple pipes to other port(s)
// @TODO: hide this method from AF
func (p *Port) PipeTo(destPorts ...*Port) *Port {
	if p.HasChainError() {
		return p
	}
	for _, destPort := range destPorts {
		if destPort == nil {
			continue
		}
		p.pipes = p.pipes.With(destPort)
	}
	return p
}

// WithLabels sets labels and returns the port
func (p *Port) WithLabels(labels common.LabelsCollection) *Port {
	if p.HasChainError() {
		return p
	}

	p.LabeledEntity.SetLabels(labels)
	return p
}

// ForwardSignals copies all buffer from source port to destination port, without clearing the source port
func ForwardSignals(source *Port, dest *Port) error {
	signals, err := source.AllSignals()
	if err != nil {
		return err
	}
	dest.PutSignals(signals...)
	if dest.HasChainError() {
		return dest.ChainError()
	}
	return nil
}

// WithChainError returns port with error
func (p *Port) WithChainError(err error) *Port {
	p.SetChainError(err)
	return p
}

// FirstSignalPayload is shortcut method
func (p *Port) FirstSignalPayload() (any, error) {
	return p.Buffer().FirstPayload()
}

// FirstSignalPayloadOrNil is shortcut method
func (p *Port) FirstSignalPayloadOrNil() any {
	return p.Buffer().First().PayloadOrNil()
}

// FirstSignalPayloadOrDefault is shortcut method
func (p *Port) FirstSignalPayloadOrDefault(defaultPayload any) any {
	return p.Buffer().First().PayloadOrDefault(defaultPayload)
}

// AllSignals is shortcut method
func (p *Port) AllSignals() (signal.Signals, error) {
	return p.Buffer().Signals()
}

// AllSignalsOrNil is shortcut method
func (p *Port) AllSignalsOrNil() signal.Signals {
	return p.Buffer().SignalsOrNil()
}

func (p *Port) AllSignalsOrDefault(defaultSignals signal.Signals) signal.Signals {
	return p.Buffer().SignalsOrDefault(defaultSignals)
}

// AllSignalsPayloads is shortcut method
func (p *Port) AllSignalsPayloads() ([]any, error) {
	return p.Buffer().AllPayloads()
}
