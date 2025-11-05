package port

import (
	"github.com/hovsep/fmesh/hook"
	"github.com/hovsep/fmesh/signal"
)

// PutContext provides context when signals are added to a port.
type PutContext struct {
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
	onSignalsAdded *hook.Group[*PutContext]
	onClear        *hook.Group[*ClearContext]
	onInboundPipe  *hook.Group[*InboundPipeContext]
	onOutboundPipe *hook.Group[*OutboundPipeContext]
}

// NewHooks creates a new hooks registry.
func NewHooks() *Hooks {
	return &Hooks{
		onSignalsAdded: hook.NewGroup[*PutContext](),
		onClear:        hook.NewGroup[*ClearContext](),
		onInboundPipe:  hook.NewGroup[*InboundPipeContext](),
		onOutboundPipe: hook.NewGroup[*OutboundPipeContext](),
	}
}

// OnSignalsAdded registers a hook called when signals are added to the port.
func (h *Hooks) OnSignalsAdded(fn func(*PutContext)) {
	h.onSignalsAdded.Add(fn)
}

// OnClear registers a hook called when signals are cleared from the port.
func (h *Hooks) OnClear(fn func(*ClearContext)) {
	h.onClear.Add(fn)
}

// OnInboundPipe registers a hook called when a pipe is created TO this port.
func (h *Hooks) OnInboundPipe(fn func(*InboundPipeContext)) {
	h.onInboundPipe.Add(fn)
}

// OnOutboundPipe registers a hook called when this port creates a pipe.
func (h *Hooks) OnOutboundPipe(fn func(*OutboundPipeContext)) {
	h.onOutboundPipe.Add(fn)
}
