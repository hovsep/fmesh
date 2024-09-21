package component

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestActivationResultCollection_Add(t *testing.T) {
	type args struct {
		activationResults []*ActivationResult
	}
	tests := []struct {
		name       string
		collection ActivationResultCollection
		args       args
		assertions func(t *testing.T, collection ActivationResultCollection)
	}{
		{
			name:       "adding nothing to empty collection",
			collection: NewActivationResultCollection(),
			args: args{
				activationResults: nil,
			},
			assertions: func(t *testing.T, collection ActivationResultCollection) {
				assert.Len(t, collection, 0)
				assert.False(t, collection.HasErrors())
				assert.False(t, collection.HasPanics())
				assert.False(t, collection.HasActivatedComponents())
			},
		},
		{
			name:       "adding to empty collection",
			collection: NewActivationResultCollection(),
			args: args{
				activationResults: []*ActivationResult{
					NewComponent("c1").newActivationResultOK(),
					NewComponent("c2").newActivationResultReturnedError(errors.New("oops")),
				},
			},
			assertions: func(t *testing.T, collection ActivationResultCollection) {
				assert.Len(t, collection, 2)
				assert.True(t, collection.HasActivatedComponents())
				assert.True(t, collection.HasErrors())
				assert.False(t, collection.HasPanics())
			},
		},
		{
			name: "adding to existing collection",
			collection: NewActivationResultCollection().Add(
				NewComponent("c1").newActivationResultOK(),
				NewComponent("c2").newActivationResultOK(),
			),
			args: args{
				activationResults: []*ActivationResult{
					NewComponent("c4").newActivationResultNoInput(),
					NewComponent("c5").newActivationResultPanicked(errors.New("panic")),
				},
			},
			assertions: func(t *testing.T, collection ActivationResultCollection) {
				assert.Len(t, collection, 4)
				assert.True(t, collection.HasPanics())
				assert.False(t, collection.HasErrors())
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
