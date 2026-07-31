package component

import (
	"context"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComponent(t *testing.T) {
	type args struct {
		name string
		opts []Option
	}
	tests := []struct {
		name       string
		args       args
		assertions func(t *testing.T, c *Component, err error)
	}{
		{
			name: "empty name is valid",
			args: args{name: ""},
			assertions: func(t *testing.T, c *Component, err error) {
				require.NoError(t, err)
				assert.NotNil(t, c)
				assert.Empty(t, c.Name())
			},
		},
		{
			name: "with name",
			args: args{name: "multiplier"},
			assertions: func(t *testing.T, c *Component, err error) {
				require.NoError(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, "multiplier", c.Name())
			},
		},
		{
			name: "with custom logger",
			args: args{
				name: "with-logger",
				opts: []Option{
					WithLogger(log.New(io.Discard, "custom-prefix:", log.LstdFlags|log.Lmsgprefix)),
				},
			},
			assertions: func(t *testing.T, c *Component, err error) {
				require.NoError(t, err)
				assert.NotNil(t, c)
				assert.NotNil(t, c.Logger())
				assert.Equal(t, "custom-prefix:", c.Logger().Prefix())
			},
		},
		{
			name: "with onCreation hook",
			args: args{
				name: "with-hook",
				opts: []Option{
					WithHooks(func(hooks *Hooks) {
						hooks.OnCreation(func(_ context.Context, component *Component) error {
							component.Labels().Set("tagging-source", "hook")
							return nil
						})
					}),
				},
			},
			assertions: func(t *testing.T, c *Component, err error) {
				require.NoError(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, "hook", c.Labels().ValueOrDefault("tagging-source", "default"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.args.name, tt.args.opts...)
			if tt.assertions != nil {
				tt.assertions(t, c, err)
			}
		})
	}
}

func TestComponent_WithDescription(t *testing.T) {
	t.Run("sets description via option", func(t *testing.T) {
		c := mustNew("c1", WithDescription("descr"))
		assert.Equal(t, "descr", c.Description())
	})

	t.Run("empty by default", func(t *testing.T) {
		c := mustNew("c1")
		assert.Empty(t, c.Description())
	})

	t.Run("WithDescription replaces previous value", func(t *testing.T) {
		c := mustNew("c1", WithDescription("first"), WithDescription("second"))
		assert.Equal(t, "second", c.Description())
	})
}

// Components hold metadata in a live store; the store's own behavior is covered
// in the meta package, so this only checks the wiring.
func TestComponent_Metadata(t *testing.T) {
	t.Run("Labels returns the live store", func(t *testing.T) {
		c := mustNew("c1")
		c.Labels().Set("env", "prod").SetMany(map[string]string{"tier": "api"})

		assert.Equal(t, 2, c.Labels().Len())
		assert.True(t, c.Labels().ValueIs("env", "prod"))

		c.Labels().Remove("env")
		assert.False(t, c.Labels().Has("env"))

		c.Labels().Clear()
		assert.Zero(t, c.Labels().Len())
	})

	t.Run("Scalars returns the live store", func(t *testing.T) {
		c := mustNew("c1")
		c.Scalars().Set("weight", 1.5)

		assert.True(t, c.Scalars().ValueIs("weight", 1.5))

		c.Scalars().Clear()
		assert.Zero(t, c.Scalars().Len())
	})

	t.Run("constructor options seed both stores", func(t *testing.T) {
		c := mustNew("c1", WithLabel("env", "prod"), WithScalar("weight", 2))

		assert.True(t, c.Labels().ValueIs("env", "prod"))
		assert.True(t, c.Scalars().ValueIs("weight", 2))
	})

	t.Run("stores are per component", func(t *testing.T) {
		c1, c2 := mustNew("c1"), mustNew("c2")
		c1.Labels().Set("only", "c1")

		assert.False(t, c2.Labels().Has("only"))
	})
}

func TestComponent_Chainability(t *testing.T) {
	t.Run("Clear+SetMany called twice replaces all labels", func(t *testing.T) {
		c := mustNew("c1")
		c.Labels().Clear().SetMany(map[string]string{"k1": "v1", "k2": "v2"})
		c.Labels().Clear().SetMany(map[string]string{"k3": "v3"})

		assert.Equal(t, 1, c.Labels().Len())
		assert.False(t, c.Labels().Has("k1"), "k1 should be replaced")
		assert.False(t, c.Labels().Has("k2"), "k2 should be replaced")
		assert.True(t, c.Labels().ValueIs("k3", "v3"))
	})

	t.Run("AddLabels called twice merges labels", func(t *testing.T) {
		c := mustNew("c1")
		c.Labels().SetMany(map[string]string{"k1": "v1", "k2": "v2"})
		c.Labels().SetMany(map[string]string{"k3": "v3", "k2": "v2-updated"})

		assert.Equal(t, 3, c.Labels().Len())
		assert.True(t, c.Labels().ValueIs("k1", "v1"))
		assert.True(t, c.Labels().ValueIs("k2", "v2-updated"), "should update existing key")
		assert.True(t, c.Labels().ValueIs("k3", "v3"))
	})

	t.Run("mixed Set and Add operations", func(t *testing.T) {
		c := mustNew("c1")
		c.Labels().
			Set("k1", "v1").
			SetMany(map[string]string{"k2": "v2", "k3": "v3"}).
			Clear().SetMany(map[string]string{"k4": "v4"}). // Wipes k1, k2, k3
			Set("k5", "v5")                                 // Merges with k4

		assert.Equal(t, 2, c.Labels().Len())
		assert.False(t, c.Labels().Has("k1"), "wiped by Clear")
		assert.False(t, c.Labels().Has("k2"), "wiped by Clear")
		assert.False(t, c.Labels().Has("k3"), "wiped by Clear")
		assert.True(t, c.Labels().ValueIs("k4", "v4"))
		assert.True(t, c.Labels().ValueIs("k5", "v5"))
	})

	t.Run("AddInputs called twice adds ports without duplicates", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddInputs("in1", "in2"))
		require.NoError(t, c.AddInputs("in3")) // in1 already exists - would error, skip in1
		_ = c.AddInputs("in1")                 // duplicate - ignore error

		assert.Equal(t, 3, c.Inputs().Len())
		assert.NotNil(t, c.Inputs().ByName("in1"))
		assert.NotNil(t, c.Inputs().ByName("in2"))
		assert.NotNil(t, c.Inputs().ByName("in3"))
	})

	t.Run("AddOutputs called twice adds ports without duplicates", func(t *testing.T) {
		c := mustNew("c1")
		require.NoError(t, c.AddOutputs("out1", "out2"))
		require.NoError(t, c.AddOutputs("out3"))
		_ = c.AddOutputs("out1") // duplicate - ignore error

		assert.Equal(t, 3, c.Outputs().Len())
		assert.NotNil(t, c.Outputs().ByName("out1"))
		assert.NotNil(t, c.Outputs().ByName("out2"))
		assert.NotNil(t, c.Outputs().ByName("out3"))
	})

	t.Run("ClearLabels removes all labels", func(t *testing.T) {
		c := mustNew("c1")
		c.Labels().
			SetMany(map[string]string{"k1": "v1", "k2": "v2"}).
			Clear().
			Set("k3", "v3")

		assert.Equal(t, 1, c.Labels().Len())
		assert.False(t, c.Labels().Has("k1"))
		assert.False(t, c.Labels().Has("k2"))
		assert.True(t, c.Labels().ValueIs("k3", "v3"))
	})

	t.Run("RemoveLabels removes specific labels", func(t *testing.T) {
		c := mustNew("c1")
		c.Labels().
			SetMany(map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}).
			Remove("k1", "k2").
			Set("k4", "v4")

		assert.Equal(t, 2, c.Labels().Len())
		assert.False(t, c.Labels().Has("k1"))
		assert.False(t, c.Labels().Has("k2"))
		assert.True(t, c.Labels().ValueIs("k3", "v3"))
		assert.True(t, c.Labels().ValueIs("k4", "v4"))
	})
}
