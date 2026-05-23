package port

import (
	"fmt"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/signal"
)

// Direction represents the direction of a port.
type Direction int

const (
	// DirectionUndefined is the zero value; a port with this direction is misconfigured.
	DirectionUndefined Direction = iota
	// DirectionIn is the direction for input ports.
	DirectionIn
	// DirectionOut is the direction for output ports.
	DirectionOut
)

// Option is a functional option for configuring a port during construction.
type Option func(*Port) error

// Port defines a connectivity point of a component.
type Port struct {
	name            string
	direction       Direction
	description     string
	labels          *labels.Collection
	signals         *signal.Group
	pipes           *Group // Outbound pipes
	parentComponent ParentComponent
	hooks           *Hooks
}

// NewInput creates a new input port, applying any provided options.
func NewInput(name string, opts ...Option) (*Port, error) {
	p := &Port{
		name:      name,
		direction: DirectionIn,
		labels:    labels.NewCollection(),
		pipes:     NewGroup(),
		signals:   signal.NewGroup(),
		hooks:     NewHooks(),
	}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, fmt.Errorf("port %q option failed: %w", name, err)
		}
	}
	return p, nil
}

// NewOutput creates a new output port, applying any provided options.
func NewOutput(name string, opts ...Option) (*Port, error) {
	p := &Port{
		name:      name,
		direction: DirectionOut,
		labels:    labels.NewCollection(),
		pipes:     NewGroup(),
		signals:   signal.NewGroup(),
		hooks:     NewHooks(),
	}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, fmt.Errorf("port %q option failed: %w", name, err)
		}
	}
	return p, nil
}

// WithDescription is a port option that sets the description.
func WithDescription(description string) Option {
	return func(p *Port) error {
		p.description = description
		return nil
	}
}

