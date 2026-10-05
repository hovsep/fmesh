package port

import (
	"context"

	"github.com/hovsep/fmesh/internal/hook"
	"github.com/hovsep/fmesh/signal"
)

// SignalsAddedContext provides context when signals are added to a port.
type SignalsAddedContext struct {
	Port         *Port
	SignalsAdded []*signal.Signal
}

// ClearContext provides context when signals are cleared from a port.
type ClearContext struct {
	Port           *Port
	SignalsCleared int
}

// SignalsDeliveredContext provides context when signals are delivered from this
// port through one pipe. One event per pipe, not one per flush.
type SignalsDeliveredContext struct {
	SourcePort       *Port
	DestinationPort  *Port
	SignalsDelivered []*signal.Signal
}

// InboundPipeContext provides context when a pipe is created TO this port.
type InboundPipeContext struct {
	DestinationPort *Port
	SourcePort      *Port
}

// OutboundPipeContext provides context when this port creates a pipe.
type OutboundPipeContext struct {
	SourcePort      *Port
	DestinationPort *Port
}

// Hooks is a registry of all hook types for Port.
// Port hooks may fire from concurrent activation goroutines (components call
// PutSignals on their own ports while activating), so hook functions must be
// safe for concurrent use when they touch shared state.
type Hooks struct {
	onSignalsAdded     *hook.Group[*SignalsAddedContext]
	onSignalsDelivered *hook.Group[*SignalsDeliveredContext]
	onClear            *hook.Group[*ClearContext]
	onInboundPipe      *hook.Group[*InboundPipeContext]
	onOutboundPipe     *hook.Group[*OutboundPipeContext]
}

// newHooks creates a new hooks registry.
func newHooks() *Hooks {
	return &Hooks{
		onSignalsAdded:     hook.NewGroup[*SignalsAddedContext](),
		onSignalsDelivered: hook.NewGroup[*SignalsDeliveredContext](),
		onClear:            hook.NewGroup[*ClearContext](),
		onInboundPipe:      hook.NewGroup[*InboundPipeContext](),
		onOutboundPipe:     hook.NewGroup[*OutboundPipeContext](),
	}
}

// OnSignalsAdded registers a hook called when signals are added to the port.
func (h *Hooks) OnSignalsAdded(fn func(context.Context, *SignalsAddedContext) error) *Hooks {
	h.onSignalsAdded.Add(fn)
	return h
}

// OnSignalsDelivered registers a hook called when signals are delivered from
// this port to one destination. Only [Port.Flush] fires it, once per pipe, after
// the destination has accepted the signals -- a refused delivery is not reported
// as delivered, and the standalone [ForwardSignals], [ForwardWithFilter] and
// [ForwardWithMap] stay silent because they are not pipe deliveries.
//
// Unlike OnSignalsAdded, which gates the put and rolls the port back when it
// fails, this hook is an observer: the destination already holds the signals by
// the time it runs, so a failure fails the flush without undoing the delivery.
func (h *Hooks) OnSignalsDelivered(fn func(context.Context, *SignalsDeliveredContext) error) *Hooks {
	h.onSignalsDelivered.Add(fn)
	return h
}

// OnClear registers a hook called when signals are cleared from the port.
func (h *Hooks) OnClear(fn func(context.Context, *ClearContext) error) *Hooks {
	h.onClear.Add(fn)
	return h
}

// OnInboundPipe registers a hook called when a pipe is created TO this port.
func (h *Hooks) OnInboundPipe(fn func(context.Context, *InboundPipeContext) error) *Hooks {
	h.onInboundPipe.Add(fn)
	return h
}

// OnOutboundPipe registers a hook called when this port creates a pipe.
func (h *Hooks) OnOutboundPipe(fn func(context.Context, *OutboundPipeContext) error) *Hooks {
	h.onOutboundPipe.Add(fn)
	return h
}
