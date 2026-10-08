package port

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/signal"
)

// emptyGroup is shared by every cleared port: groups are copy-on-write.
var emptyGroup = signal.NewGroup()

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
	meta            *meta.Meta
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
		meta:      meta.New(),
		pipes:     newGroup(),
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

// WithMeta is a port constructor option that adds or updates one metadata entry.
func WithMeta[T meta.Value](key string, value T) Option {
	return func(p *Port) error {
		p.meta.Set(key, value)
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

// Meta returns the port's metadata store.
func (p *Port) Meta() *meta.Meta {
	return p.meta
}

// Pipes returns outbound pipes. Input ports always return an empty group.
func (p *Port) Pipes() *Group {
	p.mustExist()
	return p.pipes
}

func (p *Port) setSignals(signalsGroup *signal.Group) {
	p.signals = signalsGroup
}

// mustExist panics with ErrNilPort on a nil receiver, naming the usual cause (a
// lookup by an unknown name) instead of a nil dereference frames later.
func (p *Port) mustExist() {
	if p == nil {
		panic(ErrNilPort)
	}
}

// Signals returns all signals in the port.
func (p *Port) Signals() *signal.Group {
	p.mustExist()
	return p.signals
}

// PutSignals adds signals to the port.
// When the OnSignalsAdded hook fails, the port is restored to its previous state.
//
// Seeding a mesh happens before there is a run to cancel, so this takes no
// context and the OnSignalsAdded hook receives context.Background(). Delivery
// during a run goes through Flush, which passes the run context along.
func (p *Port) PutSignals(signals ...*signal.Signal) error {
	p.mustExist()
	return p.putSignals(context.Background(), signals)
}

// PutPayloads creates signals from given payloads.
// When the OnSignalsAdded hook fails, the port is restored to its previous state.
func (p *Port) PutPayloads(payloads ...any) error {
	p.mustExist()
	return p.putSignals(context.Background(), signal.NewGroup(payloads...).All())
}

// putSignals adds signals to the port and triggers the OnSignalsAdded hook,
// rolling the port back when the hook fails.
func (p *Port) putSignals(ctx context.Context, signals []*signal.Signal) error {
	previousSignals := p.Signals()
	p.setSignals(previousSignals.With(signals...))

	// The hook context escapes to the heap whether or not anything reads it, so
	// the common case of no hook must not reach the composite literal (this runs
	// once per pipe delivery, per cycle).
	if p.hooks.onSignalsAdded.IsEmpty() {
		return nil
	}

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
	// Guarded here as well as in PutSignals: with no groups the loop never runs,
	// and a nil port would report success.
	p.mustExist()

	for _, group := range signalGroups {
		if err := p.PutSignals(group.All()...); err != nil {
			return err
		}
	}
	return nil
}

// Clear removes all signals.
func (p *Port) Clear(ctx context.Context) error {
	signalsCleared := p.Signals().Len()
	p.setSignals(emptyGroup)

	// Same allocation guard as putSignals: this runs for every port on every
	// cycle, and the context struct escapes even with no hooks registered.
	if p.hooks.onClear.IsEmpty() {
		return nil
	}

	if err := p.hooks.onClear.TriggerAll(ctx, &ClearContext{
		Port:           p,
		SignalsCleared: signalsCleared,
	}); err != nil {
		return fmt.Errorf("onClear hook failed: %w", err)
	}

	return nil
}

// Flush delivers the port's signals through every pipe, then clears it (output
// ports only). Every destination receives the same *Signal pointers.
//
// Every pipe is attempted; on any failure the errors are joined and the port is
// not cleared, so a retry re-delivers to the destinations that succeeded.
func (p *Port) Flush(ctx context.Context) error {
	p.mustExist()
	if p.IsInput() {
		return fmt.Errorf("cannot flush input port %q: only output ports can be flushed", p.Name())
	}

	if !p.HasSignals() || !p.HasPipes() {
		return nil
	}

	signals := p.Signals().All()
	// The hook context escapes to the heap whether or not anything reads it, so
	// the common case of no hook must not reach the composite literal.
	notifyDelivery := !p.hooks.onSignalsDelivered.IsEmpty()

	var deliveryErrs error
	for _, outboundPort := range p.pipes.raw() {
		if err := outboundPort.putSignals(ctx, signals); err != nil {
			deliveryErrs = errors.Join(deliveryErrs, err)
			continue
		}

		if notifyDelivery {
			if err := p.hooks.onSignalsDelivered.TriggerAll(ctx, &SignalsDeliveredContext{
				SourcePort:       p,
				DestinationPort:  outboundPort,
				SignalsDelivered: signals,
			}); err != nil {
				deliveryErrs = errors.Join(deliveryErrs, fmt.Errorf("onSignalsDelivered hook failed: %w", err))
			}
		}
	}
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
// A failing pipe hook removes that pipe; pipes made earlier in the call stay.
func (p *Port) PipeTo(destPorts ...*Port) error {
	for _, destPort := range destPorts {
		if err := validatePipe(p, destPort); err != nil {
			return fmt.Errorf("pipe validation failed: %w", err)
		}
		index := p.pipes.Len()
		p.pipes.add(destPort)

		if err := p.triggerPipeHooks(destPort); err != nil {
			p.pipes.setPorts(slices.Delete(p.pipes.raw(), index, index+1))
			return err
		}
	}
	return nil
}

// triggerPipeHooks fires OnOutboundPipe on this port, then OnInboundPipe on the
// destination. Wiring happens before there is a run to cancel, so the pipe hooks
// get context.Background() like every other construction-time hook.
func (p *Port) triggerPipeHooks(destPort *Port) error {
	if err := p.hooks.onOutboundPipe.Trigger(context.Background(), &OutboundPipeContext{
		SourcePort:      p,
		DestinationPort: destPort,
	}); err != nil {
		return fmt.Errorf("onOutboundPipe hook failed: %w", err)
	}
	if err := destPort.hooks.onInboundPipe.Trigger(context.Background(), &InboundPipeContext{
		DestinationPort: destPort,
		SourcePort:      p,
	}); err != nil {
		return fmt.Errorf("onInboundPipe hook failed: %w", err)
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
	return dest.putSignals(ctx, source.Signals().All())
}

// ForwardWithFilter copies signals that pass filter function from source to dest port.
func ForwardWithFilter(ctx context.Context, source, dest *Port, p signal.Predicate) error {
	return dest.putSignals(ctx, source.Signals().Filter(p).All())
}

// ForwardWithMap applies mapperFunc to each signal and copies it to the dest port.
func ForwardWithMap(ctx context.Context, source, dest *Port, mapperFunc signal.Mapper) error {
	return dest.putSignals(ctx, source.Signals().Map(mapperFunc).All())
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
