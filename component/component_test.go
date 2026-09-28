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
							component.Meta().Set("tagging-source", "hook")
							return nil
						})
					}),
				},
			},
			assertions: func(t *testing.T, c *Component, err error) {
				require.NoError(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, "hook", c.Meta().ValueOrDefault("tagging-source", "default"))
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
	t.Run("Meta returns the live store", func(t *testing.T) {
		c := mustNew("c1")
		c.Meta().Set("env", "prod").Set("weight", 1.5)

		assert.Equal(t, 2, c.Meta().Len())
		assert.True(t, c.Meta().ValueIs("env", "prod"))
		assert.True(t, c.Meta().ValueIs("weight", 1.5))

		c.Meta().Clear()
		assert.Zero(t, c.Meta().Len())
	})

	t.Run("constructor options seed both value types", func(t *testing.T) {
		c := mustNew("c1", WithMeta("env", "prod"), WithMeta("weight", 2.0))

		assert.True(t, c.Meta().ValueIs("env", "prod"))
		assert.True(t, c.Meta().ValueIs("weight", 2.0))
	})

	t.Run("stores are per component", func(t *testing.T) {
		c1, c2 := mustNew("c1"), mustNew("c2")
		c1.Meta().Set("only", "c1")

		assert.False(t, c2.Meta().Has("only"))
	})
}

func TestComponent_Chainability(t *testing.T) {
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
}
