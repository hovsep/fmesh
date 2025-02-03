package component

import (
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/port"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewComponent(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Component
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: New(""),
		},
		{
			name: "with name",
			args: args{
				name: "multiplier",
			},
			want: New("multiplier"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.name))
		})
	}
}

func TestComponent_WithDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name      string
		component *Component
		args      args
		want      *Component
	}{
		{
			name:      "happy path",
			component: New("c1"),
			args: args{
				description: "descr",
			},
			want: &Component{
				NamedEntity:     common.NewNamedEntity("c1"),
				DescribedEntity: common.NewDescribedEntity("descr"),
				LabeledEntity:   common.NewLabeledEntity(nil),
				Chainable:       common.NewChainable(),
				inputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionIn,
				}),
				outputs: port.NewCollection().WithDefaultLabels(common.LabelsCollection{
					port.DirectionLabel: port.DirectionOut,
				}),
				f:     nil,
				state: NewState(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.component.WithDescription(tt.args.description))
		})
	}
}

func TestComponent_WithLabels(t *testing.T) {
	type args struct {
		labels common.LabelsCollection
	}
	tests := []struct {
		name       string
		component  *Component
		args       args
		assertions func(t *testing.T, component *Component)
	}{
		{
			name:      "happy path",
			component: New("c1"),
			args: args{
				labels: common.LabelsCollection{
					"l1": "v1",
					"l2": "v2",
				},
			},
			assertions: func(t *testing.T, component *Component) {
				assert.Len(t, component.Labels(), 2)
				assert.True(t, component.HasAllLabels("l1", "l2"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			componentAfter := tt.component.WithLabels(tt.args.labels)
			if tt.assertions != nil {
				tt.assertions(t, componentAfter)
			}
		})
	}
}
