package port

import (
	"github.com/hovsep/fmesh/hook"
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
type Hooks struct {
	onSignalsAdded *hook.Group[*SignalsAddedContext]
	onClear        *hook.Group[*ClearContext]
	onInboundPipe  *hook.Group[*InboundPipeContext]
	onOutboundPipe *hook.Group[*OutboundPipeContext]
}

// NewHooks creates a new hooks registry.
func NewHooks() *Hooks {
	return &Hooks{
		onSignalsAdded: hook.NewGroup[*SignalsAddedContext](),
		onClear:        hook.NewGroup[*ClearContext](),
		onInboundPipe:  hook.NewGroup[*InboundPipeContext](),
		onOutboundPipe: hook.NewGroup[*OutboundPipeContext](),
	}
}

// OnSignalsAdded registers a hook called when signals are added to the port.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnSignalsAdded(fn func(*SignalsAddedContext) error) *Hooks {
	h.onSignalsAdded.Add(fn)
	return h
}

// OnClear registers a hook called when signals are cleared from the port.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnClear(fn func(*ClearContext) error) *Hooks {
	h.onClear.Add(fn)
	return h
}

// OnInboundPipe registers a hook called when a pipe is created TO this port.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnInboundPipe(fn func(*InboundPipeContext) error) *Hooks {
	h.onInboundPipe.Add(fn)
	return h
}

// OnOutboundPipe registers a hook called when this port creates a pipe.
// Returns the Hooks registry for method chaining.
func (h *Hooks) OnOutboundPipe(fn func(*OutboundPipeContext) error) *Hooks {
	h.onOutboundPipe.Add(fn)
	return h
}
