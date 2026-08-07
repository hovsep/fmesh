package port

import (
	"context"
	"errors"
	"fmt"

	"github.com/hovsep/fmesh/meta"
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
	labels          *meta.Labels
	scalars         *meta.Scalars
	signals         *signal.Group
	pipes           *Group // Outbound pipes
	parentComponent ParentComponent
	hooks           *Hooks
}

// NewInput creates a new input port, applying any provided options.
func NewInput(name string, opts ...Option) (*Port, error) {
	return newPort(DirectionIn, name, opts...)
}

// NewOutput creates a new output port, applying any provided options.
func NewOutput(name string, opts ...Option) (*Port, error) {
	return newPort(DirectionOut, name, opts...)
}

func newPort(direction Direction, name string, opts ...Option) (*Port, error) {
	p := &Port{
		name:      name,
		direction: direction,
		labels:    meta.NewLabels(),
		scalars:   meta.NewScalars(),
		pipes:     NewGroup(),
		signals:   signal.NewGroup(),
		hooks:     newHooks(),
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
		p.labels.Set(name, value)
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
func (p *Port) Labels() *meta.Labels {
	return p.labels
}

// Scalars returns the port's scalars store.
func (p *Port) Scalars() *meta.Scalars {
	return p.scalars
}

// WithScalar is a port constructor option that adds or updates a single scalar.
func WithScalar(name string, value float64) Option {
	return func(p *Port) error {
		p.scalars.Set(name, value)
		return nil
	}
}

// Pipes returns outbound pipes. Input ports always return an empty group.
func (p *Port) Pipes() *Group {
	return p.pipes
}

func (p *Port) setSignals(signalsGroup *signal.Group) {
	p.signals = signalsGroup
}

// Signals returns all signals in the port.
func (p *Port) Signals() *signal.Group {
	return p.signals
}

// PutSignals adds signals to the port.
// When the OnSignalsAdded hook fails, the port is restored to its previous state.
//
// Seeding a mesh happens before there is a run to cancel, so this takes no
// context and the OnSignalsAdded hook receives context.Background(). Delivery
// during a run goes through Flush, which passes the run context along.
func (p *Port) PutSignals(signals ...*signal.Signal) error {
	return p.putSignals(context.Background(), signals)
}

// PutPayloads creates signals from given payloads.
// When the OnSignalsAdded hook fails, the port is restored to its previous state.
func (p *Port) PutPayloads(payloads ...any) error {
	return p.putSignals(context.Background(), signal.NewGroup(payloads...).All())
}

// putSignals adds signals to the port and triggers the OnSignalsAdded hook,
// rolling the port back when the hook fails.
func (p *Port) putSignals(ctx context.Context, signals []*signal.Signal) error {
	previousSignals := p.Signals()
	p.setSignals(previousSignals.With(signals...))

	if err := p.hooks.onSignalsAdded.Trigger(ctx, &SignalsAddedContext{
		Port:         p,
		SignalsAdded: signals,
	}); err != nil {
		p.setSignals(previousSignals)
		return fmt.Errorf("onSignalsAdded hook failed: %w", err)
	}

	return nil
}

// PutSignalGroups adds all signals from signal groups.
func (p *Port) PutSignalGroups(signalGroups ...*signal.Group) error {
	for _, group := range signalGroups {
		signals := group.All()
		if err := p.PutSignals(signals...); err != nil {
			return err
		}
	}
	return nil
}

// Clear removes all signals.
func (p *Port) Clear(ctx context.Context) error {
	signalsCleared := p.Signals().Len()
	p.setSignals(signal.NewGroup())

	// Trigger OnClear hook
	if err := p.hooks.onClear.Trigger(ctx, &ClearContext{
		Port:           p,
		SignalsCleared: signalsCleared,
	}); err != nil {
		return fmt.Errorf("onClear hook failed: %w", err)
	}

	return nil
}

// Flush pushes signals to pipes and clears the port (output ports only).
// Fan-out forwards the same *Signal pointers to every destination — payloads
// must be treated as immutable by all receivers.
// Delivery to every pipe is attempted even when some fail; the errors are
// joined and the port is cleared only when all deliveries succeeded, so
// successfully delivered signals are never lost, but a retry after a partial
// failure re-delivers to the destinations that already received them.
// Each successful delivery fires the OnSignalsDelivered hook on this port.
func (p *Port) Flush(ctx context.Context) error {
	if p.IsInput() {
		return fmt.Errorf("cannot flush input port %q: only output ports can be flushed", p.Name())
	}

	if !p.HasSignals() || !p.HasPipes() {
		return nil
	}

	// Materialized once for the whole fan-out rather than per destination.
	signals := p.Signals().All()
	// The hook context escapes to the heap whether or not anything reads it, so
	// the common case of no hook must not reach the composite literal.
	notifyDelivery := !p.hooks.onSignalsDelivered.IsEmpty()

	var deliveryErrs error
	// ForEach avoids cloning the pipe slice on every flush (hot path)
	_ = p.pipes.ForEach(func(outboundPort *Port) error {
		// Fan-Out
		if err := outboundPort.putSignals(ctx, signals); err != nil {
			deliveryErrs = errors.Join(deliveryErrs, err)
			return nil
		}

		if notifyDelivery {
			if err := p.hooks.onSignalsDelivered.Trigger(ctx, &SignalsDeliveredContext{
				SourcePort:       p,
				DestinationPort:  outboundPort,
				SignalsDelivered: signals,
			}); err != nil {
				deliveryErrs = errors.Join(deliveryErrs, fmt.Errorf("onSignalsDelivered hook failed: %w", err))
			}
		}
		return nil
	})
	if deliveryErrs != nil {
		return deliveryErrs
	}
	return p.Clear(ctx)
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
// Flushing fans out the same *Signal pointers to every destination, so
// payloads must be treated as immutable by all receiving components.
func (p *Port) PipeTo(destPorts ...*Port) error {
	for _, destPort := range destPorts {
		if err := validatePipe(p, destPort); err != nil {
			return fmt.Errorf("pipe validation failed: %w", err)
		}
		p.pipes.add(destPort)

		// Wiring happens before there is a run to cancel, so the pipe hooks get
		// context.Background() like every other construction-time hook.
		if err := p.hooks.onOutboundPipe.Trigger(context.Background(), &OutboundPipeContext{
			SourcePort:      p,
			DestinationPort: destPort,
		}); err != nil {
			return fmt.Errorf("onOutboundPipe hook failed: %w", err)
		}

		// Trigger OnInboundPipe hook on destination port
		if err := destPort.hooks.onInboundPipe.Trigger(context.Background(), &InboundPipeContext{
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
// The same *Signal pointers are shared with the destination; payloads must be
// treated as immutable because downstream components activate concurrently.
func ForwardSignals(ctx context.Context, source, dest *Port) error {
	signals := source.Signals().All()
	return dest.putSignals(ctx, signals)
}

// ForwardWithFilter copies signals that pass filter function from source to dest port.
func ForwardWithFilter(ctx context.Context, source, dest *Port, p signal.Predicate) error {
	filteredSignals := source.Signals().Filter(p).All()
	return dest.putSignals(ctx, filteredSignals)
}

// ForwardWithMap applies mapperFunc to each signal and copies it to the dest port.
func ForwardWithMap(ctx context.Context, source, dest *Port, mapperFunc signal.Mapper) error {
	mappedSignals := source.Signals().Map(mapperFunc).All()
	return dest.putSignals(ctx, mappedSignals)
}

// ParentComponent returns the port's parent component.
func (p *Port) ParentComponent() ParentComponent {
	return p.parentComponent
}

func (p *Port) setParentComponent(parentComponent ParentComponent) {
	p.parentComponent = parentComponent
}

// SetupHooks configures port hooks using a closure.
func (p *Port) SetupHooks(configure func(*Hooks)) *Port {
	configure(p.hooks)
	return p
}
