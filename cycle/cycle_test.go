package cycle

import (
	"errors"
	"github.com/hovsep/fmesh/component"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cycle := New()
		assert.NotNil(t, cycle)
		assert.False(t, cycle.HasChainError())
	})
}

func TestCycle_ActivationResults(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Cycle
		want        component.ActivationResultCollection
	}{
		{
			name:        "no activation results",
			cycleResult: New(),
			want:        component.ActivationResultCollection{},
		},
		{
			name:        "happy path",
			cycleResult: New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true).WithActivationCode(component.ActivationCodeOK)),
			want: component.ActivationResultCollection{
				"c1": component.NewActivationResult("c1").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cycleResult.ActivationResults())
		})
	}
}

func TestCycle_HasActivatedComponents(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Cycle
		want        bool
	}{
		{
			name:        "no activation results at all",
			cycleResult: New(),
			want:        false,
		},
		{
			name: "has activation results, but no component activated",
			cycleResult: New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoFunction),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
			),
			want: false,
		},
		{
			name: "some components did activate",
			cycleResult: New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cycleResult.HasActivatedComponents())
		})
	}
}

func TestCycle_HasErrors(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Cycle
		want        bool
	}{
		{
			name:        "no activation results at all",
			cycleResult: New(),
			want:        false,
		},
		{
			name: "has activation results, but no one is error",
			cycleResult: New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoFunction),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
			),
			want: false,
		},
		{
			name: "some components returned errors",
			cycleResult: New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithError(errors.New("some error")),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cycleResult.HasErrors())
		})
	}
}

func TestCycle_HasPanics(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Cycle
		want        bool
	}{
		{
			name:        "no activation results at all",
			cycleResult: New(),
			want:        false,
		},
		{
			name: "has activation results, but no one is panic",
			cycleResult: New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoFunction),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithError(errors.New("some error")),
			),
			want: false,
		},
		{
			name: "some components panicked",
			cycleResult: New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithError(errors.New("some error")),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodePanicked).WithError(errors.New("some panic")),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cycleResult.HasPanics())
		})
	}
}

func TestCycle_WithActivationResults(t *testing.T) {
	type args struct {
		activationResults []*component.ActivationResult
	}
	tests := []struct {
		name                  string
		cycleResult           *Cycle
		args                  args
		wantActivationResults component.ActivationResultCollection
	}{
		{
			name:        "nothing added",
			cycleResult: New(),
			args: args{
				activationResults: nil,
			},
			wantActivationResults: component.NewActivationResultCollection(),
		},
		{
			name:        "adding to empty collection",
			cycleResult: New(),
			args: args{
				activationResults: []*component.ActivationResult{
					component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
					component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				},
			},
			wantActivationResults: component.ActivationResultCollection{
				"c1": component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				"c2": component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
			},
		},
		{
			name: "adding to non-empty collection",
			cycleResult: New().WithActivationResults(
				component.NewActivationResult("c1").
					SetActivated(false).
					WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK),
			),
			args: args{
				activationResults: []*component.ActivationResult{
					component.NewActivationResult("c3").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError),
					component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodePanicked),
				},
			},
			wantActivationResults: component.ActivationResultCollection{
				"c1": component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				"c2": component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				"c3": component.NewActivationResult("c3").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError),
				"c4": component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodePanicked),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantActivationResults, tt.cycleResult.WithActivationResults(tt.args.activationResults...).ActivationResults())
		})
	}
}