// WithLabel is a port option that adds a label.
func WithLabel(name, value string) Option {
	return func(p *Port) error {
		p.labels.Add(name, value)
		return nil
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

// Direction returns the port's direction (input or output).
func (p *Port) Direction() Direction {
	return p.direction
}

// IsInput returns true if the port is an input port.
func (p *Port) IsInput() bool {
	return p.direction == DirectionIn
}

// IsOutput returns true if the port is an output port.
func (p *Port) IsOutput() bool {
	return p.direction == DirectionOut
}

// Labels returns the port's labels collection.
func (p *Port) Labels() *labels.Collection {
	return p.labels
}

// SetLabels replaces all labels.
func (p *Port) SetLabels(labelMap map[string]string) *Port {
	p.labels.Clear().AddMany(labelMap)
	return p
}

// AddLabels adds or updates labels.
func (p *Port) AddLabels(labelMap map[string]string) *Port {
	p.labels.AddMany(labelMap)
	return p
}

// AddLabel adds or updates a single label.
func (p *Port) AddLabel(name, value string) *Port {
	p.labels.Add(name, value)
	return p
}

// ClearLabels removes all labels.
func (p *Port) ClearLabels() *Port {
	p.labels.Clear()
	return p
}

// RemoveLabels removes specific labels.
func (p *Port) RemoveLabels(names ...string) *Port {
	p.labels.Remove(names...)
	return p
}

// Pipes returns outbound pipes. Input ports always return an empty group.
func (p *Port) Pipes() *Group {
	return p.pipes
}

// withSignals sets the signals field.
func (p *Port) withSignals(signalsGroup *signal.Group) {
	p.signals = signalsGroup
}

// Signals returns all signals in the port.
func (p *Port) Signals() *signal.Group {
	return p.signals
}

// PutSignals adds signals to the port.
func (p *Port) PutSignals(signals ...*signal.Signal) error {
	p.withSignals(p.Signals().With(signals...))

	// Trigger OnSignalsAdded hook
	if err := p.hooks.onSignalsAdded.Trigger(&SignalsAddedContext{
		Port:         p,
		SignalsAdded: signals,
	}); err != nil {
		return fmt.Errorf("onSignalsAdded hook failed: %w", err)
	}

	return nil
}

// PutPayloads creates signals from given payloads.
func (p *Port) PutPayloads(payloads ...any) error {
	newSignals, _ := signal.NewGroup(payloads...).All()
	p.withSignals(p.Signals().With(newSignals...))

	// Trigger OnSignalsAdded hook
	if err := p.hooks.onSignalsAdded.Trigger(&SignalsAddedContext{
		Port:         p,
		SignalsAdded: newSignals,
	}); err != nil {
		return fmt.Errorf("onSignalsAdded hook failed: %w", err)
	}

	return nil
}

// PutSignalGroups adds all signals from signal groups.
func (p *Port) PutSignalGroups(signalGroups ...*signal.Group) error {
	for _, group := range signalGroups {
		signals, err := group.All()
		if err != nil {
			return err
		}
		if err := p.PutSignals(signals...); err != nil {
			return err
		}
	}
	return nil
}

// Clear removes all signals.
func (p *Port) Clear() error {
	signalsCleared := p.Signals().Len()
	p.withSignals(signal.NewGroup())

	// Trigger OnClear hook
	if err := p.hooks.onClear.Trigger(&ClearContext{
		Port:           p,
		SignalsCleared: signalsCleared,
	}); err != nil {
		return fmt.Errorf("onClear hook failed: %w", err)
	}

	return nil
}

// Flush pushes signals to pipes and clears the port (output ports only).
func (p *Port) Flush() error {
	if p.IsInput() {
		return fmt.Errorf("cannot flush input port %q: only output ports can be flushed", p.Name())
	}

	if !p.HasSignals() || !p.HasPipes() {
		return nil
	}

	pipes, _ := p.pipes.All()
	for _, outboundPort := range pipes {
		// Fan-Out
		if err := ForwardSignals(p, outboundPort); err != nil {
			return err
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

// PipeTo connects this port to destination ports.
func (p *Port) PipeTo(destPorts ...*Port) error {
	for _, destPort := range destPorts {
		if err := validatePipe(p, destPort); err != nil {
			return fmt.Errorf("pipe validation failed: %w", err)
		}
		p.pipes.add(destPort)

		// Trigger OnOutboundPipe hook on source port (this port)
		if err := p.hooks.onOutboundPipe.Trigger(&OutboundPipeContext{
			SourcePort:      p,
			DestinationPort: destPort,
		}); err != nil {
			return fmt.Errorf("onOutboundPipe hook failed: %w", err)
		}

		// Trigger OnInboundPipe hook on destination port
		if err := destPort.hooks.onInboundPipe.Trigger(&InboundPipeContext{
			DestinationPort: destPort,
			SourcePort:      p,
		}); err != nil {
			return fmt.Errorf("onInboundPipe hook failed: %w", err)
		}
	}
	return nil
}

func validatePipe(srcPort, dstPort *Port) error {
	if srcPort == nil || dstPort == nil {
		return ErrNilPort
	}

	// Pipes must go from output to input
	if !srcPort.IsOutput() || !dstPort.IsInput() {
		return ErrInvalidPipeDirection
	}

	return nil
}

// ForwardSignals copies all signals from source to destination port without clearing source.
func ForwardSignals(source, dest *Port) error {
	signals, err := source.Signals().All()
	if err != nil {
		return err
	}
	return dest.PutSignals(signals...)
}

// ForwardWithFilter copies signals that pass filter function from source to dest port.
func ForwardWithFilter(source, dest *Port, p signal.Predicate) error {
	filteredSignals, err := source.Signals().Filter(p).All()
	if err != nil {
		return err
	}
	return dest.PutSignals(filteredSignals...)
}

// ForwardWithMap applies mapperFunc to each signal and copies it to the dest port.
func ForwardWithMap(source, dest *Port, mapperFunc signal.Mapper) error {
	mappedSignals, err := source.Signals().Map(mapperFunc).All()
	if err != nil {
		return err
	}
	return dest.PutSignals(mappedSignals...)
}

// ParentComponent returns the port's parent component.
func (p *Port) ParentComponent() ParentComponent {
	return p.parentComponent
}

// setParentComponent sets the parent component.
func (p *Port) setParentComponent(parentComponent ParentComponent) {
	p.parentComponent = parentComponent
}

// SetupHooks configures port hooks using a closure.
func (p *Port) SetupHooks(configure func(*Hooks)) *Port {
	configure(p.hooks)
	return p
}

// ValidateBeforeActivation checks if the port is valid before parent component activation.
func (p *Port) ValidateBeforeActivation() error {
	if p.ParentComponent() == nil {
		return fmt.Errorf("port %q has no parent component", p.name)
	}
	return nil
}
