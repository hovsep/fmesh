package cycle

import (
	"github.com/hovsep/fmesh/component"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGroup(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		group := NewGroup()
		assert.NotNil(t, group)
	})
}

func TestGroup_With(t *testing.T) {
	type args struct {
		cycles []*Cycle
	}
	tests := []struct {
		name  string
		group *Group
		args  args
		want  *Group
	}{
		{
			name:  "no addition to empty group",
			group: NewGroup(),
			args: args{
				cycles: nil,
			},
			want: NewGroup(),
		},
		{
			name:  "adding to existing group",
			group: NewGroup().With(New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))),
			args: args{
				cycles: nil,
			},
			want: NewGroup().With(New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))),
		},
		{
			name:  "adding to empty group",
			group: NewGroup(),
			args: args{
				cycles: []*Cycle{New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))},
			},
			want: NewGroup().With(New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))),
		},
		{
			name:  "adding to existing group",
			group: NewGroup().With(New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true))),
			args: args{
				cycles: []*Cycle{New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false))},
			},
			want: NewGroup().With(
				New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true)),
				New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false)),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.group.With(tt.args.cycles...))
		})
	}
}
