package chainability

import (
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/labels"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
)

// TestChainability_CrossPackage verifies realistic cross-package chaining scenarios.
func TestChainability_CrossPackage(t *testing.T) {
	t.Run("full component setup", func(t *testing.T) {
		c := component.New("processor").
			WithDescription("main processor").
			AddLabel("env", "prod").
			AddLabel("tier", "backend").
			WithInputs("in1", "in2").
			WithOutputs("out1", "out2").
			SetLabels(labels.Map{"reset": "true"}). // Reset all labels
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
		p := port.New("data").
			WithDescription("data input").
			AddLabel("direction", "in").
			PutSignals(signal.New(1), signal.New(2)).
			AddLabel("count", "2").
			PutSignals(signal.New(3))

		assert.Equal(t, "data", p.Name())
		assert.Equal(t, "data input", p.Description())
		assert.Equal(t, 2, p.Labels().Len())
		assert.Equal(t, 3, p.Signals().Len())
	})

	t.Run("signal with multiple label operations", func(t *testing.T) {
		s := signal.New("payload").
			AddLabel("source", "api").
			AddLabels(labels.Map{"priority": "high", "retry": "true"}).
			SetLabels(labels.Map{"final": "label"}) // Reset

		assert.Equal(t, 1, s.Labels().Len())
		assert.True(t, s.Labels().ValueIs("final", "label"))
		assert.False(t, s.Labels().Has("source"), "wiped by SetLabels")
	})
}
