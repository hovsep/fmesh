package port

import (
	"fmt"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/signal"
)

const (
	DirectionLabel = "fmesh:port:direction"
	DirectionIn    = "in"
	DirectionOut   = "out"
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
		return signal.NewGroup().WithChainError(p.ChainError())
	}
	return p.buffer
}

// Pipes getter
// @TODO maybe better to return []*Port directly
func (p *Port) Pipes() *Group {
	if p.HasChainError() {
		return NewGroup().WithChainError(p.ChainError())
	}
	return p.pipes
}

// withBuffer sets buffer field
func (p *Port) withBuffer(buffer *signal.Group) *Port {
	if p.HasChainError() {
		return p
	}

	if buffer.HasChainError() {
		p.SetChainError(buffer.ChainError())
		return New("").WithChainError(p.ChainError())
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
			p.SetChainError(err)
			return New("").WithChainError(p.ChainError())
		}
		p.PutSignals(signals...)
		if p.HasChainError() {
			return New("").WithChainError(p.ChainError())
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
		//Log,this
		//Nothing to flush
		return p
	}

	pipes, err := p.pipes.Ports()
	if err != nil {
		p.SetChainError(err)
		return New("").WithChainError(p.ChainError())
	}

	for _, outboundPort := range pipes {
		//Fan-Out
		err = ForwardSignals(p, outboundPort)
		if err != nil {
			p.SetChainError(err)
			return New("").WithChainError(p.ChainError())
		}
	}
	return p.Clear()
}

// HasSignals says whether port buffer is set or not
func (p *Port) HasSignals() bool {
	return len(p.AllSignalsOrNil()) > 0
}

// HasPipes says whether port has outbound pipes
func (p *Port) HasPipes() bool {
	return p.Pipes().Len() > 0
}

// PipeTo creates one or multiple pipes to other port(s)
// @TODO: hide this method from AF
func (p *Port) PipeTo(destPorts ...*Port) *Port {
	if p.HasChainError() {
		return p
	}

	for _, destPort := range destPorts {
		if err := validatePipe(p, destPort); err != nil {
			p.SetChainError(fmt.Errorf("pipe validation failed: %w", err))
			return New("").WithChainError(p.ChainError())
		}
		p.pipes = p.pipes.With(destPort)
	}
	return p
}

func validatePipe(srcPort *Port, dstPort *Port) error {
	if srcPort == nil || dstPort == nil {
		return ErrNilPort
	}

	srcDir, dstDir := srcPort.LabelOrDefault(DirectionLabel, ""), dstPort.LabelOrDefault(DirectionLabel, "")

	if srcDir == "" || dstDir == "" {
		return ErrMissingLabel
	}

	if srcDir == "in" || dstDir == "out" {
		return ErrInvalidPipeDirection
	}

	return nil
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
