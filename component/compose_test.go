package component

import (
	"context"
	"errors"
	"testing"

	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSequential(t *testing.T) {
	t.Run("runs in order", func(t *testing.T) {
		var order []string
		note := func(s string) ActivationFunc {
			return func(context.Context, *Component) error { order = append(order, s); return nil }
		}

		require.NoError(t, Sequential(note("a"), note("b"), note("c"))(context.Background(), nil))
		assert.Equal(t, []string{"a", "b", "c"}, order)
	})

	t.Run("stops at the first error", func(t *testing.T) {
		reached := false
		boom := func(context.Context, *Component) error { return errors.New("boom") }
		after := func(context.Context, *Component) error { reached = true; return nil }

		err := Sequential(boom, after)(context.Background(), nil)

		require.ErrorContains(t, err, "boom")
		assert.False(t, reached, "nothing after a failed stage runs")
	})

	t.Run("a waiting stage suspends the component", func(t *testing.T) {
		// The error passes through as the component's own, so a stage can
		// suspend exactly as it would if the activation were one function.
		wait := func(context.Context, *Component) error { return ErrWaitingForInputsKeep }

		err := Sequential(wait, func(context.Context, *Component) error { return nil })(context.Background(), nil)

		require.ErrorIs(t, err, ErrWaitingForInputsKeep)
	})
}

func TestWhenAndRequireInputs(t *testing.T) {
	newC := func(t *testing.T) *Component {
		t.Helper()
		c, err := New("c", WithInputs("a", "b"),
			WithActivationFunc(func(context.Context, *Component) error { return nil }))
		require.NoError(t, err)
		return c
	}

	t.Run("When skips the work when the predicate fails", func(t *testing.T) {
		c := newC(t)
		ran := false

		err := When(HasSignalsOn("a"), func(context.Context, *Component) error { ran = true; return nil })(context.Background(), c)

		require.NoError(t, err)
		assert.False(t, ran, "an idle component simply does nothing")
	})

	t.Run("When runs the work when it holds", func(t *testing.T) {
		c := newC(t)
		require.NoError(t, c.InputByName("a").PutSignals(signal.New(1)))
		ran := false

		require.NoError(t, When(HasSignalsOn("a"), func(context.Context, *Component) error { ran = true; return nil })(context.Background(), c))
		assert.True(t, ran)
	})

	t.Run("RequireInputs waits instead of skipping", func(t *testing.T) {
		// A partial set of inputs must be kept, not silently dropped by
		// returning nil.
		c := newC(t)
		require.NoError(t, c.InputByName("a").PutSignals(signal.New(1)))

		err := RequireInputs("a", "b")(context.Background(), c)

		require.ErrorIs(t, err, ErrWaitingForInputsKeep)
	})

	t.Run("RequireInputs passes once everything has arrived", func(t *testing.T) {
		c := newC(t)
		require.NoError(t, c.InputByName("a").PutSignals(signal.New(1)))
		require.NoError(t, c.InputByName("b").PutSignals(signal.New(2)))

		require.NoError(t, RequireInputs("a", "b")(context.Background(), c))
	})

	t.Run("a port that does not exist is never carrying anything", func(t *testing.T) {
		// A misspelled name must not read as ready: an empty selection of ports
		// trivially satisfies "all of them have signals".
		c := newC(t)
		require.NoError(t, c.InputByName("a").PutSignals(signal.New(1)))

		assert.False(t, HasSignalsOn("typo")(c))
		assert.False(t, HasSignalsOn("a", "typo")(c))
	})

	t.Run("RequireInputs fails on a port that does not exist", func(t *testing.T) {
		c := newC(t)

		err := RequireInputs("typo")(context.Background(), c)

		require.ErrorContains(t, err, `required input port "typo" does not exist`)
		require.NotErrorIs(t, err, ErrWaitingForInputs,
			"nothing arrives on a port that does not exist, so waiting would suspend the component forever")
	})

	t.Run("an empty port does not hide a misspelled one behind a wait", func(t *testing.T) {
		// Every name is checked before any signal is, or the typo is only ever
		// reported once the real port happens to be full.
		c := newC(t)

		err := RequireInputs("a", "typo")(context.Background(), c)

		require.ErrorContains(t, err, `"typo" does not exist`)
	})
}

func TestPipeline(t *testing.T) {
	c, err := New("c", WithInputs("in"), WithOutputs("out"),
		WithActivationFunc(func(context.Context, *Component) error { return nil }))
	require.NoError(t, err)
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(2), signal.New(3)))

	double := func(g *signal.Group) (*signal.Group, error) {
		return g.Map(func(s *signal.Signal) *signal.Signal {
			return signal.New(s.PayloadOrDefault(0).(int) * 2)
		}), nil
	}

	require.NoError(t, Pipeline([]string{"in"}, "out", double, double)(context.Background(), c))

	got := c.OutputByName("out").Signals()
	require.Equal(t, 2, got.Len())
	assert.Equal(t, 8, got.First().PayloadOrDefault(0), "two doublings of 2")

	t.Run("a failing stage says which one", func(t *testing.T) {
		bad := func(*signal.Group) (*signal.Group, error) { return nil, errors.New("nope") }

		err := Pipeline([]string{"in"}, "out", double, bad)(context.Background(), c)

		require.ErrorContains(t, err, "pipeline stage 1 failed")
	})

	t.Run("a stage returning a nil group is an error, not a panic", func(t *testing.T) {
		var noGroup *signal.Group
		nilGroup := func(*signal.Group) (*signal.Group, error) { return noGroup, nil }

		err := Pipeline([]string{"in"}, "out", nilGroup)(context.Background(), c)

		require.ErrorContains(t, err, "pipeline stage 0 returned a nil group")
	})

	t.Run("ports that do not exist are errors, not panics", func(t *testing.T) {
		require.ErrorContains(t, Pipeline([]string{"in"}, "typo", double)(context.Background(), c),
			`pipeline output port "typo" does not exist`)
		require.ErrorContains(t, Pipeline([]string{"typo"}, "out", double)(context.Background(), c),
			`pipeline input port "typo" does not exist`)
	})
}

func TestPipelineReadsInputsInOrder(t *testing.T) {
	// The order is the one the caller wrote, not whatever a map iteration
	// produced: a stage that folds or pairs up its group depends on it.
	c, err := New("c", WithInputs("a", "b", "d"), WithOutputs("out"),
		WithActivationFunc(func(context.Context, *Component) error { return nil }))
	require.NoError(t, err)
	require.NoError(t, c.InputByName("a").PutSignals(signal.New("a")))
	require.NoError(t, c.InputByName("b").PutSignals(signal.New("b")))
	require.NoError(t, c.InputByName("d").PutSignals(signal.New("d")))

	require.NoError(t, Pipeline([]string{"d", "a", "b"}, "out")(context.Background(), c))

	payloads, err := c.OutputByName("out").Signals().AllPayloads()
	require.NoError(t, err)
	assert.Equal(t, []any{"d", "a", "b"}, payloads)
}
