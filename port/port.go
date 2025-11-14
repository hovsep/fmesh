package port

import (
	"errors"
	"fmt"

	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/signal"
)

// Direction represents the direction of a port (input or output).
// It's a boolean type where true = input, false = output.
type Direction bool

const (
	// DirectionIn is the direction for input ports.
	DirectionIn Direction = true
	// DirectionOut is the direction for output ports.
	DirectionOut Direction = false
)

// Port defines a connectivity point of a component.
type Port struct {
	name            string
	direction       Direction // Input or output direction
	description     string
	labels          *labels.Collection
	chainableErr    error
	signals         *signal.Group
	pipes           *Group // Outbound pipes
	parentComponent ParentComponent
	hooks           *Hooks
}

// NewInput creates a new input port.
func NewInput(name string) *Port {
	return &Port{
		name:         name,
		direction:    DirectionIn,
		labels:       labels.NewCollection(nil),
		chainableErr: nil,
		pipes:        NewGroup(),
		signals:      signal.NewGroup(),
		hooks:        NewHooks(),
	}
}

// NewOutput creates a new output port.
func NewOutput(name string) *Port {
	return &Port{
		name:         name,
		direction:    DirectionOut,
		labels:       labels.NewCollection(nil),
		chainableErr: nil,
		pipes:        NewGroup(),
		signals:      signal.NewGroup(),
		hooks:        NewHooks(),
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
	p.labels.Clear().AddMany(labelMap)
	return p
}

// AddLabels adds or updates labels and returns the port for chaining.
func (p *Port) AddLabels(labelMap labels.Map) *Port {
	if p.HasChainableErr() {
		return p
	}
	p.labels.AddMany(labelMap)
	return p
}

// AddLabel adds or updates a single label and returns the port for chaining.
func (p *Port) AddLabel(name, value string) *Port {
	if p.HasChainableErr() {
		return p
	}
	p.labels.Add(name, value)
	return p
}

// ClearLabels removes all labels and returns the port for chaining.
func (p *Port) ClearLabels() *Port {
	if p.HasChainableErr() {
		return p
	}
	p.labels.Clear()
	return p
}

// WithoutLabels removes specific labels and returns the port for chaining.
func (p *Port) WithoutLabels(names ...string) *Port {
	if p.HasChainableErr() {
		return p
	}
	p.labels.Without(names...)
	return p
}

// WithDescription sets the port description and returns the port for chaining.
func (p *Port) WithDescription(description string) *Port {
	if p.HasChainableErr() {
		return p
	}

	p.description = description
	return p
}

// Pipes returns outbound pipes (output ports only).
func (p *Port) Pipes() *Group {
	if p.HasChainableErr() {
		return NewGroup().WithChainableErr(p.ChainableErr())
	}

	if p.IsInput() {
		return NewGroup().WithChainableErr(fmt.Errorf("port '%s' is an input port and cannot have outbound pipes", p.Name()))
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
		return NewOutput("").WithChainableErr(p.ChainableErr())
	}
	p.signals = signalsGroup
	return p
}

// Signals returns all signals in the port.
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

	result := p.withSignals(p.Signals().Add(signals...))

	// Trigger OnSignalsAdded hook
	p.hooks.onSignalsAdded.Trigger(&PutContext{
		Port:         p,
		SignalsAdded: signals,
	})

	return result
}

// PutSignalGroups adds all signals from signal groups and returns the port for chaining.
func (p *Port) PutSignalGroups(signalGroups ...*signal.Group) *Port {
	if p.HasChainableErr() {
		return p
	}
	for _, group := range signalGroups {
		signals, err := group.All()
		if err != nil {
			p.WithChainableErr(err)
			return NewOutput("").WithChainableErr(p.ChainableErr())
		}
		p.PutSignals(signals...)
		if p.HasChainableErr() {
			return NewOutput("").WithChainableErr(p.ChainableErr())
		}
	}

	return p
}

// Clear removes all signals and returns the port for chaining.
func (p *Port) Clear() *Port {
	if p.HasChainableErr() {
		return p
	}

	signalsCleared := p.Signals().Len()
	result := p.withSignals(signal.NewGroup())

	// Trigger OnClear hook
	p.hooks.onClear.Trigger(&ClearContext{
		Port:           p,
		SignalsCleared: signalsCleared,
	})

	return result
}

// Flush pushes signals to pipes and clears the port (output ports only).
func (p *Port) Flush() *Port {
	if p.HasChainableErr() {
		return p
	}

	if p.IsInput() {
		p.WithChainableErr(fmt.Errorf("cannot flush input port '%s': only output ports can be flushed", p.Name()))
		return NewOutput("").WithChainableErr(p.ChainableErr())
	}

	if !p.HasSignals() || !p.HasPipes() {
		// Log this
		// Nothing to flush
		return p
	}

	pipes, err := p.pipes.All()
	if err != nil {
		p.WithChainableErr(err)
		return NewOutput("").WithChainableErr(p.ChainableErr())
	}

	for _, outboundPort := range pipes {
		// Fan-Out
		err = ForwardSignals(p, outboundPort)
		if err != nil {
			p.WithChainableErr(err)
			return NewOutput("").WithChainableErr(p.ChainableErr())
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

// PipeTo connects this port to destination ports and returns the port for chaining.
func (p *Port) PipeTo(destPorts ...*Port) *Port {
	if p.HasChainableErr() {
		return p
	}

	for _, destPort := range destPorts {
		if err := validatePipe(p, destPort); err != nil {
			p.WithChainableErr(fmt.Errorf("pipe validation failed: %w", err))
			return NewOutput("").WithChainableErr(p.ChainableErr())
		}
		p.pipes = p.pipes.Add(destPort)

		// Trigger OnOutboundPipe hook on source port (this port)
		p.hooks.onOutboundPipe.Trigger(&OutboundPipeContext{
			SourcePort:      p,
			DestinationPort: destPort,
		})

		// Trigger OnInboundPipe hook on destination port
		destPort.hooks.onInboundPipe.Trigger(&InboundPipeContext{
			DestinationPort: destPort,
			SourcePort:      p,
		})
	}
	return p
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
// The error is automatically joined with the port's name as context.
func (p *Port) WithChainableErr(err error) *Port {
	if err == nil {
		p.chainableErr = nil
		return p
	}

	contextErr := fmt.Errorf("error in port '%s'", p.Name())
	p.chainableErr = errors.Join(contextErr, err)
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

// SetupHooks configures port hooks using a closure and returns the port for chaining.
func (p *Port) SetupHooks(configure func(*Hooks)) *Port {
	if p.HasChainableErr() {
		return p
	}
	configure(p.hooks)
	return p
}
