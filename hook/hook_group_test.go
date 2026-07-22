package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookGroup_ChainableAPI(t *testing.T) {
	t.Run("Add returns self for chaining", func(t *testing.T) {
		hg := NewGroup[int]()

		result := hg.Add(func(i int) error {
			return nil
		}).
			Add(func(i int) error {
				return nil
			}).
			Add(func(i int) error {
				return nil
			})

		assert.Equal(t, hg, result)
		assert.Equal(t, 3, hg.Len())
	})
}

func TestHookGroup_BasicFunctionality(t *testing.T) {
	t.Run("Add and Trigger hooks in order", func(t *testing.T) {
		var log []int
		hg := NewGroup[int]()

		hg.Add(func(i int) error {
			log = append(log, i*1)
			return nil
		})
		hg.Add(func(i int) error {
			log = append(log, i*2)
			return nil
		})
		hg.Add(func(i int) error {
			log = append(log, i*3)
			return nil
		})

		err := hg.Trigger(10)
		require.NoError(t, err)
		assert.Equal(t, []int{10, 20, 30}, log)
	})

	t.Run("IsEmpty returns true for new group", func(t *testing.T) {
		hg := NewGroup[string]()

		assert.True(t, hg.IsEmpty())
		assert.Equal(t, 0, hg.Len())
	})

	t.Run("IsEmpty returns false after adding hooks", func(t *testing.T) {
		hg := NewGroup[string]()
		hg.Add(func(s string) error { return nil })

		assert.False(t, hg.IsEmpty())
		assert.Equal(t, 1, hg.Len())
	})

	t.Run("Generic type works with custom structs", func(t *testing.T) {
		type TestStruct struct {
			Value int
		}

		var captured int
		hg := NewGroup[*TestStruct]()
		hg.Add(func(ts *TestStruct) error {
			captured = ts.Value
			return nil
		})

		err := hg.Trigger(&TestStruct{Value: 42})
		require.NoError(t, err)
		assert.Equal(t, 42, captured)
	})
}

func TestHookGroup_EdgeCases(t *testing.T) {
	t.Run("Trigger with no hooks does nothing", func(t *testing.T) {
		hg := NewGroup[int]()

		// Should not panic
		err := hg.Trigger(42)
		require.NoError(t, err)
		assert.True(t, hg.IsEmpty())
	})
}

func TestHookGroup_AdditionalMethods(t *testing.T) {
	t.Run("All returns all hooks", func(t *testing.T) {
		hg := NewGroup[int]()
		hg.Add(func(i int) error {
			return nil
		}).
			Add(func(i int) error {
				return nil
			}).
			Add(func(i int) error {
				return nil
			})

		hooks := hg.All()

		assert.Len(t, hooks, 3)
	})

	t.Run("Clear removes all hooks", func(t *testing.T) {
		hg := NewGroup[int]()
		hg.Add(func(i int) error {
			return nil
		}).Add(func(i int) error {
			return nil
		}).Add(func(i int) error {
			return nil
		})

		result := hg.Clear()

		assert.Equal(t, hg, result) // Returns self
		assert.True(t, hg.IsEmpty())
		assert.Equal(t, 0, hg.Len())
	})

	t.Run("Clear is chainable", func(t *testing.T) {
		executed := false
		hg := NewGroup[int]()
		hg.Add(func(i int) error {
			return nil
		}).Clear().Add(func(i int) error {
			executed = true
			return nil
		})

		err := hg.Trigger(42)
		require.NoError(t, err)
		assert.True(t, executed)
		assert.Equal(t, 1, hg.Len())
	})
}
