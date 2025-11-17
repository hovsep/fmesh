package hooks

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
)

func TestPortHooks_OnSignalsAdded(t *testing.T) {
	var hookFired bool
	var portName string
	var signalsAdded int

	p := port.NewInput("data").
		SetupHooks(func(h *port.Hooks) {
			h.OnSignalsAdded(func(ctx *port.SignalsAddedContext) error {
				hookFired = true
				portName = ctx.Port.Name()
				signalsAdded = len(ctx.SignalsAdded)
				return nil
			})
		})

	p.PutSignals(signal.New(1), signal.New(2), signal.New(3))

	assert.True(t, hookFired)
	assert.Equal(t, "data", portName)
	assert.Equal(t, 3, signalsAdded)
	assert.Equal(t, 3, p.Signals().Len())
}

func TestPortHooks_OnSignalsAdded_MultipleCalls(t *testing.T) {
	var callCount int
	var totalSignalsHistory []int

	p := port.NewOutput("result").
		SetupHooks(func(h *port.Hooks) {
			h.OnSignalsAdded(func(ctx *port.SignalsAddedContext) error {
				callCount++
				totalSignalsHistory = append(totalSignalsHistory, ctx.Port.Signals().Len())
				return nil
			})
		})

	p.PutSignals(signal.New(1))
	p.PutSignals(signal.New(2), signal.New(3))
	p.PutSignals(signal.New(4))

	assert.Equal(t, 3, callCount)
	assert.Equal(t, []int{1, 3, 4}, totalSignalsHistory)
}

func TestPortHooks_OnClear(t *testing.T) {
	var clearFired bool
	var signalsCleared int

	p := port.NewInput("data").
		SetupHooks(func(h *port.Hooks) {
			h.OnClear(func(ctx *port.ClearContext) error {
				clearFired = true
				signalsCleared = ctx.SignalsCleared
				return nil
			})
		})

	p.PutSignals(signal.New(1), signal.New(2), signal.New(3), signal.New(4))
	p.Clear()

	assert.True(t, clearFired)
	assert.Equal(t, 4, signalsCleared)
	assert.Equal(t, 0, p.Signals().Len())
}

func TestPortHooks_OnClear_EmptyPort(t *testing.T) {
	var clearFired bool
	var signalsCleared int

	p := port.NewInput("data").
		SetupHooks(func(h *port.Hooks) {
			h.OnClear(func(ctx *port.ClearContext) error {
				clearFired = true
				signalsCleared = ctx.SignalsCleared
				return nil
			})
		})

	p.Clear()

	assert.True(t, clearFired)
	assert.Equal(t, 0, signalsCleared)
}

func TestPortHooks_OnOutboundPipe(t *testing.T) {
	var outboundFired bool
	var sourceName string
	var destName string

	outPort := port.NewOutput("out").
		SetupHooks(func(h *port.Hooks) {
			h.OnOutboundPipe(func(ctx *port.OutboundPipeContext) error {
				outboundFired = true
				sourceName = ctx.SourcePort.Name()
				destName = ctx.DestinationPort.Name()
				return nil
			})
		})

	inPort := port.NewInput("in")

	outPort.PipeTo(inPort)

	assert.True(t, outboundFired)
	assert.Equal(t, "out", sourceName)
	assert.Equal(t, "in", destName)
}

func TestPortHooks_OnInboundPipe(t *testing.T) {
	var inboundFired bool
	var sourceName string
	var destName string

	outPort := port.NewOutput("out")

	inPort := port.NewInput("in").
		SetupHooks(func(h *port.Hooks) {
			h.OnInboundPipe(func(ctx *port.InboundPipeContext) error {
				inboundFired = true
				sourceName = ctx.SourcePort.Name()
				destName = ctx.DestinationPort.Name()
				return nil
			})
		})

	outPort.PipeTo(inPort)

	assert.True(t, inboundFired)
	assert.Equal(t, "out", sourceName)
	assert.Equal(t, "in", destName)
}

func TestPortHooks_OnOutboundAndInbound_BothFire(t *testing.T) {
	var outboundFired bool
	var inboundFired bool

	outPort := port.NewOutput("out").
		SetupHooks(func(h *port.Hooks) {
			h.OnOutboundPipe(func(ctx *port.OutboundPipeContext) error {
				outboundFired = true
				return nil
			})
		})

	inPort := port.NewInput("in").
		SetupHooks(func(h *port.Hooks) {
			h.OnInboundPipe(func(ctx *port.InboundPipeContext) error {
				inboundFired = true
				return nil
			})
		})

	outPort.PipeTo(inPort)

	assert.True(t, outboundFired, "Outbound hook should fire")
	assert.True(t, inboundFired, "Inbound hook should fire")
}

func TestPortHooks_OnOutboundPipe_MultipleDest(t *testing.T) {
	var outboundCount int
	var destNames []string

	outPort := port.NewOutput("out").
		SetupHooks(func(h *port.Hooks) {
			h.OnOutboundPipe(func(ctx *port.OutboundPipeContext) error {
				outboundCount++
				destNames = append(destNames, ctx.DestinationPort.Name())
				return nil
			})
		})

	in1 := port.NewInput("in1")
	in2 := port.NewInput("in2")
	in3 := port.NewInput("in3")

	outPort.PipeTo(in1, in2, in3)

	assert.Equal(t, 3, outboundCount)
	assert.Equal(t, []string{"in1", "in2", "in3"}, destNames)
}

