package meta

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/hovsep/fmesh/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

// Test_MetaOnSignals verifies that metadata can be attached to signals and
// batch-stamped across a group.
func Test_MetaOnSignals(t *testing.T) {
	t.Run("WithMetaOnEach stamps every signal in a group", func(t *testing.T) {
		grp := signal.NewGroup(1, 2, 3).WithMetaOnEach("priority", 5.0)
		assert.Equal(t, 3, grp.Len())

		signals := slices.Collect(grp.All())
		for _, s := range signals {
			v, err := s.Meta().Value[float64]("priority")
			require.NoError(t, err)
			assert.InDelta(t, 5.0, v, 1e-9)
		}
	})

	t.Run("WithoutMetaOnEach removes the entry from every signal", func(t *testing.T) {
		grp := signal.NewGroup(1, 2).
			WithMetaOnEach("temp", 36.6).
			WithMetaOnEach("humidity", 0.55).
			WithoutMetaOnEach("humidity")

		signals := slices.Collect(grp.All())
		for _, s := range signals {
			assert.True(t, s.Meta().Has("temp"))
			assert.False(t, s.Meta().Has("humidity"))
		}
	})
}

// Test_MetaOnComponents verifies metadata on components.
func Test_MetaOnComponents(t *testing.T) {
	t.Run("component metadata is independent of signal metadata", func(t *testing.T) {
		c := testutil.MustComponent("proc",
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithMeta("tier", "premium"),
			component.WithMeta("version", 2.0),
		)

		v, err := c.Meta().Value[float64]("version")
		require.NoError(t, err)
		assert.InDelta(t, 2.0, v, 1e-9)

		// Signal metadata is separate
		sig := signal.New("data").WithMeta("weight", 1.5)
		sv, err := sig.Meta().Value[float64]("weight")
		require.NoError(t, err)
		assert.InDelta(t, 1.5, sv, 1e-9)

		_, err = c.Meta().Value[float64]("weight")
		require.Error(t, err, "component must not have signal's entry")
	})
}

// Test_GroupOwnMeta verifies that a Group's OWN metadata is independent of its contents.
func Test_GroupOwnMeta(t *testing.T) {
	t.Run("group own entry is separate from element entries", func(t *testing.T) {
		grp := signal.NewGroup(1, 2, 3).
			WithMeta("batch_id", 42.0).
			WithMetaOnEach("temp", 37.0)

		// Group's own entry
		v, err := grp.Meta().Value[float64]("batch_id")
		require.NoError(t, err)
		assert.InDelta(t, 42.0, v, 1e-9)

		// Elements have "temp" but not "batch_id"
		signals := slices.Collect(grp.All())
		for _, s := range signals {
			assert.True(t, s.Meta().Has("temp"))
			assert.False(t, s.Meta().Has("batch_id"))
		}
	})
}

// Test_PortMetaWithOptions verifies the WithMeta port constructor option.
func Test_PortMetaWithOptions(t *testing.T) {
	t.Run("port metadata set via constructor option", func(t *testing.T) {
		p := testutil.MustInputPort("sensor-in", port.WithMeta("sample_rate", 100.0))
		v, err := p.Meta().Value[float64]("sample_rate")
		require.NoError(t, err)
		assert.InDelta(t, 100.0, v, 1e-9)
	})

	t.Run("port metadata set through the store", func(t *testing.T) {
		p := testutil.MustOutputPort("data-out")
		p.Meta().Set("bandwidth", 1e6)
		v, err := p.Meta().Value[float64]("bandwidth")
		require.NoError(t, err)
		assert.InDelta(t, 1e6, v, 1e-9)
	})
}

// Test_MetaFlowsThroughAMesh shows metadata surviving a mesh run and being
// rewritten en route. The store itself is unit-tested in the meta package.
func Test_MetaFlowsThroughAMesh(t *testing.T) {
	t.Run("transform metadata in mesh processing", func(t *testing.T) {
		// Scenario: Process signals and normalize their metadata keys during mesh execution
		normalizer := testutil.MustComponent("normalizer",
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
				inPort := this.InputByName("in")
				outPort := this.OutputByName("out")

				// Process each signal and normalize its keys
				for sig := range inPort.Signals().All() {
					// Normalize keys to lowercase
					newSignal := signal.New(sig.Payload())
					for _, k := range sig.Meta().Keys() {
						newSignal = newSignal.WithMeta(strings.ToLower(k), sig.Meta().ValueOrDefault(k, ""))
					}
					if err := outPort.PutSignals(newSignal); err != nil {
						return err
					}
				}
				return nil
			}),
		)

		fm := testutil.MustFMesh("meta-transform-mesh")
		require.NoError(t, fm.AddComponents(normalizer))

		// Input signal with mixed-case keys
		require.NoError(t, fm.ComponentByName("normalizer").InputByName("in").PutSignals(
			signal.New(100).WithMetaMany(map[string]string{
				"ENV":    "prod",
				"Region": "us-west",
				"TIER":   "frontend",
			}),
		))

		_, err := fm.Run(context.Background())
		require.NoError(t, err)

		outSignals := fm.ComponentByName("normalizer").OutputByName("out").Signals()
		assert.Equal(t, 1, outSignals.Len())

		firstSignal := outSignals.First()
		require.NotNil(t, firstSignal)

		assert.True(t, firstSignal.Meta().Has("env"))
		assert.True(t, firstSignal.Meta().Has("region"))
		assert.True(t, firstSignal.Meta().Has("tier"))
		assert.True(t, firstSignal.Meta().ValueIs("env", "prod"))
		assert.True(t, firstSignal.Meta().ValueIs("region", "us-west"))
	})
}
