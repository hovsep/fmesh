package labels

import (
	"strings"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_LabelTransformation demonstrates real-world use cases for labels.Collection.Map().
func Test_LabelTransformation(t *testing.T) {
	t.Run("normalize label keys to uppercase", func(t *testing.T) {
		// Scenario: You receive signals with mixed-case labels and want to normalize them
		sig := signal.New(100).WithLabels(labels.Map{
			"env":    "production",
			"Region": "us-east",
			"Tier":   "backend",
		})

		// Normalize all label keys to uppercase for consistency
		normalizedLabels := sig.Labels().Map(func(k, v string) (string, string) {
			return strings.ToUpper(k), v
		})

		assert.Equal(t, 3, normalizedLabels.Len())
		assert.True(t, normalizedLabels.Has("ENV"))
		assert.True(t, normalizedLabels.Has("REGION"))
		assert.True(t, normalizedLabels.Has("TIER"))
		assert.True(t, normalizedLabels.ValueIs("ENV", "production"))
	})

	t.Run("add namespace prefix to labels", func(t *testing.T) {
		// Scenario: You want to namespace all labels to avoid conflicts
		c := component.New("processor").
			WithInputs("in").
			WithOutputs("out").
			WithLabels(labels.Map{
				"version": "1.0",
				"type":    "transformer",
			})

		// Add application-specific namespace to all labels
		namespacedLabels := c.Labels().Map(func(k, v string) (string, string) {
			return "app." + k, v
		})

		assert.True(t, namespacedLabels.Has("app.version"))
		assert.True(t, namespacedLabels.Has("app.type"))
		assert.True(t, namespacedLabels.ValueIs("app.version", "1.0"))
	})

	t.Run("add prefix to label values", func(t *testing.T) {
		// Scenario: You want to mark all label values as coming from a specific source
		sig := signal.New("data").WithLabels(labels.Map{
			"source": "api",
			"format": "json",
		})

		// Add source prefix to all values
		prefixedLabels := sig.Labels().Map(func(k, v string) (string, string) {
			return k, "external:" + v
		})

		assert.True(t, prefixedLabels.ValueIs("source", "external:api"))
		assert.True(t, prefixedLabels.ValueIs("format", "external:json"))
	})

	t.Run("convert labels to debug format", func(t *testing.T) {
		// Scenario: You want to wrap all label values in brackets for debugging
		sig := signal.New(42).WithLabels(labels.Map{
			"id":   "123",
			"name": "test",
		})

		// Wrap values for debugging
		debugLabels := sig.Labels().Map(func(k, v string) (string, string) {
			return "[DEBUG:" + k + "]", "[" + v + "]"
		})

		assert.True(t, debugLabels.Has("[DEBUG:id]"))
		assert.True(t, debugLabels.Has("[DEBUG:name]"))
		assert.True(t, debugLabels.ValueIs("[DEBUG:id]", "[123]"))
	})

	t.Run("transform labels in mesh processing", func(t *testing.T) {
		// Scenario: Process signals and normalize their labels during mesh execution
		normalizer := component.New("normalizer").
			WithInputs("in").
			WithOutputs("out").
			WithActivationFunc(func(this *component.Component) error {
				inPort := this.InputByName("in")
				outPort := this.OutputByName("out")

				// Process each signal and normalize its labels
				inPort.Signals().ForEach(func(sig *signal.Signal) {
					// Normalize label keys to lowercase
					normalizedLabels := sig.Labels().Map(func(k, v string) (string, string) {
						return strings.ToLower(k), v
					})

					// Create new signal with normalized labels
					labelsMap, err := normalizedLabels.All()
					if err == nil {
						newSignal := signal.New(sig.PayloadOrNil()).WithLabels(labelsMap)
						outPort.PutSignals(newSignal)
					}
				})

				return nil
			})

		fm := fmesh.New("label-transform-mesh").WithComponents(normalizer)

		// Input signal with mixed-case labels
		fm.ComponentByName("normalizer").InputByName("in").PutSignals(
			signal.New(100).WithLabels(labels.Map{
				"ENV":    "prod",
				"Region": "us-west",
				"TIER":   "frontend",
			}),
		)

		_, err := fm.Run()
		require.NoError(t, err)

		// Check that output signal has normalized labels
		outSignals := fm.ComponentByName("normalizer").OutputByName("out").Signals()
		assert.Equal(t, 1, outSignals.Len())

		firstSignal := outSignals.FirstOrNil()
		require.NotNil(t, firstSignal)

		assert.True(t, firstSignal.Labels().Has("env"))
		assert.True(t, firstSignal.Labels().Has("region"))
		assert.True(t, firstSignal.Labels().Has("tier"))
		assert.True(t, firstSignal.Labels().ValueIs("env", "prod"))
		assert.True(t, firstSignal.Labels().ValueIs("region", "us-west"))
	})

	t.Run("chain Map with Filter", func(t *testing.T) {
		// Scenario: Transform labels and then filter them
		sig := signal.New("data").WithLabels(labels.Map{
			"app.version": "1.0",
			"app.type":    "service",
			"sys.cpu":     "high",
			"sys.mem":     "low",
		})

		// First, uppercase all keys, then filter to only keep "app." prefixed ones
		transformed := sig.Labels().
			Map(func(k, v string) (string, string) {
				return strings.ToUpper(k), strings.ToUpper(v)
			}).
			Filter(func(k, v string) bool {
				return strings.HasPrefix(k, "APP.")
			})

		assert.Equal(t, 2, transformed.Len())
		assert.True(t, transformed.Has("APP.VERSION"))
		assert.True(t, transformed.Has("APP.TYPE"))
		assert.True(t, transformed.ValueIs("APP.VERSION", "1.0"))
		assert.True(t, transformed.ValueIs("APP.TYPE", "SERVICE"))
	})
}
