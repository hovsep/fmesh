package hooks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

// Port hooks: signals arriving and leaving, and pipes being wired.

func TestPortHooks_OnSignalsAdded(t *testing.T) {
	var calls, lastBatch int
	var portName string

	p := testutil.MustInputPort("data").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsAdded(func(_ context.Context, added *port.SignalsAddedContext) error {
			calls++
			lastBatch = len(added.SignalsAdded)
			portName = added.Port.Name()
			return nil
		})
	})

	require.NoError(t, p.PutSignals(signal.New(1)))
	require.NoError(t, p.PutSignals(signal.New(2), signal.New(3)))

	// One call per Put, carrying only the signals of that Put.
	assert.Equal(t, 2, calls)
	assert.Equal(t, 2, lastBatch)
	assert.Equal(t, "data", portName)
	assert.Equal(t, 3, p.Signals().Len())
}

func TestPortHooks_OnSignalsAddedFailureRollsBackThePort(t *testing.T) {
	p := testutil.MustInputPort("data").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsAdded(func(context.Context, *port.SignalsAddedContext) error {
			return errors.New("rejected")
		})
	})

	err := p.PutSignals(signal.New(1))

	require.Error(t, err)
	assert.True(t, p.Signals().IsEmpty(), "a rejected Put must leave the port untouched")
}

func TestPortHooks_OnClear(t *testing.T) {
	var calls, cleared int

	p := testutil.MustInputPort("data").SetupHooks(func(h *port.Hooks) {
		h.OnClear(func(_ context.Context, ctx *port.ClearContext) error {
			calls++
			cleared = ctx.SignalsCleared
			return nil
		})
	})

	require.NoError(t, p.PutSignals(signal.New(1), signal.New(2)))
	require.NoError(t, p.Clear(context.Background()))
	// Clearing an already-empty port still fires, reporting zero.
	require.NoError(t, p.Clear(context.Background()))

	assert.Equal(t, 2, calls)
	assert.Equal(t, 0, cleared)
}

func TestPortHooks_PipeHooksFireOnBothEnds(t *testing.T) {
	var outbound, inbound []string

	source := testutil.MustOutputPort("out").SetupHooks(func(h *port.Hooks) {
		h.OnOutboundPipe(func(_ context.Context, pipe *port.OutboundPipeContext) error {
			outbound = append(outbound, pipe.SourcePort.Name()+"->"+pipe.DestinationPort.Name())
			return nil
		})
	})
	dest1 := testutil.MustInputPort("in1").SetupHooks(func(h *port.Hooks) {
		h.OnInboundPipe(func(_ context.Context, pipe *port.InboundPipeContext) error {
			inbound = append(inbound, pipe.SourcePort.Name()+"->"+pipe.DestinationPort.Name())
			return nil
		})
	})
	dest2 := testutil.MustInputPort("in2")

	require.NoError(t, source.PipeTo(dest1, dest2))

	// One outbound event per destination; the inbound hook fires only on the port
	// that registered it.
	assert.Equal(t, []string{"out->in1", "out->in2"}, outbound)
	assert.Equal(t, []string{"out->in1"}, inbound)
}

func TestPortHooks_FailingPipeHookRemovesThatPipe(t *testing.T) {
	source := testutil.MustOutputPort("out")
	ok := testutil.MustInputPort("ok")
	rejecting := testutil.MustInputPort("rejecting").SetupHooks(func(h *port.Hooks) {
		h.OnInboundPipe(func(context.Context, *port.InboundPipeContext) error { return errors.New("no") })
	})

	require.ErrorContains(t, source.PipeTo(ok, rejecting), "onInboundPipe hook failed")

	assert.Equal(t, []*port.Port{ok}, source.Pipes().All(), "the earlier pipe stays, the rejected one is gone")
}

func TestPortHooks_MultipleHooksRunInRegistrationOrder(t *testing.T) {
	var log recorder

	p := testutil.MustInputPort("data").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsAdded(func(context.Context, *port.SignalsAddedContext) error {
			log.add("first")
			return nil
		})
		h.OnSignalsAdded(func(context.Context, *port.SignalsAddedContext) error {
			log.add("second")
			return nil
		})
	})

	require.NoError(t, p.PutSignals(signal.New(1)))

	assert.Equal(t, []string{"first", "second"}, log.events)
}

func TestPortHooks_FireDuringDelivery(t *testing.T) {
	// The case that matters for port hooks: signals arriving through a pipe during
	// a run, rather than being seeded by hand.
	var deliveries int

	producer := testutil.MustComponent("producer",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("out").PutSignals(signal.New("delivered"))
		}))
	consumer := testutil.MustComponent("consumer",
		component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	consumer.InputByName("in").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsAdded(func(context.Context, *port.SignalsAddedContext) error {
			deliveries++
			return nil
		})
	})

	fm := testutil.MustFMesh("delivery")
	require.NoError(t, fm.AddComponents(producer, consumer))
	require.NoError(t, producer.OutputByName("out").PipeTo(consumer.InputByName("in")))
	require.NoError(t, producer.InputByName("in").PutSignals(signal.New("go")))

	_, err := fm.Run(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 1, deliveries, "the hook must see the signal arrive through the pipe")
}

