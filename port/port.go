package port

import (
	"fmt"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/signal"
)

const (
	// DirectionLabel is the label for the port direction.
	DirectionLabel = "fmesh:port:direction"
	// DirectionIn is the direction for input ports.
	DirectionIn = "in"
	// DirectionOut is the direction for output ports.
	DirectionOut = "out"
)

// Port defines a connectivity point of a component.
type Port struct {
	name         string
	description  string
	labels       *labels.Collection
	chainableErr error
	buffer       *signal.Group
	pipes        *Group // Outbound pipes
}

// New creates a new port.
func New(name string) *Port {
	return &Port{
		name:         name,
		labels:       labels.NewCollection(nil),
		chainableErr: nil,
		pipes:        NewGroup(),
		buffer:       signal.NewGroup(),
	}
}

// Name getter.
func (p *Port) Name() string {
	return p.name
}

// Description getter.
func (p *Port) Description() string {
	return p.description
}

// Labels getter.
func (p *Port) Labels() *labels.Collection {
	if p.HasChainableErr() {
		return labels.NewCollection(nil).WithChainableErr(p.ChainableErr())
	}
	return p.labels
}

// WithDescription sets a description.
func (p *Port) WithDescription(description string) *Port {
	if p.HasChainableErr() {
		return p
	}

	p.description = description
	return p
}

// Buffer getter
// @TODO: maybe we can hide this and return signals to user code.
func (p *Port) Buffer() *signal.Group {
	if p.HasChainableErr() {
		return signal.NewGroup().WithChainableErr(p.ChainableErr())
	}
	return p.buffer
}

// Pipes getter
// @TODO maybe better to return []*Port directly.
func (p *Port) Pipes() *Group {
	if p.HasChainableErr() {
		return NewGroup().WithChainableErr(p.ChainableErr())
	}
	return p.pipes
}

// withBuffer sets buffer field.
func (p *Port) withBuffer(buffer *signal.Group) *Port {
	if p.HasChainableErr() {
		return p
	}

	if buffer.HasChainableErr() {
		p.WithChainableErr(buffer.ChainableErr())
		return New("").WithChainableErr(p.ChainableErr())
	}
	p.buffer = buffer
	return p
}

// PutSignals adds signals to buffer.
func (p *Port) PutSignals(signals ...*signal.Signal) *Port {
	if p.HasChainableErr() {
		return p
	}
	return p.withBuffer(p.Buffer().With(signals...))
}

// WithSignals appends signals into the buffer and returns the port.
func (p *Port) WithSignals(signals ...*signal.Signal) *Port {
	if p.HasChainableErr() {
		return p
	}

	return p.PutSignals(signals...)
}

// WithSignalGroups puts groups of buffer and returns the port.
func (p *Port) WithSignalGroups(signalGroups ...*signal.Group) *Port {
	if p.HasChainableErr() {
		return p
	}
	for _, group := range signalGroups {
		signals, err := group.Signals()
		if err != nil {
			p.WithChainableErr(err)
			return New("").WithChainableErr(p.ChainableErr())
		}
		p.PutSignals(signals...)
		if p.HasChainableErr() {
			return New("").WithChainableErr(p.ChainableErr())
		}
	}

	return p
}

// Clear removes all signals from the port buffer.
func (p *Port) Clear() *Port {
	if p.HasChainableErr() {
		return p
	}
	return p.withBuffer(signal.NewGroup())
}

// Flush pushes buffer to pipes and clears the port
// @TODO: hide this method from user.
func (p *Port) Flush() *Port {
	if p.HasChainableErr() {
		return p
	}

	if !p.HasSignals() || !p.HasPipes() {
		// Log this
		// Nothing to flush
		return p
	}

	pipes, err := p.pipes.Ports()
	if err != nil {
		p.WithChainableErr(err)
		return New("").WithChainableErr(p.ChainableErr())
	}

	for _, outboundPort := range pipes {
		// Fan-Out
		err = ForwardSignals(p, outboundPort)
		if err != nil {
			p.WithChainableErr(err)
			return New("").WithChainableErr(p.ChainableErr())
		}
	}
	return p.Clear()
}

// HasSignals says whether port buffer is set or not.
func (p *Port) HasSignals() bool {
	return len(p.AllSignalsOrNil()) > 0
}

// HasPipes says whether a port has outbound pipes.
func (p *Port) HasPipes() bool {
	return p.Pipes().Len() > 0
}

