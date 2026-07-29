package fmesh

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFMesh_Plugin(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		counter := &activationCounter{}

		fm, err := New("fm1", WithPlugins(counter))
		require.NoError(t, err)
		require.NotNil(t, fm)
		assert.True(t, fm.PluginRegistered("activationCounter"))
		assert.True(t, fm.Labels().ValueIs("plugin/counter/version", "v1"))

		// The plugin instruments whatever arrives, not what was there when it was
		// initialized -- a mesh is empty at construction time.
		require.NoError(t, fm.AddComponents(
			mustNewComponent("c1",
				component.WithInputs("i1"),
				component.WithOutputs("o1"),
				component.WithActivationFunc(func(this *component.Component) error {
					return this.OutputByName("o1").PutPayloads(1)
				})),
			mustNewComponent("c2",
				component.WithInputs("i1"),
				component.WithActivationFunc(func(*component.Component) error { return nil })),
		))
		assert.Equal(t, 2, counter.instrumented, "both components instrumented on arrival")

		mustPipeTo(fm.Components().ByName("c1").OutputByName("o1"),
			fm.Components().ByName("c2").InputByName("i1"))
		mustPutSignals(fm.Components().ByName("c1").InputByName("i1"), signal.New("go"))

		_, err = fm.Run()
		require.NoError(t, err)
		assert.Equal(t, 2, counter.activations, "every activation observed without any component knowing")
	})

	t.Run("plugins can be registered only once", func(t *testing.T) {
		fm, err := New("fm1", WithPlugins(&activationCounter{}, &activationCounter{}))

		require.Error(t, err)
		require.ErrorContains(t, err, "plugin activationCounter already registered")
		assert.Nil(t, fm)
	})

	t.Run("a failing plugin fails construction", func(t *testing.T) {
		fm, err := New("fm1", WithPlugins(brokenPlugin{}))

		require.Error(t, err)
		require.ErrorContains(t, err, `fmesh "fm1" plugin brokenPlugin initialization failed`)
		assert.Nil(t, fm)
	})

	t.Run("plugins initialize in name order", func(t *testing.T) {
		// Hooks fire in registration order, so plugin init order is observable
		// behavior rather than an implementation detail. Ranging the map would
		// make it vary between runs of the same binary.
		var order []string
		record := func(name string) *recordingPlugin {
			return &recordingPlugin{name: name, seen: &order}
		}

		fm, err := New("fm1", WithPlugins(record("zulu"), record("alpha"), record("mike")))
		require.NoError(t, err)
		require.NotNil(t, fm)

		assert.Equal(t, []string{"alpha", "mike", "zulu"}, order)
	})
}

// activationCounter is a mesh plugin that observes every component activation
// without any component being told about it.
type activationCounter struct {
	instrumented int
	activations  int
}

func (p *activationCounter) GetName() string { return "activationCounter" }

func (p *activationCounter) Init(fm *FMesh) error {
	fm.AddLabel("plugin/counter/version", "v1")

	fm.SetupHooks(func(hooks *Hooks) {
		hooks.OnComponentAdded(func(ctx *ComponentAddedContext) error {
			p.instrumented++
			ctx.Component.SetupHooks(func(ch *component.Hooks) {
				ch.AfterActivation(func(*component.ActivationContext) error {
					p.activations++
					return nil
				})
			})
			return nil
		})
	})
	return nil
}

type brokenPlugin struct{}

func (brokenPlugin) GetName() string   { return "brokenPlugin" }
func (brokenPlugin) Init(*FMesh) error { return errors.New("no") }

type recordingPlugin struct {
	name string
	seen *[]string
}

func (p *recordingPlugin) GetName() string { return p.name }

func (p *recordingPlugin) Init(*FMesh) error {
	*p.seen = append(*p.seen, p.name)
	return nil
}
