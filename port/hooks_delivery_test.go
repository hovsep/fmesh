package port

import (
	"context"
	"errors"
	"testing"

	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// delivery is one recorded OnSignalsDelivered event, flattened so assertions can
// compare names rather than pointers.
type delivery struct {
	source      string
	destination string
	signals     []*signal.Signal
}

// recordDeliveries registers a hook on src collecting every delivery it makes.
func recordDeliveries(src *Port) *[]delivery {
	recorded := new([]delivery)
	src.SetupHooks(func(hooks *Hooks) {
		hooks.OnSignalsDelivered(func(_ context.Context, d *SignalsDeliveredContext) error {
			*recorded = append(*recorded, delivery{
				source:      d.SourcePort.Name(),
				destination: d.DestinationPort.Name(),
				signals:     d.SignalsDelivered,
			})
			return nil
		})
	})
	return recorded
}

func TestPort_OnSignalsDelivered(t *testing.T) {
	t.Run("fires once per pipe, carrying the batch", func(t *testing.T) {
		// The whole point of the hook: OnSignalsAdded on the destinations sees
		// two unrelated puts, this sees one source fanning out to two pipes.
		src := mustOutput("out")
		dst1, dst2 := mustInput("in1"), mustInput("in2")
		require.NoError(t, src.PipeTo(dst1, dst2))
		require.NoError(t, src.PutPayloads(1, 2, 3))

		recorded := recordDeliveries(src)
		require.NoError(t, src.Flush(context.Background()))

		require.Len(t, *recorded, 2)
		assert.Equal(t, "out", (*recorded)[0].source)
		assert.Equal(t, []string{"in1", "in2"},
			[]string{(*recorded)[0].destination, (*recorded)[1].destination},
			"destinations are reported in pipe order")
		assert.Len(t, (*recorded)[0].signals, 3, "the whole batch, not one event per signal")
		assert.Len(t, (*recorded)[1].signals, 3)
	})

	t.Run("fan-out reports the same signal pointers to every destination", func(t *testing.T) {
		// Flush shares one *Signal across all destinations and the hook must not
		// disturb that -- the batch is materialized once for the whole fan-out.
		src := mustOutput("out")
		require.NoError(t, src.PipeTo(mustInput("in1"), mustInput("in2")))
		require.NoError(t, src.PutPayloads(1))

		recorded := recordDeliveries(src)
		require.NoError(t, src.Flush(context.Background()))

		require.Len(t, *recorded, 2)
		assert.Same(t, (*recorded)[0].signals[0], (*recorded)[1].signals[0])
	})

	t.Run("a flush that moves nothing fires nothing", func(t *testing.T) {
		// Unlike OnClear, which fires even for an empty port, this hook only ever
		// reports a transfer that happened.
		noPipes := mustOutput("out")
		require.NoError(t, noPipes.PutPayloads(1))
		recordedNoPipes := recordDeliveries(noPipes)
		require.NoError(t, noPipes.Flush(context.Background()))
		assert.Empty(t, *recordedNoPipes, "signals but nowhere to send them")

		noSignals := mustOutput("out")
		require.NoError(t, noSignals.PipeTo(mustInput("in")))
		recordedNoSignals := recordDeliveries(noSignals)
		require.NoError(t, noSignals.Flush(context.Background()))
		assert.Empty(t, *recordedNoSignals, "a destination but nothing to deliver")
	})

	t.Run("a refused delivery is not reported as delivered", func(t *testing.T) {
		// This is what makes a source-side counter more accurate than counting on
		// the destination's OnSignalsAdded, which fires before the put is known
		// to have stuck.
		src := mustOutput("out")
		refusing, accepting := mustInput("refusing"), mustInput("accepting")
		refusing.SetupHooks(func(hooks *Hooks) {
			hooks.OnSignalsAdded(func(context.Context, *SignalsAddedContext) error {
				return errors.New("refused")
			})
		})
		require.NoError(t, src.PipeTo(refusing, accepting))
		require.NoError(t, src.PutPayloads(1))

		recorded := recordDeliveries(src)
		err := src.Flush(context.Background())

		require.Error(t, err)
		require.ErrorContains(t, err, "refused")
		require.Len(t, *recorded, 1, "only the destination that accepted")
		assert.Equal(t, "accepting", (*recorded)[0].destination)
		assert.True(t, src.HasSignals(), "a partial failure leaves the source uncleared")
	})

	t.Run("a failing hook fails the flush without undoing the delivery", func(t *testing.T) {
		// The hook is an observer, not a gate: by the time it runs the
		// destination already holds the signals and there is no un-deliver.
		src := mustOutput("out")
		dst := mustInput("in")
		require.NoError(t, src.PipeTo(dst))
		require.NoError(t, src.PutPayloads(1))

		src.SetupHooks(func(hooks *Hooks) {
			hooks.OnSignalsDelivered(func(context.Context, *SignalsDeliveredContext) error {
				return errors.New("observer failed")
			})
		})

		err := src.Flush(context.Background())

		require.Error(t, err)
		require.ErrorContains(t, err, "onSignalsDelivered hook failed")
		assert.True(t, dst.HasSignals(), "the delivery stands")
		assert.True(t, src.HasSignals(), "and the source is not cleared")
	})

	t.Run("the source still holds its signals when the hook runs", func(t *testing.T) {
		// Fires before Clear, so a hook can inspect what is being sent.
		src := mustOutput("out")
		require.NoError(t, src.PipeTo(mustInput("in")))
		require.NoError(t, src.PutPayloads(1, 2))

		var lenAtHookTime int
		src.SetupHooks(func(hooks *Hooks) {
			hooks.OnSignalsDelivered(func(_ context.Context, d *SignalsDeliveredContext) error {
				lenAtHookTime = d.SourcePort.Signals().Len()
				return nil
			})
		})
		require.NoError(t, src.Flush(context.Background()))

		assert.Equal(t, 2, lenAtHookTime)
		assert.False(t, src.HasSignals(), "cleared afterwards, as always")
	})

	t.Run("forwarding by hand is not a pipe delivery", func(t *testing.T) {
		// ForwardSignals and friends move signals between arbitrary ports; only
		// Flush is a delivery through a pipe.
		src := mustOutput("out")
		dst := mustInput("in")
		require.NoError(t, src.PutPayloads(1))

		recorded := recordDeliveries(src)
		require.NoError(t, ForwardSignals(context.Background(), src, dst))
		require.NoError(t, ForwardWithFilter(context.Background(), src, dst,
			func(*signal.Signal) bool { return true }))

		assert.Empty(t, *recorded)
	})
}
