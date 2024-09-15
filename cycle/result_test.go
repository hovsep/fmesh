package cycle

import (
	"errors"
	"github.com/hovsep/fmesh/component"
	"reflect"
	"testing"
)

func TestResults_Add(t *testing.T) {
	type args struct {
		cycleResult *Result
	}
	tests := []struct {
		name         string
		cycleResults Results
		args         args
		want         Results
	}{
		{
			name:         "happy path",
			cycleResults: NewResults(),
			args: args{
				cycleResult: NewResult().SetCycleNumber(1).WithActivationResults(component.NewActivationResult("c1").SetActivated(true)),
			},
			want: Results{
				{
					cycleNumber: 1,
					activationResults: component.ActivationResults{
						"c1": component.NewActivationResult("c1").SetActivated(true),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResults.Add(tt.args.cycleResult); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewResult(t *testing.T) {
	tests := []struct {
		name string
		want *Result
	}{
		{
			name: "happy path",
			want: &Result{
				cycleNumber:       0,
				activationResults: component.ActivationResults{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewResult(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewResults(t *testing.T) {
	tests := []struct {
		name string
		want Results
	}{
		{
			name: "happy path",
			want: Results{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewResults(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_ActivationResults(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Result
		want        component.ActivationResults
	}{
		{
			name:        "no activation results",
			cycleResult: NewResult(),
			want:        component.ActivationResults{},
		},
		{
			name:        "happy path",
			cycleResult: NewResult().WithActivationResults(component.NewActivationResult("c1").SetActivated(true).WithActivationCode(component.ActivationCodeOK)),
			want: component.ActivationResults{
				"c1": component.NewActivationResult("c1").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResult.ActivationResults(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ActivationResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_CycleNumber(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Result
		want        uint
	}{
		{
			name:        "default number",
			cycleResult: NewResult(),
			want:        0,
		},
		{
			name:        "mutated number",
			cycleResult: NewResult().SetCycleNumber(777),
			want:        777,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResult.CycleNumber(); got != tt.want {
				t.Errorf("CycleNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_HasActivatedComponents(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Result
		want        bool
	}{
		{
			name:        "no activation results at all",
			cycleResult: NewResult(),
			want:        false,
		},
		{
			name: "has activation results, but no component activated",
			cycleResult: NewResult().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoFunction),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeWaitingForInput),
			),
			want: false,
		},
		{
			name: "some components did activate",
			cycleResult: NewResult().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeWaitingForInput),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResult.HasActivatedComponents(); got != tt.want {
				t.Errorf("HasActivatedComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_HasErrors(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Result
		want        bool
	}{
		{
			name:        "no activation results at all",
			cycleResult: NewResult(),
			want:        false,
		},
		{
			name: "has activation results, but no one is error",
			cycleResult: NewResult().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoFunction),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeWaitingForInput),
			),
			want: false,
		},
		{
			name: "some components returned errors",
			cycleResult: NewResult().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithError(errors.New("some error")),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeWaitingForInput),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResult.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_HasPanics(t *testing.T) {
	tests := []struct {
		name        string
		cycleResult *Result
		want        bool
	}{
		{
			name:        "no activation results at all",
			cycleResult: NewResult(),
			want:        false,
		},
		{
			name: "has activation results, but no one is panic",
			cycleResult: NewResult().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoFunction),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeWaitingForInput),
				component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithError(errors.New("some error")),
			),
			want: false,
		},
		{
			name: "some components panicked",
			cycleResult: NewResult().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError).WithError(errors.New("some error")),
				component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeWaitingForInput),
				component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodePanicked).WithError(errors.New("some panic")),
			),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResult.HasPanics(); got != tt.want {
				t.Errorf("HasPanics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_SetCycleNumber(t *testing.T) {
	type args struct {
		n uint
	}
	tests := []struct {
		name        string
		cycleResult *Result
		args        args
		want        *Result
	}{
		{
			name:        "happy path",
			cycleResult: NewResult(),
			args: args{
				n: 23,
			},
			want: &Result{
				cycleNumber:       23,
				activationResults: component.ActivationResults{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResult.SetCycleNumber(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetCycleNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_WithActivationResults(t *testing.T) {
	type args struct {
		activationResults []*component.ActivationResult
	}
	tests := []struct {
		name        string
		cycleResult *Result
		args        args
		want        *Result
	}{
		{
			name:        "nothing added",
			cycleResult: NewResult(),
			args: args{
				activationResults: nil,
			},
			want: NewResult(),
		},
		{
			name:        "adding to empty collection",
			cycleResult: NewResult(),
			args: args{
				activationResults: []*component.ActivationResult{
					component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
					component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				},
			},
			want: &Result{
				cycleNumber: 0,
				activationResults: component.ActivationResults{
					"c1": component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
					"c2": component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				},
			},
		},
		{
			name: "adding to existing collection",
			cycleResult: NewResult().WithActivationResults(
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
			want: &Result{
				cycleNumber: 0,
				activationResults: component.ActivationResults{
					"c1": component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
					"c2": component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
					"c3": component.NewActivationResult("c3").SetActivated(true).WithActivationCode(component.ActivationCodeReturnedError),
					"c4": component.NewActivationResult("c4").SetActivated(true).WithActivationCode(component.ActivationCodePanicked),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cycleResult.WithActivationResults(tt.args.activationResults...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithActivationResults() = %v, want %v", got, tt.want)
			}
		})
	}
}
