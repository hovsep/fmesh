package chainability

import (
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustComponent(name string, opts ...component.Option) *component.Component {
	c, err := component.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func mustInputPort(name string, opts ...port.Option) *port.Port {
	p, err := port.NewInput(name, opts...)
	if err != nil {
		panic(err)
	}
	return p
}

// TestChainability_CrossPackage verifies realistic cross-package chaining scenarios.
func TestChainability_CrossPackage(t *testing.T) {
	t.Run("full component setup", func(t *testing.T) {
		c := mustComponent("processor",
			component.WithInputs("in1", "in2"),
			component.WithOutputs("out1", "out2"),
		).
			SetDescription("main processor").
			AddLabel("env", "prod").
			AddLabel("tier", "backend").
			SetLabels(map[string]string{"reset": "true"}). // Reset all labels
			AddLabel("final", "label")

		assert.Equal(t, "processor", c.Name())
		assert.Equal(t, "main processor", c.Description())
		assert.Equal(t, 2, c.Labels().Len())
		assert.True(t, c.Labels().ValueIs("reset", "true"))
		assert.True(t, c.Labels().ValueIs("final", "label"))
		assert.False(t, c.Labels().Has("env"), "wiped by SetLabels")
		assert.Equal(t, 2, c.Inputs().Len())
		assert.Equal(t, 2, c.Outputs().Len())
	})

	t.Run("port with signals and labels", func(t *testing.T) {
		p := mustInputPort("data", port.WithDescription("data input")).
			AddLabel("type", "data")
		require.NoError(t, p.PutSignals(signal.New(1), signal.New(2)))
		p.AddLabel("count", "2")
		require.NoError(t, p.PutSignals(signal.New(3)))

		assert.Equal(t, "data", p.Name())
		assert.Equal(t, "data input", p.Description())
		assert.Equal(t, 2, p.Labels().Len())
		assert.Equal(t, 3, p.Signals().Len())
	})

	t.Run("signal with multiple label operations", func(t *testing.T) {
		s := signal.New("payload").
			WithLabel("source", "api").
			WithLabels(map[string]string{"priority": "high", "retry": "true"}).
			WithOnlyLabels(map[string]string{"final": "label"}) // Reset

		assert.Equal(t, 1, s.Labels().Len())
		assert.True(t, s.Labels().ValueIs("final", "label"))
		assert.False(t, s.Labels().Has("source"), "wiped by WithOnlyLabels")
	})

	t.Run("component with label cleanup", func(t *testing.T) {
		// Simulate component lifecycle: setup with debug labels, then clean them up
		c := mustComponent("worker",
			component.WithInputs("tasks"),
			component.WithOutputs("results"),
		).
			SetDescription("background worker").
			AddLabels(map[string]string{
				"env":      "prod",
				"team":     "backend",
				"debug":    "true",
				"trace-id": "abc123",
			}).
			RemoveLabels("debug", "trace-id") // Clean up temporary labels

		assert.Equal(t, 2, c.Labels().Len(), "should have only permanent labels")
		assert.True(t, c.Labels().Has("env"))
		assert.True(t, c.Labels().Has("team"))
		assert.False(t, c.Labels().Has("debug"), "debug label should be removed")
		assert.False(t, c.Labels().Has("trace-id"), "trace-id should be removed")
	})

	t.Run("port with label reset workflow", func(t *testing.T) {
		// Port initially configured with temporary setup labels, then cleared for production
		p := mustInputPort("input").
			AddLabels(map[string]string{
				"setup": "true",
				"test":  "mode",
				"debug": "enabled",
			})
		require.NoError(t, p.PutSignals(signal.New(1), signal.New(2)))
		p.ClearLabels(). // Clear all setup labels
					AddLabels(map[string]string{
				"required":  "true",
				"validated": "true",
			})

		assert.Equal(t, 2, p.Labels().Len())
		assert.False(t, p.Labels().Has("setup"), "setup labels cleared")
		assert.False(t, p.Labels().Has("test"), "test labels cleared")
		assert.False(t, p.Labels().Has("debug"), "debug labels cleared")
		assert.True(t, p.Labels().Has("required"), "production labels present")
		assert.True(t, p.Labels().Has("validated"), "production labels present")
		assert.True(t, p.IsInput(), "direction is built-in")
		assert.Equal(t, 2, p.Signals().Len(), "signals should remain")
	})

	t.Run("signal filtering and relabeling", func(t *testing.T) {
		// Signal with metadata that gets filtered and relabeled
		s := signal.New(map[string]any{"data": "value"}).
			WithLabels(map[string]string{
				"source":    "api",
				"priority":  "low",
				"timestamp": "2024-01-01",
				"temp":      "metadata",
			}).
			WithoutLabels("temp").         // Remove temporary metadata
			WithoutLabels("priority").     // Remove old priority
			WithLabel("priority", "high"). // Set new priority
			WithLabel("processed", "true")

		assert.Equal(t, 4, s.Labels().Len())
		assert.False(t, s.Labels().Has("temp"))
		assert.True(t, s.Labels().ValueIs("priority", "high"), "priority updated")
		assert.True(t, s.Labels().ValueIs("source", "api"), "source preserved")
		assert.True(t, s.Labels().ValueIs("timestamp", "2024-01-01"), "timestamp preserved")
		assert.True(t, s.Labels().ValueIs("processed", "true"), "processed added")
	})

	t.Run("complex label lifecycle", func(t *testing.T) {
		// Realistic scenario: component setup -> debug -> cleanup -> finalize
		c := mustComponent("api-handler",
			component.WithInputs("request"),
			component.WithOutputs("response", "errors"),
		).
			SetDescription("HTTP API handler").
			AddLabels(map[string]string{
				"env":  "dev",
				"team": "platform",
			}).
			AddLabels(map[string]string{ // Add debug labels
				"debug":    "true",
				"verbose":  "true",
				"trace-id": "xyz789",
			}).
			RemoveLabels("debug", "verbose", "trace-id"). // Remove all debug labels
			AddLabel("env", "prod").                      // Update env to prod
			AddLabel("deployed", "true")                  // Add deployment marker

		assert.Equal(t, 3, c.Labels().Len())
		assert.True(t, c.Labels().ValueIs("env", "prod"), "env updated to prod")
		assert.True(t, c.Labels().Has("team"))
		assert.True(t, c.Labels().Has("deployed"))
		assert.False(t, c.Labels().Has("debug"))
		assert.False(t, c.Labels().Has("verbose"))
		assert.False(t, c.Labels().Has("trace-id"))
	})

	t.Run("port and signal label coordination", func(t *testing.T) {
		// Port and signals with coordinated label management
		s1 := signal.New(1).
			WithLabels(map[string]string{"priority": "high", "source": "user"}).
			WithoutLabels("source").
			WithLabel("source", "validated")

		s2 := signal.New(2).
			WithLabels(map[string]string{"priority": "low", "source": "batch"}).
			WithNoLabels().
			WithLabels(map[string]string{"priority": "high", "source": "validated"})

		p := mustInputPort("validated-input").
			AddLabel("type", "input")
		require.NoError(t, p.PutSignals(s1, s2))
		p.AddLabel("count", "2").
			RemoveLabels("type") // Remove type

		assert.Equal(t, 2, p.Signals().Len())
		assert.Equal(t, 1, p.Labels().Len())
		assert.False(t, p.Labels().Has("type"))
		assert.True(t, p.Labels().Has("count"))
		assert.True(t, p.IsInput()) // Direction is built-in, not a label

		// Both signals should have consistent labels
		assert.True(t, s1.Labels().ValueIs("priority", "high"))
		assert.True(t, s1.Labels().ValueIs("source", "validated"))
		assert.True(t, s2.Labels().ValueIs("priority", "high"))
		assert.True(t, s2.Labels().ValueIs("source", "validated"))
	})
}