func TestPortHooks_OnSignalsDeliveredNamesBothEnds(t *testing.T) {
	// OnSignalsAdded fires on the destination and cannot say who sent the batch.
	// This is the hook that can, and it fires once per pipe.
	type edge struct{ from, to string }
	var edges []edge

	producer := testutil.MustComponent("producer",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("out").PutSignals(signal.New("delivered"))
		}))
	left := testutil.MustComponent("left",
		component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))
	right := testutil.MustComponent("right",
		component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	var hasDeadline bool
	producer.OutputByName("out").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsDelivered(func(ctx context.Context, d *port.SignalsDeliveredContext) error {
			_, hasDeadline = ctx.Deadline()
			edges = append(edges, edge{
				from: d.SourcePort.ParentComponent().Name() + "." + d.SourcePort.Name(),
				to:   d.DestinationPort.ParentComponent().Name() + "." + d.DestinationPort.Name(),
			})
			return nil
		})
	})

	fm := testutil.MustFMesh("delivery")
	require.NoError(t, fm.AddComponents(producer, left, right))
	require.NoError(t, producer.OutputByName("out").
		PipeTo(left.InputByName("in"), right.InputByName("in")))
	require.NoError(t, producer.InputByName("in").PutSignals(signal.New("go")))

	_, err := fm.Run(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []edge{
		{from: "producer.out", to: "left.in"},
		{from: "producer.out", to: "right.in"},
	}, edges)
	// Delivery only ever happens inside a run, so the hook always gets the run
	// context -- narrowed by TimeLimit, never context.Background().
	assert.True(t, hasDeadline, "the drain passes the run context, deadline and all")
}

func TestPortHooks_DeliveredFiresAfterTheDestinationAccepted(t *testing.T) {
	// Ordering matters for anyone counting: the destination's gate runs first and
	// may still reject, so the source-side event is the one that means "moved".
	var log recorder

	producer := testutil.MustComponent("producer",
		component.WithInputs("in"), component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("out").PutSignals(signal.New("delivered"))
		}))
	consumer := testutil.MustComponent("consumer",
		component.WithInputs("in"),
		component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))

	consumer.InputByName("in").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsAdded(func(context.Context, *port.SignalsAddedContext) error {
			log.add("added")
			return nil
		})
	})
	producer.OutputByName("out").SetupHooks(func(h *port.Hooks) {
		h.OnSignalsDelivered(func(context.Context, *port.SignalsDeliveredContext) error {
			log.add("delivered")
			return nil
		})
	})

	fm := testutil.MustFMesh("ordering")
	require.NoError(t, fm.AddComponents(producer, consumer))
	require.NoError(t, producer.OutputByName("out").PipeTo(consumer.InputByName("in")))
	require.NoError(t, producer.InputByName("in").PutSignals(signal.New("go")))

	_, err := fm.Run(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"added", "delivered"}, log.events)
}

func TestPortHooks_FailuresDuringTheDrainFailTheRun(t *testing.T) {
	t.Run("a refusing OnClear on an input", func(t *testing.T) {
		fm := twoCycleMesh(t)
		fm.ComponentByName("producer").InputByName("in").SetupHooks(func(h *port.Hooks) {
			h.OnClear(func(context.Context, *port.ClearContext) error { return errors.New("sticky") })
		})

		_, err := fm.Run(context.Background())

		require.ErrorIs(t, err, fmesh.ErrFailedToDrain)
		require.ErrorContains(t, err, `failed to clear input ports: component "producer"`)
	})

	t.Run("a refusing OnSignalsAdded on the destination", func(t *testing.T) {
		fm := twoCycleMesh(t)
		fm.ComponentByName("consumer").InputByName("in").SetupHooks(func(h *port.Hooks) {
			h.OnSignalsAdded(func(context.Context, *port.SignalsAddedContext) error { return errors.New("refused") })
		})

		_, err := fm.Run(context.Background())

		require.ErrorIs(t, err, fmesh.ErrFailedToDrain)
		require.ErrorContains(t, err, `failed to flush outputs of component "producer"`)
	})

	t.Run("a refusing OnClear on an output fails before the first cycle", func(t *testing.T) {
		// Outputs are cleared when a run starts, so this one never gets going.
		fm := twoCycleMesh(t)
		fm.ComponentByName("producer").OutputByName("out").SetupHooks(func(h *port.Hooks) {
			h.OnClear(func(context.Context, *port.ClearContext) error { return errors.New("sticky") })
		})

		ri, err := fm.Run(context.Background())

		require.ErrorContains(t, err, "onClear hook failed")
		assert.Nil(t, ri, "no run started, so no runtime info")
	})
}
