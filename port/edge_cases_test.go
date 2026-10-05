package port

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh/signal"
)

func refusing(p *Port) *Port {
	p.SetupHooks(func(h *Hooks) {
		h.OnSignalsAdded(func(context.Context, *SignalsAddedContext) error { return errors.New("refused") })
	})
	return p
}

func TestNewInput_OptionFailureNamesThePort(t *testing.T) {
	_, err := NewInput("p", func(*Port) error { return errors.New("bad option") })
	require.ErrorContains(t, err, `port "p" option failed`)
	require.ErrorContains(t, err, "bad option")
}

func TestPort_PutSignalGroups_PropagatesHookFailure(t *testing.T) {
	p := refusing(mustInput("in"))
	require.ErrorContains(t, p.PutSignalGroups(signal.NewGroup(1)), "refused")
}

func TestPort_Clear_PropagatesHookFailure(t *testing.T) {
	p := mustInput("in")
	p.SetupHooks(func(h *Hooks) {
		h.OnClear(func(context.Context, *ClearContext) error { return errors.New("sticky") })
	})
	require.ErrorContains(t, p.Clear(context.Background()), "onClear hook failed")
}

func TestPort_PipeTo_PropagatesOutboundHookFailure(t *testing.T) {
	out := mustOutput("out")
	out.SetupHooks(func(h *Hooks) {
		h.OnOutboundPipe(func(context.Context, *OutboundPipeContext) error { return errors.New("no wiring") })
	})
	require.ErrorContains(t, out.PipeTo(mustInput("in")), "onOutboundPipe hook failed")
}

func TestMultiForward_NamesTheFailingPair(t *testing.T) {
	src := mustOutput("src")
	require.NoError(t, src.PutSignals(signal.New(1)))

	err := MultiForward(context.Background(), Pair{From: src, To: refusing(mustInput("dst"))})
	require.ErrorContains(t, err, "forwarding")
	require.ErrorContains(t, err, "refused")
}

func TestCollection_AnyRemoveAndMetadata(t *testing.T) {
	assert.Nil(t, mustNewCollection().First())

	col := mustNewCollection(mustInput("b"), mustInput("a"))
	assert.Equal(t, "a", col.First().Name(), "name order, so the answer is stable")

	col.Remove("a")
	assert.Nil(t, col.ByName("a"))
	assert.Equal(t, 1, col.Len())

	col.Meta().Set("k", "v")
	col.Meta().Set("s", 1.0)
	assert.True(t, col.Meta().ValueIs("k", "v"))
	assert.True(t, col.Meta().ValueIs("s", 1.0))
}

func TestCollection_Map_RejectsCollidingNames(t *testing.T) {
	col := mustNewCollection(mustInput("a"), mustInput("b"))
	_, err := col.Map(func(*Port) *Port { return mustInput("same") })
	require.Error(t, err)
}

func TestCollection_ErrorsSurfaceFromEachPort(t *testing.T) {
	t.Run("PutSignalsOnEach", func(t *testing.T) {
		col := mustNewCollection(refusing(mustInput("a")))
		require.ErrorContains(t, col.PutSignalsOnEach(signal.New(1)), "refused")
	})

	t.Run("Flush", func(t *testing.T) {
		out := mustOutput("out")
		require.NoError(t, out.PipeTo(refusing(mustInput("in"))))
		require.NoError(t, out.PutSignals(signal.New(1)))
		require.ErrorContains(t, mustNewCollection(out).Flush(context.Background()), "refused")
	})

	t.Run("PipeEachTo", func(t *testing.T) {
		col := mustNewCollection(mustInput("a"))
		require.ErrorIs(t, col.PipeEachTo(mustInput("b")), ErrInvalidPipeDirection)
	})
}

func TestGroup_OwnMetadata(t *testing.T) {
	g := NewGroup()
	g.Meta().Set("k", "v").Set("s", 2.0)
	assert.True(t, g.Meta().ValueIs("k", "v"))
	assert.True(t, g.Meta().ValueIs("s", 2.0))
}