func TestPortHooks_MultipleHooksPerType(t *testing.T) {
	var log []string

	p := port.NewInput("data").
		SetupHooks(func(h *port.Hooks) {
			h.OnSignalsAdded(func(ctx *port.SignalsAddedContext) error {
				log = append(log, "put1")
				return nil
			})
			h.OnSignalsAdded(func(ctx *port.SignalsAddedContext) error {
				log = append(log, "put2")
				return nil
			})
			h.OnClear(func(ctx *port.ClearContext) error {
				log = append(log, "clear1")
				return nil
			})
			h.OnClear(func(ctx *port.ClearContext) error {
				log = append(log, "clear2")
				return nil
			})
		})

	p.PutSignals(signal.New(1))
	p.Clear()

	assert.Equal(t, []string{"put1", "put2", "clear1", "clear2"}, log)
}

func TestPortHooks_ContextAccess(t *testing.T) {
	var portName string
	var signalPayloads []int

	p := port.NewInput("sensor").
		SetupHooks(func(h *port.Hooks) {
			h.OnSignalsAdded(func(ctx *port.SignalsAddedContext) error {
				portName = ctx.Port.Name()
				// Access actual signal data
				for _, sig := range ctx.SignalsAdded {
					payload, err := sig.Payload()
					if err == nil {
						if val, ok := payload.(int); ok {
							signalPayloads = append(signalPayloads, val)
						}
					}
				}
				return nil
			})
		})

	p.PutSignals(signal.New(100), signal.New(200), signal.New(300))

	assert.Equal(t, "sensor", portName)
	assert.Equal(t, []int{100, 200, 300}, signalPayloads)
}

func TestPortHooks_PracticalVolumeMonitoring(t *testing.T) {
	// Practical example: Monitor signal throughput
	type VolumeMetrics struct {
		TotalPuts         int
		TotalSignalsAdded int
		MaxSignalsAtOnce  int
		TotalClears       int
	}
	metrics := VolumeMetrics{}

	p := port.NewOutput("stream").
		SetupHooks(func(h *port.Hooks) {
			h.OnSignalsAdded(func(ctx *port.SignalsAddedContext) error {
				metrics.TotalPuts++
				metrics.TotalSignalsAdded += len(ctx.SignalsAdded)
				if len(ctx.SignalsAdded) > metrics.MaxSignalsAtOnce {
					metrics.MaxSignalsAtOnce = len(ctx.SignalsAdded)
				}
				return nil
			})
			h.OnClear(func(ctx *port.ClearContext) error {
				metrics.TotalClears++
				return nil
			})
		})

	// Simulate data flow
	p.PutSignals(signal.New(1))
	p.PutSignals(signal.New(2), signal.New(3))
	p.Clear()
	p.PutSignals(signal.New(4), signal.New(5), signal.New(6), signal.New(7))
	p.Clear()

	assert.Equal(t, 3, metrics.TotalPuts)
	assert.Equal(t, 7, metrics.TotalSignalsAdded)
	assert.Equal(t, 4, metrics.MaxSignalsAtOnce)
	assert.Equal(t, 2, metrics.TotalClears)
}

func TestPortHooks_PracticalTopologyTracking(t *testing.T) {
	// Practical example: Track mesh topology
	type TopologyMap struct {
		Connections map[string][]string // source -> destinations
	}
	topology := TopologyMap{
		Connections: make(map[string][]string),
	}

	out1 := port.NewOutput("out1").
		SetupHooks(func(h *port.Hooks) {
			h.OnOutboundPipe(func(ctx *port.OutboundPipeContext) error {
				srcName := ctx.SourcePort.Name()
				destName := ctx.DestinationPort.Name()
				topology.Connections[srcName] = append(topology.Connections[srcName], destName)
				return nil
			})
		})

	out2 := port.NewOutput("out2").
		SetupHooks(func(h *port.Hooks) {
			h.OnOutboundPipe(func(ctx *port.OutboundPipeContext) error {
				srcName := ctx.SourcePort.Name()
				destName := ctx.DestinationPort.Name()
				topology.Connections[srcName] = append(topology.Connections[srcName], destName)
				return nil
			})
		})

	in1 := port.NewInput("in1")
	in2 := port.NewInput("in2")
	in3 := port.NewInput("in3")

	// Create topology: out1 -> in1, in2; out2 -> in2, in3
	out1.PipeTo(in1, in2)
	out2.PipeTo(in2, in3)

	assert.Equal(t, []string{"in1", "in2"}, topology.Connections["out1"])
	assert.Equal(t, []string{"in2", "in3"}, topology.Connections["out2"])
}

func TestPortHooks_PracticalDataValidation(t *testing.T) {
	// Practical example: Validate incoming data
	p := port.NewInput("validated").
		SetupHooks(func(h *port.Hooks) {
			h.OnSignalsAdded(func(ctx *port.SignalsAddedContext) error {
				// Validate: must receive exactly 3 signals
				if len(ctx.SignalsAdded) != 3 {
					return errors.New("expected 3 signals")
				}

				// Validate: all payloads must be positive integers
				for _, sig := range ctx.SignalsAdded {
					payload, err := sig.Payload()
					if err != nil {
						return errors.New("payload error")
					}
					if val, ok := payload.(int); !ok || val <= 0 {
						return errors.New("invalid payload")
					}
				}

				return nil
			})
		})

	p.PutSignals(signal.New(10), signal.New(20), signal.New(30))
	assert.False(t, p.HasChainableErr())

	p.PutSignals(signal.New(900))
	assert.True(t, p.HasChainableErr())
	assert.ErrorContains(t, p.ChainableErr(), "expected 3 signals")
}
