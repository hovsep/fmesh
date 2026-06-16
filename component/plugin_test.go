package component

import (
	"testing"

	"github.com/hovsep/fmesh/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent_Plugin(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// Build component
		c, err := New("dummy",
			WithInputs("i1"),
			WithOutputs("o1"),
			WithDescription("Bypass int from i1 to o1"),
			WithActivationFunc(func(this *Component) error {
				i1, _ := this.InputByName("i1").Signals().FirstPayloadOrNil().(int)

				return this.OutputByName("o1").PutPayloads(i1)
			}),
			// Attach plugins
			WithPlugins(PricePlugin{}))

		require.NoError(t, err)
		require.NotNil(t, c)
		assert.True(t, c.OutputByName("o1").Signals().IsEmpty(), "no signals on output port before activation")
		assert.Equal(t, 2, c.Inputs().Len())
		assert.Equal(t, 2, c.Outputs().Len())

		// Pass inputs
		require.NoError(t, c.InputByName("i1").PutPayloads(1))
		require.NoError(t, c.InputByName("price_in").PutPayloads(122.333))

		// Activate component
		activationResult := c.MaybeActivate()

		assert.True(t, activationResult.activated)
		assert.False(t, activationResult.IsError())
		assert.False(t, activationResult.IsPanic())
		assert.InDelta(t, 77.5, c.State().Get("base_price"), 0.0001)
		assert.InDelta(t, 122.333, c.State().Get("new_price"), 0.0001)
		assert.False(t, c.OutputByName("price_out").Signals().IsEmpty())
		assert.InDelta(t, 1000.1, c.OutputByName("price_out").Signals().ReducePayloads(0.0, func(acc any, payload any) any {
			return acc.(float64) + payload.(float64)
		}), 0.0001)
		assert.True(t, c.Labels().ValueIs("plugin/price/version", "v1.2.4"))
		assert.True(t, c.Scalars().ValueIs("plugin/price/threshold", 105.54))
	})
}

type PricePlugin struct {
}

func (pp PricePlugin) GetName() string {
	return "PricePlugin"
}

func (pp PricePlugin) Init(c *Component) error {
	// Modify component interface (ports)
	priceIn, _ := port.NewInput("price_in", port.WithDescription("plugins can dynamically add ports"))
	priceOut, _ := port.NewOutput("price_out")
	_ = c.Inputs().Add(priceIn)
	_ = c.Outputs().Add(priceOut)

	// Plug in to component via hooks
	c.SetupHooks(func(hooks *Hooks) {
		// Mutate state
		hooks.OnCreation(func(this *Component) error {
			this.State().Set("base_price", 77.5)
			return nil
		})

		// Modify behavior (activation function)
		hooks.OnActivation(func(this *Component) error {
			if this.InputByName("price_in").HasSignals() {
				this.State().Upsert("new_price", func(old any) any {
					return this.InputByName("price_in").Signals().FirstPayloadOrDefault(0.0)
				})
			}
			return this.OutputByName("price_out").PutPayloads(999.0, 1.0, 0.1)
		})
	})

	// Modify metadata
	c.AddLabel("plugin/price/version", "v1.2.4")
	c.AddScalar("plugin/price/threshold", 105.54)
	return nil
}
