package component

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponent_WithInitialState(t *testing.T) {
	tests := []struct {
		name      string
		component *Component
		wantState State
	}{
		{
			name:      "no initial state",
			component: New("c1"),
			wantState: NewState(),
		},
		{
			name: "with initial state",
			component: New("c1").WithInitialState(func(state State) {
				state.Set("battery", 100.00)
				state.Set("speed", 200)
				state.Set("data", []byte{1, 2, 3})
				state.Set("secret", "LEON")
			}),
			wantState: State{
				"battery": 100.00,
				"speed":   200,
				"data":    []byte{1, 2, 3},
				"secret":  "LEON",
			},
		},
		{
			name:      "with nil state initializer",
			component: New("c1").WithInitialState(nil),
			wantState: NewState(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantState, tt.component.State())
		})
	}
}

func Test_StateManipulations(t *testing.T) {
	t.Run("Has", func(t *testing.T) {
		c := New("c1")
		assert.False(t, c.State().Has("name"))

		c.WithInitialState(func(state State) {
			state.Set("name", "Leon")
		})
		assert.True(t, c.State().Has("name"))
	})

	t.Run("Get", func(t *testing.T) {
		c := New("c1")
		assert.Nil(t, c.State().Get("name"))

		c.WithInitialState(func(state State) {
			state.Set("name", "Leon")
		})
		assert.Equal(t, "Leon", c.State().Get("name"))
		assert.Equal(t, "Vita", c.State().GetOrDefault("non-existent", "Vita"))
	})

	t.Run("Delete", func(t *testing.T) {
		c := New("c1")
		assert.Empty(t, c.State())
		c.State().Delete("non-existent")
		assert.Empty(t, c.State())

		c.WithInitialState(func(state State) {
			state.Set("name", "Leon")
			state.Set("fruit", "banana")
		})
		c.State().Delete("name")
		assert.False(t, c.State().Has("name"))
		assert.True(t, c.State().Has("fruit"))
	})

	t.Run("Reset", func(t *testing.T) {
		c := New("c1").
			WithInitialState(func(state State) {
				state.Set("name", "Leon")
				state.Set("fruit", "banana")
			})
		assert.Len(t, c.State(), 2)

		c.ResetState()
		assert.Empty(t, c.State())
	})
}
