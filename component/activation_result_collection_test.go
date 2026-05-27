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
		all, err := collection.All()
		require.NoError(t, err)
		assert.Len(t, all, 2)
		assert.Contains(t, all, "c1")
		assert.Contains(t, all, "c2")
	})

	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		all, err := collection.All()
		require.NoError(t, err)
		assert.Empty(t, all)
	})
}

func TestActivationResultCollection_IsEmpty(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		assert.True(t, collection.IsEmpty())
	})

	t.Run("non-empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(NewActivationResult("c1"))
		assert.False(t, collection.IsEmpty())
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

func TestActivationResultCollection_Any(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("at least one matches", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Any(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.True(t, result)
	})

	t.Run("none match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		result := collection.Any(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.False(t, result)
	})

	t.Run("empty collection returns false", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.Any(func(r *ActivationResult) bool {
			return true
		})
		assert.False(t, result)
	})
}

func TestActivationResultCollection_Count(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)
	r3 := NewActivationResult("c3").SetActivated(true)

	t.Run("counts matching results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		count := collection.Count(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.Equal(t, 2, count)
	})

	t.Run("no matches", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		count := collection.Count(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.Equal(t, 0, count)
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

func TestActivationResultCollection_Clear(t *testing.T) {
	r1 := NewActivationResult("c1")
	r2 := NewActivationResult("c2")

	t.Run("clears all results", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		assert.Equal(t, 2, collection.Len())
		collection.Clear()
		assert.Equal(t, 0, collection.Len())
		assert.True(t, collection.IsEmpty())
	})
}

func TestActivationResultCollection_Without(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)
	r3 := NewActivationResult("c3").SetActivated(true)

	t.Run("removes by component name", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		result := collection.Without("c2")
		assert.Equal(t, 2, result.Len())
	})

	t.Run("removes multiple", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2, r3)
		result := collection.Without("c1", "c2")
		assert.Equal(t, 1, result.Len())
	})

	t.Run("removes all", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Without("c1", "c2")
		assert.Equal(t, 0, result.Len())
	})
}

func TestActivationResult_ActivationErrorWithComponentName(t *testing.T) {
	err := errors.New("activation failed")
	r := NewActivationResult("my-component").WithActivationError(err)

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
		r := NewActivationResult("c").WithActivationCode(ActivationCodeWaitingForInputsClear)
		assert.True(t, IsWaitingForInput(r))
	})

	t.Run("is waiting and keeping inputs", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeWaitingForInputsKeep)
		assert.True(t, IsWaitingForInput(r))
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeOK)
		assert.False(t, IsWaitingForInput(r))
	})
}

func TestActivationResult_WantsToKeepInputs(t *testing.T) {
	t.Run("wants to keep", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeWaitingForInputsKeep)
		assert.True(t, WantsToKeepInputs(r))
	})

	t.Run("does not want to keep", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeWaitingForInputsClear)
		assert.False(t, WantsToKeepInputs(r))
	})

	t.Run("not waiting", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationCode(ActivationCodeOK)
		assert.False(t, WantsToKeepInputs(r))
	})
}

func TestActivationResultCollection_FindAny(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("one found", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.FindAny(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.Equal(t, "c1", result.ComponentName())
	})

	t.Run("none match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		result := collection.FindAny(func(r *ActivationResult) bool {
			return r.ComponentName() == "c3"
		})
		assert.Nil(t, result)
	})

	t.Run("empty collection returns nil", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.FindAny(func(r *ActivationResult) bool {
			return true
		})
		assert.Nil(t, result)
	})
}

func TestActivationResultCollection_Filter(t *testing.T) {
	r1 := NewActivationResult("c1").SetActivated(true)
	r2 := NewActivationResult("c2").SetActivated(false)

	t.Run("one found", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r1, r2)
		result := collection.Filter(func(r *ActivationResult) bool {
			return r.Activated()
		})
		assert.False(t, result.IsEmpty())
		assert.Equal(t, 1, result.Len())
	})

	t.Run("none match", func(t *testing.T) {
		collection := NewActivationResultCollection().Add(r2)
		result := collection.Filter(func(r *ActivationResult) bool {
			return r.ComponentName() == "c3"
		})
		assert.True(t, result.IsEmpty())
	})

	t.Run("empty collection returns empty collection", func(t *testing.T) {
		collection := NewActivationResultCollection()
		result := collection.Filter(func(r *ActivationResult) bool {
			return true
		})
		assert.True(t, result.IsEmpty())
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

func TestActivationResult_WithActivationError_Accumulates(t *testing.T) {
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	err3 := errors.New("third error")

	t.Run("single error", func(t *testing.T) {
		r := NewActivationResult("c").WithActivationError(err1)
		assert.Len(t, r.ActivationErrors(), 1)
		require.Error(t, r.ActivationError())
		assert.ErrorIs(t, r.ActivationError(), err1)
	})

	t.Run("multiple errors accumulate", func(t *testing.T) {
		r := NewActivationResult("c").
			WithActivationError(err1).
			WithActivationError(err2).
			WithActivationError(err3)
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
