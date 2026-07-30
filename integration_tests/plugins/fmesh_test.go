// Package plugins exercises the bundled mesh plugins together on one mesh.
package plugins

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/plugin"
	"github.com/hovsep/fmesh/signal"
)

// A small habitat: a clock drives everything that keeps time, two environmental
// factors publish readings, and a body consumes all three.
//
// The point of the test is what is absent. Four components, five connections,
// and not one PipeTo between components -- two naming conventions derive the
// whole graph as the components arrive, and a profiler measures the result.
// Adding a fifth component that declares an input named "time" would wire it to
// the clock with no edit anywhere.
func TestPlugins_ConventionWiredMesh(t *testing.T) {
	profiler := plugin.NewProfiler()

	fm := testutil.MustFMesh("habitat", fmesh.WithPlugins(
		profiler,
		// Everything that declared an input called "time" hears the clock.
		plugin.AutowireBroadcastAs("tick", "time"),
		// Everything that asked for a factor by name gets it:
		// sun's "uvi" output reaches the input named "env_sun_uvi".
		plugin.AutowirePrefixed("env_"),
	))

	assert.True(t, fm.PluginRegistered("profiler"))
	assert.True(t, fm.PluginRegistered("autowire:broadcast:tick->time"))
	assert.True(t, fm.PluginRegistered("autowire:prefixed:env_"))

	// The clock is the only component that wires anything by hand, and only to
	// itself: a loopback is not something a naming convention can express.
	clock := testutil.MustComponent("clock",
		component.WithInputs("i1"),
		component.WithOutputs("tick", "beat"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			left, _ := this.State().GetOrDefault("left", 3).(int)
			this.State().Set("left", left-1)

			if err := this.OutputByName("tick").PutPayloads(left); err != nil {
				return err
			}
			if left-1 > 0 {
				return this.OutputByName("beat").PutPayloads(left - 1)
			}
			return nil
		}))
	require.NoError(t, clock.LoopbackPipe("beat", "i1"))

	sun := testutil.MustComponent("sun",
		component.WithInputs("time"),
		component.WithOutputs("uvi"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("uvi").PutPayloads(7)
		}))

	air := testutil.MustComponent("air",
		component.WithInputs("time"),
		component.WithOutputs("gas"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("gas").PutPayloads(21)
		}))

	body := testutil.MustComponent("body",
		component.WithInputs("time", "env_sun_uvi", "env_air_gas"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			this.State().Upsert("uvi", func(old any) any {
				prev, _ := old.(int)
				return prev + sumInts(this, "env_sun_uvi")
			})
			this.State().Upsert("gas", func(old any) any {
				prev, _ := old.(int)
				return prev + sumInts(this, "env_air_gas")
			})
			this.State().Upsert("ticks", func(old any) any {
				prev, _ := old.(int)
				return prev + this.InputByName("time").Signals().Len()
			})
			return nil
		}))

	// The body is added before anything it listens to, so most of its wiring can
	// only happen on the later arrivals. Order is not supposed to matter.
	require.NoError(t, fm.AddComponents(body))
	require.NoError(t, fm.AddComponents(sun, air, clock))

	t.Run("the conventions derived the graph", func(t *testing.T) {
		assert.Equal(t, 3, clock.OutputByName("tick").Pipes().Len(),
			"the tick reaches sun, air and body -- but not the clock itself")
		assert.Equal(t, 1, sun.OutputByName("uvi").Pipes().Len())
		assert.Equal(t, 1, air.OutputByName("gas").Pipes().Len())
		assert.Equal(t, 1, clock.OutputByName("beat").Pipes().Len(),
			"only the hand-written loopback")
	})

	testutil.MustPutSignals(clock.InputByName("i1"), signal.New("start"))
	_, err := fm.Run(context.Background())
	require.NoError(t, err)

	t.Run("signals flowed over the derived pipes", func(t *testing.T) {
		assert.Equal(t, 3, body.State().Get("ticks"), "three ticks reached the body")
		// The sun and air publish on every tick they hear, so all three readings
		// arrive -- each one cycle behind the tick that produced it, since the
		// clock reaches the body and the factors in the same cycle.
		assert.Equal(t, 3*7, body.State().Get("uvi"))
		assert.Equal(t, 3*21, body.State().Get("gas"))
	})

	t.Run("the profiler measured every component", func(t *testing.T) {
		assert.Equal(t, 1, profiler.Runs().Count)
		assert.Positive(t, profiler.Cycles().Count)

		measured := make(map[string]int)
		for _, s := range profiler.Components() {
			measured[s.Component] = s.Count
			assert.Positive(t, s.Total, "%s was timed", s.Component)
			assert.LessOrEqual(t, s.Min, s.Max)
		}
		assert.Equal(t, map[string]int{"clock": 3, "sun": 3, "air": 3, "body": 4}, measured,
			"no component opted in to being measured")

		require.NotEmpty(t, profiler.TopN(1))
		assert.Equal(t, "body", profiler.TopN(1)[0].Component,
			"the body activates once more than the rest: a final pass on the last readings")

		assert.Contains(t, profiler.Report(), "clock")
	})

	t.Run("a second run pools into the same profile", func(t *testing.T) {
		clock.State().Delete("left")
		testutil.MustPutSignals(clock.InputByName("i1"), signal.New("start"))
		_, err := fm.Run(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 2, profiler.Runs().Count)

		profiler.Reset()
		assert.Zero(t, profiler.Runs().Count)
		assert.Empty(t, profiler.Components())
	})
}

// sumInts adds up the int payloads waiting on a port.
func sumInts(c *component.Component, portName string) int {
	sum, _ := c.InputByName(portName).Signals().ReducePayloads(0, func(acc, payload any) any {
		a, _ := acc.(int)
		v, _ := payload.(int)
		return a + v
	}).(int)
	return sum
}
