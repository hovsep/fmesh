package component

import (
	"errors"
	"testing"

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
		assert.True(t, c.PluginRegistered("PricePlugin"))
		assert.False(t, c.PluginRegistered("nope"))
	})

	t.Run("plugins can be registered only once", func(t *testing.T) {
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
			WithPlugins(PricePlugin{}, PricePlugin{}))

		require.Error(t, err)
		require.ErrorContains(t, err, "plugin PricePlugin already registered")
		assert.Nil(t, c)
	})

	t.Run("plugins initialize in name order", func(t *testing.T) {
		// Same reasoning as at mesh level: plugins register hooks, hooks fire in
		// registration order, so init order is observable behavior.
		var order []string
		record := func(name string) recordingPlugin {
			return recordingPlugin{name: name, seen: &order}
		}

		c, err := New("dummy", WithPlugins(record("zulu"), record("alpha"), record("mike")))

		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, []string{"alpha", "mike", "zulu"}, order)
	})

	t.Run("a failing plugin fails construction", func(t *testing.T) {
		c, err := New("dummy", WithPlugins(brokenPlugin{}))

		require.Error(t, err)
		require.ErrorContains(t, err, `component "dummy" plugin brokenPlugin initialization failed`)
		assert.Nil(t, c)
	})
}

type recordingPlugin struct {
	name string
	seen *[]string
}

func (p recordingPlugin) GetName() string { return p.name }

func (p recordingPlugin) Init(*Component) error {
	*p.seen = append(*p.seen, p.name)
	return nil
}

type brokenPlugin struct{}

func (brokenPlugin) GetName() string       { return "brokenPlugin" }
func (brokenPlugin) Init(*Component) error { return errors.New("no") }

type PricePlugin struct {
}

func (pp PricePlugin) GetName() string {
	return "PricePlugin"
}

func (pp PricePlugin) Init(c *Component) error {
	// Modify component interface (ports)
	_ = c.AddInputs("price_in")
	_ = c.AddOutputs("price_out")

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
