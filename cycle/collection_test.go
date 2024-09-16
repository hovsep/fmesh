package cycle

import (
	"github.com/hovsep/fmesh/component"
	"reflect"
	"testing"
)

func TestNewCollection(t *testing.T) {
	tests := []struct {
		name string
		want Collection
	}{
		{
			name: "happy path",
			want: Collection{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCollection(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollection_Add(t *testing.T) {
	type args struct {
		cycleResults []*Cycle
	}
	tests := []struct {
		name         string
		cycleResults Collection
		args         args
		want         Collection
	}{
		{
			name:         "happy path",
			cycleResults: NewCollection(),
			args: args{
				cycleResults: []*Cycle{
					New().WithActivationResults(component.NewActivationResult("c1").SetActivated(false)),
					New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true)),
				},
			},
			want: Collection{
				{
					activationResults: component.ActivationResultCollection{
						"c1": component.NewActivationResult("c1").SetActivated(false),
					},
				},
				{
					activationResults: component.ActivationResultCollection{
						"c1": component.NewActivationResult("c1").SetActivated(true),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResults.Add(tt.args.cycleResults...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