// PipeTo creates one or multiple pipes to other port(s)
// @TODO: hide this method from AF.
func (p *Port) PipeTo(destPorts ...*Port) *Port {
	if p.HasChainableErr() {
		return p
	}

	for _, destPort := range destPorts {
		if err := validatePipe(p, destPort); err != nil {
			p.WithChainableErr(fmt.Errorf("pipe validation failed: %w", err))
			return New("").WithChainableErr(p.ChainableErr())
		}
		p.pipes = p.pipes.With(destPort)
	}
	return p
}

func validatePipe(srcPort, dstPort *Port) error {
	if srcPort == nil || dstPort == nil {
		return ErrNilPort
	}

	srcDir, dstDir := srcPort.labels.ValueOrDefault(DirectionLabel, ""), dstPort.labels.ValueOrDefault(DirectionLabel, "")

	if srcDir == "" || dstDir == "" {
		return ErrMissingLabel
	}

	if srcDir == "in" || dstDir == "out" {
		return ErrInvalidPipeDirection
	}

	return nil
}

// WithLabels sets labels and returns the port.
func (p *Port) WithLabels(labels labels.Map) *Port {
	if p.HasChainableErr() {
		return p
	}

	p.labels.WithMany(labels)
	return p
}

// ForwardSignals copies all signals from source port to destination port, without clearing the source port.
func ForwardSignals(source, dest *Port) error {
	if source.HasChainableErr() {
		return source.ChainableErr()
	}

	if dest.HasChainableErr() {
		return dest.ChainableErr()
	}

	signals, err := source.AllSignals()
	if err != nil {
		return err
	}
	dest.PutSignals(signals...)
	if dest.HasChainableErr() {
		return dest.ChainableErr()
	}
	return nil
}

// ForwardWithFilter copies signals that pass filter function from source to dest port.
func ForwardWithFilter(source, dest *Port, p signal.Predicate) error {
	if source.HasChainableErr() {
		return source.ChainableErr()
	}

	if dest.HasChainableErr() {
		return dest.ChainableErr()
	}

	filteredSignals := source.Buffer().Filter(p).SignalsOrNil()

	dest.PutSignals(filteredSignals...)
	if dest.HasChainableErr() {
		return dest.ChainableErr()
	}
	return nil
}

// ForwardWithMap applies mapperFunc to each signal and copies it to the dest port.
func ForwardWithMap(source, dest *Port, mapperFunc signal.Mapper) error {
	if source.HasChainableErr() {
		return source.ChainableErr()
	}

	if dest.HasChainableErr() {
		return dest.ChainableErr()
	}

	mappedSignals := source.Buffer().Map(mapperFunc).SignalsOrNil()

	dest.PutSignals(mappedSignals...)
	if dest.HasChainableErr() {
		return dest.ChainableErr()
	}
	return nil
}

// WithChainableErr sets a chainable error and returns the port.
func (p *Port) WithChainableErr(err error) *Port {
	p.chainableErr = err
	return p
}

// HasChainableErr returns true when a chainable error is set.
func (p *Port) HasChainableErr() bool {
	return p.chainableErr != nil
}

// ChainableErr returns chainable error.
func (p *Port) ChainableErr() error {
	return p.chainableErr
}

// FirstSignalPayload is a shortcut method.
func (p *Port) FirstSignalPayload() (any, error) {
	return p.Buffer().FirstPayload()
}

// FirstSignalPayloadOrNil is a shortcut method.
func (p *Port) FirstSignalPayloadOrNil() any {
	return p.Buffer().FirstSignal().PayloadOrNil()
}

// FirstSignalPayloadOrDefault is a shortcut method.
func (p *Port) FirstSignalPayloadOrDefault(defaultPayload any) any {
	return p.Buffer().FirstSignal().PayloadOrDefault(defaultPayload)
}

// AllSignals is shortcut method.
func (p *Port) AllSignals() (signal.Signals, error) {
	return p.Buffer().Signals()
}

// AllSignalsOrNil is a shortcut method.
func (p *Port) AllSignalsOrNil() signal.Signals {
	return p.Buffer().SignalsOrNil()
}

// AllSignalsOrDefault is a shortcut method.
func (p *Port) AllSignalsOrDefault(defaultSignals signal.Signals) signal.Signals {
	return p.Buffer().SignalsOrDefault(defaultSignals)
}

// AllSignalsPayloads is a shortcut method.
func (p *Port) AllSignalsPayloads() ([]any, error) {
	return p.Buffer().AllPayloads()
}
