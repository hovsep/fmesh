package cycle

import (
	"errors"
	"strings"
	"testing"

	"github.com/hovsep/fmesh/component"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		cycle := New()
		assert.NotNil(t, cycle)
	})
}

func TestCycle_ActivationResults(t *testing.T) {
	t.Parallel()
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
			cycleResult: New().AddActivationResults(component.NewActivationResult("c1", component.ActivationCodeOK)),
			want:        component.NewActivationResultCollection().Add(component.NewActivationResult("c1", component.ActivationCodeOK)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cycleResult.ActivationResults())
		})
	}
}

func TestCycle_HasActivatedComponents(t *testing.T) {
	t.Parallel()
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
			cycleResult: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeNoInput),
			),
			want: false,
		},
		{
			name: "some components did activate",
			cycleResult: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeOK),
				component.NewActivationResult("c3", component.ActivationCodeNoInput),
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
	t.Parallel()
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
			cycleResult: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeNoInput),
			),
			want: false,
		},
		{
			name: "some components returned errors",
			cycleResult: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeReturnedError, errors.New("some error")),
				component.NewActivationResult("c3", component.ActivationCodeNoInput),
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
	t.Parallel()
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
			cycleResult: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeReturnedError, errors.New("some error")),
			),
			want: false,
		},
		{
			name: "some components panicked",
			cycleResult: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeReturnedError, errors.New("some error")),
				component.NewActivationResult("c3", component.ActivationCodeNoInput),
				component.NewActivationResult("c4", component.ActivationCodePanicked, errors.New("some panic")),
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

func TestCycle_AddActivationResults(t *testing.T) {
	t.Parallel()
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
					component.NewActivationResult("c1", component.ActivationCodeNoInput),
					component.NewActivationResult("c2", component.ActivationCodeOK),
				},
			},
			wantActivationResults: component.NewActivationResultCollection().Add(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeOK),
			),
		},
		{
			name: "adding to non-empty collection",
			cycleResult: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeOK),
			),
			args: args{
				activationResults: []*component.ActivationResult{
					component.NewActivationResult("c3", component.ActivationCodeReturnedError),
					component.NewActivationResult("c4", component.ActivationCodePanicked),
				},
			},
			wantActivationResults: component.NewActivationResultCollection().Add(
				component.NewActivationResult("c1", component.ActivationCodeNoInput),
				component.NewActivationResult("c2", component.ActivationCodeOK),
				component.NewActivationResult("c3", component.ActivationCodeReturnedError),
				component.NewActivationResult("c4", component.ActivationCodePanicked),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantActivationResults, tt.cycleResult.AddActivationResults(tt.args.activationResults...).ActivationResults())
		})
	}
}

func TestCycle_Chainability(t *testing.T) {
	t.Parallel()
	t.Run("AddActivationResults called twice adds results", func(t *testing.T) {
		r1 := component.NewActivationResult("c1", component.ActivationCodeUndefined)
		r2 := component.NewActivationResult("c2", component.ActivationCodeUndefined)
		r3 := component.NewActivationResult("c3", component.ActivationCodeUndefined)

		c := New().
			AddActivationResults(r1, r2).
			AddActivationResults(r3)

		assert.Equal(t, 3, c.ActivationResults().Len())
	})

	t.Run("AddActivationResults called multiple times adds results", func(t *testing.T) {
		r1 := component.NewActivationResult("c1", component.ActivationCodeUndefined)
		r2 := component.NewActivationResult("c2", component.ActivationCodeUndefined)
		r3 := component.NewActivationResult("c3", component.ActivationCodeUndefined)

		c := New().
			AddActivationResults(r1).
			AddActivationResults(r2).
			AddActivationResults(r3)

		assert.Equal(t, 3, c.ActivationResults().Len())
	})

	t.Run("AddActivationResults supports variadic", func(t *testing.T) {
		r1 := component.NewActivationResult("c1", component.ActivationCodeUndefined)
		r2 := component.NewActivationResult("c2", component.ActivationCodeUndefined)
		r3 := component.NewActivationResult("c3", component.ActivationCodeUndefined)
		r4 := component.NewActivationResult("c4", component.ActivationCodeUndefined)

		c := New().
			AddActivationResults(r1).
			AddActivationResults(r2, r3).
			AddActivationResults(r4)

		assert.Equal(t, 4, c.ActivationResults().Len())
	})

	t.Run("WithNumber replaces previous value", func(t *testing.T) {
		c := New().
			SetNumber(1).
			SetNumber(2)

		assert.Equal(t, 2, c.Number())
	})
}

func TestCycle_AllErrorsCombined(t *testing.T) {
	t.Parallel()
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	tests := []struct {
		name    string
		cycle   *Cycle
		wantErr bool
		wantMsg string
	}{
		{
			name:    "no errors",
			cycle:   New().AddActivationResults(component.NewActivationResult("c1", component.ActivationCodeOK)),
			wantErr: false,
		},
		{
			name: "single error",
			cycle: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeReturnedError, err1),
			),
			wantErr: true,
			wantMsg: "error 1",
		},
		{
			name: "multiple errors",
			cycle: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodeReturnedError, err1),
				component.NewActivationResult("c2", component.ActivationCodeReturnedError, err2),
			),
			wantErr: true,
			wantMsg: "error 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cycle.AllErrorsCombined()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCycle_AllPanicsCombined(t *testing.T) {
	t.Parallel()
	panic1 := errors.New("panic 1")
	panic2 := errors.New("panic 2")

	tests := []struct {
		name    string
		cycle   *Cycle
		wantErr bool
		wantMsg string
	}{
		{
			name:    "no panics",
			cycle:   New().AddActivationResults(component.NewActivationResult("c1", component.ActivationCodeOK)),
			wantErr: false,
		},
		{
			name: "single panic",
			cycle: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodePanicked, panic1),
			),
			wantErr: true,
			wantMsg: "panic 1",
		},
		{
			name: "multiple panics",
			cycle: New().AddActivationResults(
				component.NewActivationResult("c1", component.ActivationCodePanicked, panic1),
				component.NewActivationResult("c2", component.ActivationCodePanicked, panic2),
			),
			wantErr: true,
			wantMsg: "panic 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cycle.AllPanicsCombined()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCycle_AllErrorsCombined_IsInComponentNameOrder(t *testing.T) {
	// The results live in a map; the joined error must not follow map order, or
	// the same failure reads differently from one run to the next.
	c := New()
	for _, name := range []string{"charlie", "alpha", "bravo"} {
		c.AddActivationResults(component.NewActivationResult(name, component.ActivationCodeReturnedError, errors.New("boom")))
	}

	want := c.AllErrorsCombined().Error()
	for range 20 {
		assert.Equal(t, want, c.AllErrorsCombined().Error())
	}
	alpha, bravo, charlie := strings.Index(want, "alpha"), strings.Index(want, "bravo"), strings.Index(want, "charlie")
	assert.Less(t, alpha, bravo)
	assert.Less(t, bravo, charlie)
}
