package component

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivationResultCollection_Add(t *testing.T) {
	type args struct {
		activationResults []*ActivationResult
	}
	tests := []struct {
		name       string
		collection *ActivationResultCollection
		args       args
		assertions func(t *testing.T, collection *ActivationResultCollection)
	}{
		{
			name:       "adding nothing to empty collection",
			collection: NewActivationResultCollection(),
			args: args{
				activationResults: nil,
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				assert.Zero(t, collection.Len())
				assert.False(t, collection.HasActivationErrors())
				assert.False(t, collection.HasActivationPanics())
				assert.False(t, collection.HasActivatedComponents())
			},
		},
		{
			name:       "adding to empty collection",
			collection: NewActivationResultCollection(),
			args: args{
				activationResults: []*ActivationResult{
					mustNew("c1").newActivationResultOK(),
					mustNew("c2").newActivationResultReturnedError(errors.New("oops")),
				},
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				assert.Equal(t, 2, collection.Len())
				assert.True(t, collection.HasActivatedComponents())
				assert.True(t, collection.HasActivationErrors())
				assert.False(t, collection.HasActivationPanics())
			},
		},
		{
			name: "adding to non-empty collection",
			collection: NewActivationResultCollection().Add(
				mustNew("c1").newActivationResultOK(),
				mustNew("c2").newActivationResultOK(),
			),
			args: args{
				activationResults: []*ActivationResult{
					mustNew("c4").newActivationResultNoInput(),
					mustNew("c5").newActivationResultPanicked(errors.New("panic")),
				},
			},
			assertions: func(t *testing.T, collection *ActivationResultCollection) {
				assert.Equal(t, 4, collection.Len())
				assert.True(t, collection.HasActivationPanics())
				assert.False(t, collection.HasActivationErrors())
				assert.True(t, collection.HasActivatedComponents())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.collection.Add(tt.args.activationResults...)
			if tt.assertions != nil {
				tt.assertions(t, tt.collection)
			}
		})
	}
}

func TestActivationResultCollection_ByName(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)
	collection := NewActivationResultCollection().Add(r1, r2)

	t.Run("existing result", func(t *testing.T) {
		result := collection.ByName("c1")
		assert.NotNil(t, result)
		assert.Equal(t, "c1", result.ComponentName())
		assert.True(t, result.Activated())
	})

	t.Run("non-existing result", func(t *testing.T) {
		result := collection.ByName("c3")
		assert.Nil(t, result)
	})
}

func TestActivationResultCollection_All(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("returns all results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		all := collection.All()
		assert.Len(t, all, 2)
		assert.Contains(t, all, "c1")
		assert.Contains(t, all, "c2")
	})

	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		all := collection.All()
		assert.Empty(t, all)
	})
}

func TestActivationResultCollection_Every(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(true)
	r3 := NewActivationResult("c3").SetActivated(false)

	t.Run("all match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Every(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.True(t, result)
	})

	t.Run("not all match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r3)
		result := collection.Every(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.False(t, result)
	})

	t.Run("empty collection returns true", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.Every(func(r *ActivationResult) bool {
			return false
		})
		assert.True(t, result)
	})
}

func TestActivationResultCollection_ForEach(t *testing.T) {
	r1 := NewActivationResult("c1")
	r2 := NewActivationResult("c2")
	r3 := NewActivationResult("c3")

	t.Run("applies action to all results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		count := 0
		require.NoError(t, collection.ForEach(func(r *ActivationResult) error {
			count++
			return nil
		}))
		assert.Equal(t, 3, count)
	})

	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		count := 0
		require.NoError(t, collection.ForEach(func(r *ActivationResult) error {
			count++
			return nil
		}))
		assert.Equal(t, 0, count)
	})
}

func TestActivationResult_ActivationErrorWithComponentName(t *testing.T) {
	err := errors.New("activation failed")
	r := NewActivationResult("my-component").AddActivationError(err)

	t.Run("returns error with component name", func(t *testing.T) {
		wrappedErr := r.ActivationErrorWithComponentName()
		require.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "my-component")
		assert.Contains(t, wrappedErr.Error(), "activation failed")
	})

	t.Run("wraps nil activation error", func(t *testing.T) {
		r := NewActivationResult("comp")
		wrappedErr := r.ActivationErrorWithComponentName()
		// The method wraps even nil errors, so it always returns an error
		require.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "comp")
	})
}

func TestActivationResult_IsWaitingForInput(t *testing.T) {
	t.Run("is waiting", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeWaitingForInputsClear)
		assert.True(t, IsWaitingForInput(r))
	})

	t.Run("is waiting and keeping inputs", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeWaitingForInputsKeep)
		assert.True(t, IsWaitingForInput(r))
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeOK)
		assert.False(t, IsWaitingForInput(r))
	})
}

func TestActivationResult_WantsToKeepInputs(t *testing.T) {
	t.Run("wants to keep", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeWaitingForInputsKeep)
		assert.True(t, WantsToKeepInputs(r))
	})

	t.Run("does not want to keep", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeWaitingForInputsClear)
		assert.False(t, WantsToKeepInputs(r))
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c").SetActivationCode(ActivationCodeOK)
		assert.False(t, WantsToKeepInputs(r))
	})
}

func TestActivationResult_WithActivationError_Accumulates(t *testing.T) {
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	err3 := errors.New("third error")

	t.Run("single error", func(t *testing.T) {
		r := NewActivationResult("c").AddActivationError(err1)
		assert.Len(t, r.ActivationErrors(), 1)
		require.Error(t, r.ActivationError())
		assert.ErrorIs(t, r.ActivationError(), err1)
	})

	t.Run("multiple errors accumulate", func(t *testing.T) {
		r := NewActivationResult("c").
			AddActivationError(err1).
			AddActivationError(err2).
			AddActivationError(err3)
		assert.Len(t, r.ActivationErrors(), 3)
		require.ErrorIs(t, r.ActivationError(), err1)
		require.ErrorIs(t, r.ActivationError(), err2)
		assert.ErrorIs(t, r.ActivationError(), err3)
	})

	t.Run("no errors returns nil", func(t *testing.T) {
		r := NewActivationResult("c")
		assert.Empty(t, r.ActivationErrors())
		assert.NoError(t, r.ActivationError())
	})
}

// mustNew is a test helper that creates a component and panics on error.
func mustNew(name string, opts ...Option) *Component {
	c, err := New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}
