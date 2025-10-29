package cycle

import (
	"errors"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cycle := New()
		assert.NotNil(t, cycle)
		assert.False(t, cycle.HasChainableErr())
	})
}

func TestCycle_ActivationResults(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Cycle
		want        *component.ActivationResultCollection
	}{
		{
			name:        "no activation results",
			cycleResult: New(),
			want:        component.NewActivationResultCollection(),
		},
		{
			name:        "happy path",
			cycleResult: New().WithActivationResults(component.NewActivationResult("c1").SetActivated(true).WithActivationCode(component.ActivationCodeOK)),
			want:        component.NewActivationResultCollection().With(component.NewActivationResult("c1").SetActivated(true).WithActivationCode(component.ActivationCodeOK)),
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
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithActivationError(errors.New("some error")),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cycleResult.HasActivationErrors())
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
				component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithActivationError(errors.New("some error")),
			),
			want: false,
		},
		{
			name: "some components panicked",
			cycleResult: New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithActivationError(errors.New("some error")),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodePanicked).WithActivationError(errors.New("some panic")),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cycleResult.HasActivationPanics())
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
		wantActivationResults *component.ActivationResultCollection
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
			wantActivationResults: component.NewActivationResultCollection().With(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
			),
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
			wantActivationResults: component.NewActivationResultCollection().With(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c3").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError),
				component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodePanicked),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantActivationResults, tt.cycleResult.WithActivationResults(tt.args.activationResults...).ActivationResults())
		})
	}
}

func TestCycle_Chainability(t *testing.T) {
	t.Run("WithActivationResults called twice adds results", func(t *testing.T) {
		r1 := component.NewActivationResult("c1")
		r2 := component.NewActivationResult("c2")
		r3 := component.NewActivationResult("c3")

		c := New().
			WithActivationResults(r1, r2).
			WithActivationResults(r3)

		assert.Equal(t, 3, c.ActivationResults().Len())
	})

	t.Run("AddActivationResult called multiple times adds results", func(t *testing.T) {
		r1 := component.NewActivationResult("c1")
		r2 := component.NewActivationResult("c2")
		r3 := component.NewActivationResult("c3")

		c := New().
			AddActivationResult(r1).
			AddActivationResult(r2).
			AddActivationResult(r3)

		assert.Equal(t, 3, c.ActivationResults().Len())
	})

	t.Run("mixed Add and With", func(t *testing.T) {
		r1 := component.NewActivationResult("c1")
		r2 := component.NewActivationResult("c2")
		r3 := component.NewActivationResult("c3")
		r4 := component.NewActivationResult("c4")

		c := New().
			AddActivationResult(r1).
			WithActivationResults(r2, r3).
			AddActivationResult(r4)

		assert.Equal(t, 4, c.ActivationResults().Len())
	})

	t.Run("WithNumber replaces previous value", func(t *testing.T) {
		c := New().
			WithNumber(1).
			WithNumber(2)

		assert.Equal(t, 2, c.Number())
	})
}
