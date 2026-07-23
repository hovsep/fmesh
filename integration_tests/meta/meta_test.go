package meta

import (
	"testing"

	"github.com/hovsep/fmesh/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

// Test_ScalarsOnSignals verifies that scalars can be attached to signals and
// aggregate methods on signal.Group work as expected.
func Test_ScalarsOnSignals(t *testing.T) {
	t.Run("scalars flow through a mesh and are aggregated at the sink", func(t *testing.T) {
		// Scenario: a sensor component emits temperature readings as signals with
		// a "temp" scalar. A monitor component collects them and checks the average.

		var collectedGroup *signal.Group

		sensor := testutil.MustComponent("sensor",
			component.WithInputs("trigger"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(this *component.Component) error {
				readings := []float64{36.6, 37.1, 38.2, 36.9}
				for _, r := range readings {
					sig := signal.New("reading").WithScalar("temp", r)
					if err := this.Outputs().ByName("out").PutSignals(sig); err != nil {
						return err
					}
				}
				return nil
			}),
		)

		monitor := testutil.MustComponent("monitor",
			component.WithInputs("in"),
			component.WithActivationFunc(func(this *component.Component) error {
				grp := this.Inputs().ByName("in").Signals()
				collectedGroup = grp
				return nil
			}),
		)

		fm := testutil.MustFMesh("temp-mesh")
		require.NoError(t, fm.AddComponents(sensor, monitor))
		require.NoError(t, sensor.Outputs().ByName("out").PipeTo(monitor.Inputs().ByName("in")))

		// Seed the sensor trigger to kick off activation
		require.NoError(t, sensor.Inputs().ByName("trigger").PutSignals(signal.New("go")))

		_, err := fm.Run()
		require.NoError(t, err)

		require.NotNil(t, collectedGroup)
		assert.Equal(t, 4, collectedGroup.Len())

		avg, err := collectedGroup.AvgScalar("temp")
		require.NoError(t, err)
		assert.InDelta(t, 37.2, avg, 0.01)

		minTemp, err := collectedGroup.MinScalar("temp")
		require.NoError(t, err)
		assert.InDelta(t, 36.6, minTemp, 1e-9)

		maxTemp, err := collectedGroup.MaxScalar("temp")
		require.NoError(t, err)
		assert.InDelta(t, 38.2, maxTemp, 1e-9)

		sum := collectedGroup.SumScalar("temp")
		assert.InDelta(t, 148.8, sum, 0.01)
	})

	t.Run("WithScalarOnEach stamps every signal in a group", func(t *testing.T) {
		grp := signal.NewGroup(1, 2, 3).WithScalarOnEach("priority", 5.0)
		assert.Equal(t, 3, grp.Len())

		signals := grp.All()
		for _, s := range signals {
			v, err := s.Scalars().Value("priority")
			require.NoError(t, err)
			assert.InDelta(t, 5.0, v, 1e-9)
		}
	})

	t.Run("RemoveScalarOnEach removes the scalar from every signal", func(t *testing.T) {
		grp := signal.NewGroup(1, 2).
			WithScalarOnEach("temp", 36.6).
			WithScalarOnEach("humidity", 0.55).
			RemoveScalarOnEach("humidity")

		signals := grp.All()
		for _, s := range signals {
			assert.True(t, s.Scalars().Has("temp"))
			assert.False(t, s.Scalars().Has("humidity"))
		}
	})

	t.Run("AvgScalar returns ok=false when no signal has the scalar", func(t *testing.T) {
		grp := signal.NewGroup(1, 2, 3)
		_, err := grp.AvgScalar("nonexistent")
		require.Error(t, err)
	})
}

// Test_ScalarsOnComponents verifies scalar metadata on components.
func Test_ScalarsOnComponents(t *testing.T) {
	t.Run("component scalars are independent of signal scalars", func(t *testing.T) {
		c := testutil.MustComponent("proc",
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithLabel("tier", "premium"),
			component.WithScalar("version", 2.0),
		)

		v, err := c.Scalars().Value("version")
		require.NoError(t, err)
		assert.InDelta(t, 2.0, v, 1e-9)

		// Signal scalars are separate
		sig := signal.New("data").WithScalar("weight", 1.5)
		sv, err := sig.Scalars().Value("weight")
		require.NoError(t, err)
		assert.InDelta(t, 1.5, sv, 1e-9)

		_, err = c.Scalars().Value("weight")
		require.Error(t, err, "component must not have signal's scalar")
	})
}

// Test_ScalarGroupMetadata verifies that a Group's OWN scalars are independent of its contents.
func Test_ScalarGroupMetadata(t *testing.T) {
	t.Run("group own scalar is separate from element scalars", func(t *testing.T) {
		grp := signal.NewGroup(1, 2, 3).
			WithScalar("batch_id", 42.0).
			WithScalarOnEach("temp", 37.0)

		// Group's own scalar
		v, err := grp.Scalars().Value("batch_id")
		require.NoError(t, err)
		assert.InDelta(t, 42.0, v, 1e-9)

		// Elements have "temp" but not "batch_id"
		signals := grp.All()
		for _, s := range signals {
			assert.True(t, s.Scalars().Has("temp"))
			assert.False(t, s.Scalars().Has("batch_id"))
		}
	})
}

// Test_PortScalarsWithOptions verifies the WithScalar port constructor option.
func Test_PortScalarsWithOptions(t *testing.T) {
	t.Run("port scalars set via constructor option", func(t *testing.T) {
		p := testutil.MustInputPort("sensor-in", port.WithScalar("sample_rate", 100.0))
		v, err := p.Scalars().Value("sample_rate")
		require.NoError(t, err)
		assert.InDelta(t, 100.0, v, 1e-9)
	})

	t.Run("port WithScalar mutating method", func(t *testing.T) {
		p := testutil.MustOutputPort("data-out").AddScalar("bandwidth", 1e6)
		v, err := p.Scalars().Value("bandwidth")
		require.NoError(t, err)
		assert.InDelta(t, 1e6, v, 1e-9)
	})
}
