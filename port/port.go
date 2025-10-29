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
	name            string
	description     string
	labels          *labels.Collection
	chainableErr    error
	signals         *signal.Group
	pipes           *Group // Outbound pipes
	parentComponent ParentComponent
}

// New creates a new port.
func New(name string) *Port {
	return &Port{
		name:         name,
		labels:       labels.NewCollection(nil),
		chainableErr: nil,
		pipes:        NewGroup(),
		signals:      signal.NewGroup(),
	}
}

// Name returns the port's name.
func (p *Port) Name() string {
	return p.name
}

// Description returns the port's description.
func (p *Port) Description() string {
	return p.description
}

// Labels returns the port's labels collection.
func (p *Port) Labels() *labels.Collection {
	if p.HasChainableErr() {
		return labels.NewCollection(nil).WithChainableErr(p.ChainableErr())
	}
	return p.labels
}

// SetLabels replaces all labels and returns the port for chaining.
func (p *Port) SetLabels(labelMap labels.Map) *Port {
	if p.HasChainableErr() {
		return p
	}
	p.labels.Clear().WithMany(labelMap)
	return p
}

// AddLabels adds or updates labels and returns the port for chaining.
func (p *Port) AddLabels(labelMap labels.Map) *Port {
	if p.HasChainableErr() {
		return p
	}
	p.labels.WithMany(labelMap)
	return p
}

// WithDescription sets a description.
func (p *Port) WithDescription(description string) *Port {
	if p.HasChainableErr() {
		return p
	}

	p.description = description
	return p
}

// Pipes returns the group of outbound pipes.
func (p *Port) Pipes() *Group {
	if p.HasChainableErr() {
		return NewGroup().WithChainableErr(p.ChainableErr())
	}
	return p.pipes
}

// withSignals sets the signals field.
func (p *Port) withSignals(signalsGroup *signal.Group) *Port {
	if p.HasChainableErr() {
		return p
	}

	if signalsGroup.HasChainableErr() {
		p.WithChainableErr(signalsGroup.ChainableErr())
		return New("").WithChainableErr(p.ChainableErr())
	}
	p.signals = signalsGroup
	return p
}

// Signals returns the signal group containing all signals in the port.
func (p *Port) Signals() *signal.Group {
	if p.HasChainableErr() {
		return signal.NewGroup().WithChainableErr(p.ChainableErr())
	}
	return p.signals
}

// PutSignals adds signals to the port and returns the port for chaining.
func (p *Port) PutSignals(signals ...*signal.Signal) *Port {
	if p.HasChainableErr() {
		return p
	}
	return p.withSignals(p.Signals().With(signals...))
}

// PutSignalGroups adds signals from multiple groups and returns the port for chaining.
func (p *Port) PutSignalGroups(signalGroups ...*signal.Group) *Port {
	if p.HasChainableErr() {
		return p
	}
	for _, group := range signalGroups {
		signals, err := group.All()
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

// Clear removes all signals from the port.
func (p *Port) Clear() *Port {
	if p.HasChainableErr() {
		return p
	}
	return p.withSignals(signal.NewGroup())
}

// Flush pushes signals to all pipes and clears the port.
func (p *Port) Flush() *Port {
	if p.HasChainableErr() {
		return p
	}

	if !p.HasSignals() || !p.HasPipes() {
		// Log this
		// Nothing to flush
		return p
	}

	pipes, err := p.pipes.All()
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

// HasSignals returns true if the port has any signals.
func (p *Port) HasSignals() bool {
	return !p.Signals().IsEmpty()
}

// HasPipes says whether a port has outbound pipes.
func (p *Port) HasPipes() bool {
	return !p.Pipes().IsEmpty()
}

// PipeTo creates one or multiple pipes to other port(s).
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

// ForwardSignals copies all signals from source port to destination port, without clearing the source port.
func ForwardSignals(source, dest *Port) error {
	if source.HasChainableErr() {
		return source.ChainableErr()
	}

	if dest.HasChainableErr() {
		return dest.ChainableErr()
	}

	signals, err := source.Signals().All()
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

	filteredSignals, err := source.Signals().Filter(p).All()
	if err != nil {
		return err
	}

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

	mappedSignals, err := source.Signals().Map(mapperFunc).All()
	if err != nil {
		return err
	}

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

// ChainableErr returns the chainable error.
func (p *Port) ChainableErr() error {
	return p.chainableErr
}

// ParentComponent returns the port's parent component.
func (p *Port) ParentComponent() ParentComponent {
	return p.parentComponent
}

// WithParentComponent sets the parent component.
func (p *Port) WithParentComponent(parentComponent ParentComponent) *Port {
	p.parentComponent = parentComponent
	return p
}
