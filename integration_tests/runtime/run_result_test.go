package runtime

import (
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MultipleRun(t *testing.T) {
	t.Run("run result is initialized before each run", func(t *testing.T) {

		fm := fmesh.New("test fm").
			AddComponents(
				component.New("bypass").
					WithDescription("Bypasses all signals").
					AddInputs("in").
					AddOutputs("out").WithActivationFunc(func(this *component.Component) error {
					return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
				}))

		// Run the mesh in loop (typical simulation use-case)

		for i := 0; i < 5; i++ {
			fm.ComponentByName("bypass").InputByName("in").PutSignals(signal.New(i))
			runResult, err := fm.Run()
			require.NoError(t, err)
			assert.NotNil(t, runResult)
			// Only 2 cycles are expected per run
			assert.Equal(t, 2, runResult.Cycles.Len())
			// Only 1 cycle is expected to have activated components
			assert.Equal(t, 1, runResult.Cycles.CountMatch(func(c *cycle.Cycle) bool {
				return c.HasActivatedComponents()
			}))
		}

	})
}
