package hook

import (
	"errors"
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

	t.Run("Add stops on chainable error", func(t *testing.T) {
		hg := NewGroup[int]().WithChainableErr(errors.New("test error"))

		hg.Add(func(i int) error {
			return nil
		})

		assert.Equal(t, 0, hg.Len())
		assert.True(t, hg.HasChainableErr())
	})

	t.Run("Trigger skips execution on chainable error", func(t *testing.T) {
		executed := false
		hg := NewGroup[int]()
		hg.Add(func(i int) error {
			executed = true
			return nil
		})

		hg.WithChainableErr(errors.New("test error"))
		err := hg.Trigger(42)
		require.Error(t, err)
		require.ErrorContains(t, err, "test error")
		assert.False(t, executed)
		assert.True(t, hg.HasChainableErr())
	})

	t.Run("WithChainableErr returns self", func(t *testing.T) {
		hg := NewGroup[int]()
		testErr := errors.New("test error")

		result := hg.WithChainableErr(testErr)

		assert.Equal(t, hg, result)
		assert.Equal(t, testErr, hg.ChainableErr())
	})

	t.Run("ChainableErr returns nil by default", func(t *testing.T) {
		hg := NewGroup[int]()

		assert.False(t, hg.HasChainableErr())
		assert.NoError(t, hg.ChainableErr())
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

	t.Run("Multiple chainable errors overwrite previous", func(t *testing.T) {
		hg := NewGroup[int]()
		err1 := errors.New("first error")
		err2 := errors.New("second error")

		hg.WithChainableErr(err1).WithChainableErr(err2)

		// Last error wins (WithChainableErr overwrites)
		assert.Equal(t, err2, hg.ChainableErr())
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

		hooks, err := hg.All()

		require.NoError(t, err)
		assert.Len(t, hooks, 3)
	})

	t.Run("All returns error on chainable error", func(t *testing.T) {
		hg := NewGroup[int]()
		testErr := errors.New("test error")
		hg.WithChainableErr(testErr)

		hooks, err := hg.All()

		assert.Nil(t, hooks)
		assert.Equal(t, testErr, err)
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

	t.Run("ForEach operates on hook functions", func(t *testing.T) {
		count := 0
		hg := NewGroup[int]()
		hg.Add(func(i int) error {
			return nil
		}).Add(func(i int) error {
			return nil
		}).Add(func(i int) error {
			return nil
		})

		result := hg.ForEach(func(hook func(int) error) {
			count++
		})

		assert.Equal(t, hg, result) // Returns self
		assert.Equal(t, 3, count)
	})

	t.Run("ForEach skips on chainable error", func(t *testing.T) {
		count := 0
		hg := NewGroup[int]()
		hg.Add(func(i int) error {
			return nil
		}).WithChainableErr(errors.New("test error"))

		hg.ForEach(func(hook func(int) error) {
			count++
		})

		assert.Equal(t, 0, count)
	})
}
