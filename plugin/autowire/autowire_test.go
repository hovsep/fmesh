package autowire

import (
	"context"
	"errors"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutowire(t *testing.T) {
	t.Run("wires by convention regardless of the order components arrive", func(t *testing.T) {
		// The consumer is added first, so it can only be wired by the second
		// component's arrival -- which is the half that a one-directional
		// implementation silently gets wrong.
		fm, err := fmesh.New("m", fmesh.WithPlugins(Prefixed("env_")))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			testutil.MustComponent("body",
				component.WithInputs("env_sun_uvi"),
				component.WithActivationFunc(recordArrival("env_sun_uvi"))),
		))
		require.NoError(t, fm.AddComponents(
			testutil.MustComponent("sun",
				component.WithInputs("i1"),
				component.WithOutputs("uvi"),
				component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
					return this.OutputByName("uvi").PutPayloads(7)
				})),
		))

		require.NoError(t, fm.Components().ByName("sun").InputByName("i1").
			PutSignals(signal.New("go")))
		_, err = fm.Run(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 7, fm.Components().ByName("body").State().Get("arrived"),
			"the sun reached the body with no wiring written")
	})

	t.Run("Broadcast feeds every same-named input", func(t *testing.T) {
		fm, err := fmesh.New("m", fmesh.WithPlugins(Broadcast("time")))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			testutil.MustComponent("clock",
				component.WithInputs("i1"),
				component.WithOutputs("time"),
				component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
					return this.OutputByName("time").PutPayloads(1)
				})),
			testutil.MustComponent("heart",
				component.WithInputs("time"),
				component.WithActivationFunc(recordArrival("time"))),
			testutil.MustComponent("lung",
				component.WithInputs("time"),
				component.WithActivationFunc(recordArrival("time"))),
		))

		require.NoError(t, fm.Components().ByName("clock").InputByName("i1").
			PutSignals(signal.New("go")))
		_, err = fm.Run(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 1, fm.Components().ByName("heart").State().Get("arrived"))
		assert.Equal(t, 1, fm.Components().ByName("lung").State().Get("arrived"))
	})

	t.Run("declining to name a port wires nothing", func(t *testing.T) {
		fm, err := fmesh.New("m", fmesh.WithPlugins(Broadcast("time")))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			testutil.MustComponent("clock",
				component.WithOutputs("not_time"),
				component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })),
			testutil.MustComponent("heart",
				component.WithInputs("time"),
				component.WithActivationFunc(recordArrival("time"))),
		))

		assert.Empty(t, fm.Components().ByName("clock").OutputByName("not_time").Pipes().Len(),
			"a declined port is left unwired")
	})

	t.Run("a plugin with no naming rule fails construction", func(t *testing.T) {
		fm, err := fmesh.New("m", fmesh.WithPlugins(&Plugin{}))

		require.Error(t, err)
		require.ErrorContains(t, err, "Name must be set")
		assert.Nil(t, fm)
	})

	t.Run("naming an input the destination does not have wires nothing", func(t *testing.T) {
		// Distinct from declining: the rule does name a port, there is just no
		// such input on this component.
		fm, err := fmesh.New("m", fmesh.WithPlugins(Broadcast("time")))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			testutil.MustComponent("clock",
				component.WithOutputs("time"),
				component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })),
			testutil.MustComponent("deaf",
				component.WithInputs("not_time"),
				component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })),
		))

		assert.Zero(t, fm.Components().ByName("clock").OutputByName("time").Pipes().Len())
	})

	t.Run("two conventions agreeing on a pair wire it once", func(t *testing.T) {
		// Pipes are not deduplicated and a port flushes once per pipe, so a
		// second identical pipe would deliver every signal twice.
		everythingIsTime := &Plugin{
			PluginName: "autowire:everything-is-time",
			Name:       func(*component.Component, *port.Port) string { return "time" },
		}

		fm, err := fmesh.New("m", fmesh.WithPlugins(Broadcast("time"), everythingIsTime))
		require.NoError(t, err)

		require.NoError(t, fm.AddComponents(
			testutil.MustComponent("clock",
				component.WithInputs("i1"),
				component.WithOutputs("time"),
				component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
					return this.OutputByName("time").PutPayloads(1)
				})),
			testutil.MustComponent("heart",
				component.WithInputs("time"),
				component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
					this.State().Set("received", this.InputByName("time").Signals().Len())
					return nil
				})),
		))

		assert.Equal(t, 1, fm.Components().ByName("clock").OutputByName("time").Pipes().Len())

		require.NoError(t, fm.Components().ByName("clock").InputByName("i1").
			PutSignals(signal.New("go")))
		_, err = fm.Run(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 1, fm.Components().ByName("heart").State().Get("received"),
			"the tick arrived once, not once per convention")
	})

	t.Run("a failing pipe fails AddComponents", func(t *testing.T) {
		fm, err := fmesh.New("m", fmesh.WithPlugins(Broadcast("time")))
		require.NoError(t, err)

		heart := testutil.MustComponent("heart",
			component.WithInputs("time"),
			component.WithActivationFunc(func(context.Context, *component.Component) error { return nil }))
		heart.InputByName("time").SetupHooks(func(hooks *port.Hooks) {
			hooks.OnInboundPipe(func(context.Context, *port.InboundPipeContext) error {
				return errors.New("refused")
			})
		})

		err = fm.AddComponents(
			testutil.MustComponent("clock",
				component.WithOutputs("time"),
				component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })),
			heart,
		)

		require.Error(t, err)
		assert.ErrorContains(t, err, "refused")
	})
}

// recordArrival notes what turned up on a port, since the mesh drains ports once
// a cycle is done and nothing is left to inspect after Run returns.
func recordArrival(portName string) component.ActivationFunc {
	return func(_ context.Context, this *component.Component) error {
		this.State().Set("arrived", this.InputByName(portName).Signals().FirstPayloadOrNil())
		return nil
	}
}
