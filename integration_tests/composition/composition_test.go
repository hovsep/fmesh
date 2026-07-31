// Package composition covers the two ways a component is built out of smaller
// pieces: the activation combinators, and running a whole mesh inside one
// component.
package composition

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/signal"
)

func TestCombinators_SequentialRunsStagesInOrder(t *testing.T) {
	var order []string

	stage := func(name string) component.ActivationFunc {
		return func(_ context.Context, this *component.Component) error {
			order = append(order, name)
			return this.OutputByName("out").PutSignals(signal.New(name))
		}
	}

	c := testutil.MustComponent("staged",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(component.Sequential(stage("a"), stage("b"), stage("c"))))

	fm := testutil.MustFMesh("sequential")
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(signal.New(1)))

	_, err := fm.Run(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"a", "b", "c"}, order)
	assert.Equal(t, 3, c.OutputByName("out").Signals().Len())
}

func TestCombinators_SequentialStopsAtTheFirstWait(t *testing.T) {
	var reached []string

	c := testutil.MustComponent("guarded",
		component.WithInputs("a", "b"),
		component.WithOutputs("out"),
		component.WithActivationFunc(component.Sequential(
			// RequireInputs suspends the component until every named port has signals.
			component.RequireInputs("a", "b"),
			func(_ context.Context, this *component.Component) error {
				reached = append(reached, "body")
				return this.OutputByName("out").PutSignals(signal.New("ran"))
			},
		)))

	fm := testutil.MustFMesh("gated")
	require.NoError(t, fm.AddComponents(c))
	// Only "a" is ever fed, so the guard never passes and the body never runs.
	require.NoError(t, c.InputByName("a").PutSignals(signal.New(1)))

	_, err := fm.Run(context.Background())
	require.ErrorIs(t, err, fmesh.ErrLivelockDetected)

	assert.Empty(t, reached, "a suspended component must not run its later stages")
	assert.True(t, c.OutputByName("out").Signals().IsEmpty())
}

func TestCombinators_WhenSkipsInsteadOfSuspending(t *testing.T) {
	var ran bool

	c := testutil.MustComponent("conditional",
		component.WithInputs("a", "b"),
		component.WithOutputs("out"),
		component.WithActivationFunc(component.When(
			component.HasSignalsOn("a", "b"),
			func(_ context.Context, this *component.Component) error {
				ran = true
				return this.OutputByName("out").PutSignals(signal.New("ran"))
			})))

	fm := testutil.MustFMesh("conditional")
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("a").PutSignals(signal.New(1)))

	_, err := fm.Run(context.Background())

	// When skips the activation and returns cleanly, so the mesh stops naturally
	// rather than waiting — the difference from RequireInputs.
	require.NoError(t, err)
	assert.False(t, ran)
}

func TestCombinators_PipelineTransformsSignalsThroughStages(t *testing.T) {
	double := func(signals *signal.Group) (*signal.Group, error) {
		return signals.MapPayloads(func(p any) any { return p.(int) * 2 }), nil
	}
	drop := func(signals *signal.Group) (*signal.Group, error) {
		return signals.Filter(func(s *signal.Signal) bool { return s.Payload().(int) < 20 }), nil
	}

	c := testutil.MustComponent("pipeline",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(
			component.Pipeline([]string{"in"}, "out", double, drop)))

	fm := testutil.MustFMesh("pipeline")
	require.NoError(t, fm.AddComponents(c))
	require.NoError(t, c.InputByName("in").PutSignals(
		signal.New(1), signal.New(5), signal.New(50)))

	_, err := fm.Run(context.Background())
	require.NoError(t, err)

	// 1,5,50 -> doubled 2,10,100 -> kept 2,10
	assert.Equal(t, []any{2, 10}, c.OutputByName("out").Signals().AllPayloads())
}

func TestNestedMesh_RunsInsideAComponent(t *testing.T) {
	// A component whose activation runs a mesh of its own. The inner run is
	// synchronous and its result is written to the outer component's output, so
	// the outer mesh sees one ordinary component.
	buildInner := func() (*fmesh.FMesh, *component.Component, *component.Component) {
		upper := testutil.MustComponent("upper",
			component.WithInputs("in"), component.WithOutputs("out"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				text, err := signal.As[string](this.InputByName("in").Signals().First())
				if err != nil {
					return err
				}
				return this.OutputByName("out").PutSignals(signal.New(strings.ToUpper(text)))
			}))
		exclaim := testutil.MustComponent("exclaim",
			component.WithInputs("in"), component.WithOutputs("out"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				text, err := signal.As[string](this.InputByName("in").Signals().First())
				if err != nil {
					return err
				}
				return this.OutputByName("out").PutSignals(signal.New(text + "!"))
			}))

		inner := testutil.MustFMesh("inner")
		require.NoError(t, inner.AddComponents(upper, exclaim))
		require.NoError(t, upper.OutputByName("out").PipeTo(exclaim.InputByName("in")))
		return inner, upper, exclaim
	}

	outerComponent := testutil.MustComponent("nested",
		component.WithInputs("in"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(ctx context.Context, this *component.Component) error {
			inner, innerIn, innerOut := buildInner()

			text, err := signal.As[string](this.InputByName("in").Signals().First())
			if err != nil {
				return err
			}
			if err = innerIn.InputByName("in").PutSignals(signal.New(text)); err != nil {
				return err
			}
			// The inner mesh inherits the outer run's context, so canceling the
			// outer run stops the inner one too.
			if _, err = inner.Run(ctx); err != nil {
				return err
			}

			result, err := innerOut.OutputByName("out").Signals().FirstPayload()
			if err != nil {
				return err
			}
			return this.OutputByName("out").PutSignals(signal.New(result))
		}))

	outer := testutil.MustFMesh("outer")
	require.NoError(t, outer.AddComponents(outerComponent))
	require.NoError(t, outerComponent.InputByName("in").PutSignals(signal.New("hello")))

	_, err := outer.Run(context.Background())
	require.NoError(t, err)

	payload, err := outerComponent.OutputByName("out").Signals().FirstPayload()
	require.NoError(t, err)
	assert.Equal(t, "HELLO!", payload)
}

func TestNestedMesh_InnerFailureSurfacesAsAnActivationError(t *testing.T) {
	inner := testutil.MustFMesh("inner") // empty: running it fails

	outerComponent := testutil.MustComponent("nested",
		component.WithInputs("in"),
		component.WithActivationFunc(func(ctx context.Context, _ *component.Component) error {
			_, err := inner.Run(ctx)
			return err
		}))

	outer := testutil.MustFMesh("outer")
	require.NoError(t, outer.AddComponents(outerComponent))
	require.NoError(t, outerComponent.InputByName("in").PutSignals(signal.New(1)))

	_, err := outer.Run(context.Background())

	require.ErrorIs(t, err, fmesh.ErrHitAnErrorOrPanic)
	assert.Contains(t, err.Error(), "no components found")
}
