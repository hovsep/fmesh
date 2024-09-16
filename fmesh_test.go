package fmesh

import (
	"errors"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *FMesh
	}{
		{
			name: "empty name is valid",
			args: args{
				name: "",
			},
			want: &FMesh{
				components: component.Collection{},
			},
		},
		{
			name: "with name",
			args: args{
				name: "fm1",
			},
			want: &FMesh{
				name:       "fm1",
				components: component.Collection{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "empty description",
			fm:   New("fm1"),
			args: args{
				description: "",
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "",
				components:            component.Collection{},
				errorHandlingStrategy: 0,
			},
		},
		{
			name: "with description",
			fm:   New("fm1"),
			args: args{
				description: "descr",
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "descr",
				components:            component.Collection{},
				errorHandlingStrategy: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithDescription(tt.args.description); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithErrorHandlingStrategy(t *testing.T) {
	type args struct {
		strategy ErrorHandlingStrategy
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "default strategy",
			fm:   New("fm1"),
			args: args{
				strategy: 0,
			},
			want: &FMesh{
				name:                  "fm1",
				components:            component.Collection{},
				errorHandlingStrategy: StopOnFirstError,
			},
		},
		{
			name: "custom strategy",
			fm:   New("fm1"),
			args: args{
				strategy: IgnoreAll,
			},
			want: &FMesh{
				name:                  "fm1",
				components:            component.Collection{},
				errorHandlingStrategy: IgnoreAll,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithErrorHandlingStrategy(tt.args.strategy); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithErrorHandlingStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_WithComponents(t *testing.T) {
	type args struct {
		components []*component.Component
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "no components",
			fm:   New("fm1"),
			args: args{
				components: nil,
			},
			want: &FMesh{
				name:                  "fm1",
				description:           "",
				components:            component.Collection{},
				errorHandlingStrategy: 0,
			},
		},
		{
			name: "with single component",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Collection{
					"c1": component.NewComponent("c1"),
				},
			},
		},
		{
			name: "with multiple components",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1"),
					component.NewComponent("c2"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Collection{
					"c1": component.NewComponent("c1"),
					"c2": component.NewComponent("c2"),
				},
			},
		},
		{
			name: "components with duplicating name are collapsed",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.NewComponent("c1").WithDescription("descr1"),
					component.NewComponent("c2").WithDescription("descr2"),
					component.NewComponent("c2").WithDescription("descr3"), //This will overwrite the previous one
					component.NewComponent("c4").WithDescription("descr4"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Collection{
					"c1": component.NewComponent("c1").WithDescription("descr1"),
					"c2": component.NewComponent("c2").WithDescription("descr3"),
					"c4": component.NewComponent("c4").WithDescription("descr4"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.WithComponents(tt.args.components...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_Name(t *testing.T) {
	tests := []struct {
		name string
		fm   *FMesh
		want string
	}{
		{
			name: "empty name is valid",
			fm:   New(""),
			want: "",
		},
		{
			name: "with name",
			fm:   New("fm1"),
			want: "fm1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_Description(t *testing.T) {
	tests := []struct {
		name string
		fm   *FMesh
		want string
	}{
		{
			name: "empty description",
			fm:   New("fm1"),
			want: "",
		},
		{
			name: "with description",
			fm:   New("fm1").WithDescription("descr"),
			want: "descr",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fm.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_Run(t *testing.T) {
	tests := []struct {
		name    string
		fm      *FMesh
		initFM  func(fm *FMesh)
		want    cycle.Collection
		wantErr bool
	}{
		{
			name:    "empty mesh stops after first cycle",
			fm:      New("fm"),
			want:    cycle.NewCollection().Add(cycle.New()),
			wantErr: false,
		},
		{
			name: "unsupported error handling strategy",
			fm: New("fm").WithErrorHandlingStrategy(100).
				WithComponents(
					component.NewComponent("c1").
						WithDescription("This component simply puts a constant on o1").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignal(signal.New(77))
							return nil
						}),
				),
			initFM: func(fm *FMesh) {
				//Fire the mesh
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignal(signal.New("start c1"))
			},
			want: cycle.NewCollection().Add(
				cycle.New().
					WithActivationResults(component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeOK)),
			),
			wantErr: true,
		},
		{
			name: "stop on first error on first cycle",
			fm: New("fm").
				WithErrorHandlingStrategy(StopOnFirstError).
				WithComponents(
					component.NewComponent("c1").
						WithDescription("This component just returns an unexpected error").
						WithInputs("i1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							return errors.New("boom")
						})),
			initFM: func(fm *FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignal(signal.New("start"))
			},
			want: cycle.NewCollection().Add(
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeReturnedError).
							WithError(errors.New("component returned an error: boom")),
					),
			),
			wantErr: true,
		},
		{
			name: "stop on first panic on cycle 3",
			fm: New("fm").
				WithErrorHandlingStrategy(StopOnFirstPanic).
				WithComponents(
					component.NewComponent("c1").
						WithDescription("This component just sends a number to c2").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignal(signal.New(10))
							return nil
						}),
					component.NewComponent("c2").
						WithDescription("This component receives a number from c1 and passes it to c4").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignal(inputs.ByName("i1").Signal())
							return nil
						}),
					component.NewComponent("c3").
						WithDescription("This component returns an error, but the mesh is configured to ignore errors").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							return errors.New("boom")
						}),
					component.NewComponent("c4").
						WithDescription("This component receives a number from c2 and panics").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							panic("no way")
							return nil
						}),
				),
			initFM: func(fm *FMesh) {
				c1, c2, c3, c4 := fm.Components().ByName("c1"), fm.Components().ByName("c2"), fm.Components().ByName("c3"), fm.Components().ByName("c4")
				//Piping
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))
				c2.Outputs().ByName("o1").PipeTo(c4.Inputs().ByName("i1"))

				//Input data
				c1.Inputs().ByName("i1").PutSignal(signal.New("start c1"))
				c3.Inputs().ByName("i1").PutSignal(signal.New("start c3"))
			},
			want: cycle.NewCollection().Add(
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeReturnedError).
							WithError(errors.New("component returned an error: boom")),
						component.NewActivationResult("c4").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c3").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(true).
							WithActivationCode(component.ActivationCodePanicked).
							WithError(errors.New("panicked with: no way")),
					),
			),
			wantErr: true,
		},
		{
			name: "all errors and panics are ignored",
			fm: New("fm").
				WithErrorHandlingStrategy(IgnoreAll).
				WithComponents(
					component.NewComponent("c1").
						WithDescription("This component just sends a number to c2").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignal(signal.New(10))
							return nil
						}),
					component.NewComponent("c2").
						WithDescription("This component receives a number from c1 and passes it to c4").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignal(inputs.ByName("i1").Signal())
							return nil
						}),
					component.NewComponent("c3").
						WithDescription("This component returns an error, but the mesh is configured to ignore errors").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							return errors.New("boom")
						}),
					component.NewComponent("c4").
						WithDescription("This component receives a number from c2 and panics, but the mesh is configured to ignore even panics").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignal(inputs.ByName("i1").Signal())

							// Even component panicked, it managed to set some data on output "o1"
							// so that data will be available in next cycle
							panic("no way")
							return nil
						}),
					component.NewComponent("c5").
						WithDescription("This component receives a number from c4").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignal(inputs.ByName("i1").Signal())
							return nil
						}),
				),
			initFM: func(fm *FMesh) {
				c1, c2, c3, c4, c5 := fm.Components().ByName("c1"), fm.Components().ByName("c2"), fm.Components().ByName("c3"), fm.Components().ByName("c4"), fm.Components().ByName("c5")
				//Piping
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))
				c2.Outputs().ByName("o1").PipeTo(c4.Inputs().ByName("i1"))
				c4.Outputs().ByName("o1").PipeTo(c5.Inputs().ByName("i1"))

				//Input data
				c1.Inputs().ByName("i1").PutSignal(signal.New("start c1"))
				c3.Inputs().ByName("i1").PutSignal(signal.New("start c3"))
			},
			want: cycle.NewCollection().Add(
				//c1 and c3 activated, c3 finishes with error
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeReturnedError).
							WithError(errors.New("component returned an error: boom")),
						component.NewActivationResult("c4").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
				// Only c2 is activated
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c3").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
				//Only c4 is activated and panicked
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(true).
							WithActivationCode(component.ActivationCodePanicked).
							WithError(errors.New("panicked with: no way")),
						component.NewActivationResult("c5").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
				//Only c5 is activated (after c4 panicked in previous cycle)
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeOK),
					),
				//Last (control) cycle, no component activated, so f-mesh stops naturally
				cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFM != nil {
				tt.initFM(tt.fm)
			}
			got, err := tt.fm.Run()
			assert.Equal(t, len(tt.want), len(got))
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			//Compare cycle results one by one
			for i := 0; i < len(got); i++ {
				assert.Equal(t, len(tt.want[i].ActivationResults()), len(got[i].ActivationResults()), "ActivationResultCollection len mismatch")

				//Compare activation results
				for componentName, gotActivationResult := range got[i].ActivationResults() {
					assert.Equal(t, tt.want[i].ActivationResults()[componentName].Activated(), gotActivationResult.Activated())
					assert.Equal(t, tt.want[i].ActivationResults()[componentName].ComponentName(), gotActivationResult.ComponentName())
					assert.Equal(t, tt.want[i].ActivationResults()[componentName].Code(), gotActivationResult.Code())

					if tt.want[i].ActivationResults()[componentName].HasError() {
						assert.EqualError(t, tt.want[i].ActivationResults()[componentName].Error(), gotActivationResult.Error().Error())
					} else {
						assert.False(t, gotActivationResult.HasError())
					}
				}
			}
		})
	}
}

func TestFMesh_runCycle(t *testing.T) {
	tests := []struct {
		name   string
		fm     *FMesh
		initFM func(fm *FMesh)
		want   *cycle.Cycle
	}{
		{
			name: "empty mesh",
			fm:   New("empty mesh"),
			want: cycle.New(),
		},
		{
			name: "mesh has components, but no one is activated",
			fm: New("test").WithComponents(
				component.NewComponent("c1").
					WithDescription("I do not have any input signal set, hence I will never be activated").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("o1").PutSignal(signal.New("this signal will never be sent"))
						return nil
					}),

				component.NewComponent("c2").
					WithDescription("I do not have activation func set").
					WithInputs("i1").
					WithOutputs("o1"),

				component.NewComponent("c3").
					WithDescription("I'm waiting for specific input").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						if !inputs.ByNames("i1", "i2").AllHaveSignal() {
							return component.NewErrWaitForInputs(true)
						}
						return nil
					}),
			),
			initFM: func(fm *FMesh) {
				//Only i1 is set, while component is waiting for both i1 and i2 to be set
				fm.Components().ByName("c3").Inputs().ByName("i1").PutSignal(signal.New(123))
			},
			want: cycle.New().
				WithActivationResults(
					component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
					component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoFunction),
					component.NewActivationResult("c3").SetActivated(false).WithActivationCode(component.ActivationCodeWaitingForInput)),
		},
		{
			name: "all components activated in one cycle (concurrently)",
			fm: New("test").WithComponents(
				component.NewComponent("c1").WithDescription("").WithInputs("i1").WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					return nil
				}),
				component.NewComponent("c2").WithDescription("").WithInputs("i1").WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					return nil
				}),
				component.NewComponent("c3").WithDescription("").WithInputs("i1").WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					return nil
				}),
			),
			initFM: func(fm *FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignal(signal.New(1))
				fm.Components().ByName("c2").Inputs().ByName("i1").PutSignal(signal.New(2))
				fm.Components().ByName("c3").Inputs().ByName("i1").PutSignal(signal.New(3))
			},
			want: cycle.New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c2").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c3").SetActivated(true).WithActivationCode(component.ActivationCodeOK),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFM != nil {
				tt.initFM(tt.fm)
			}
			if got := tt.fm.runCycle(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("runCycle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFMesh_drainComponents(t *testing.T) {
	tests := []struct {
		name                 string
		fm                   *FMesh
		initFM               func(fm *FMesh)
		assertionsAfterDrain func(t *testing.T, fm *FMesh)
	}{
		{
			name: "no components",
			fm:   New("empty_fm"),
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				assert.Empty(t, fm.Components())
			},
		},
		{
			name: "no signals to be drained",
			fm: New("fm").WithComponents(
				component.NewComponent("c1").WithInputs("i1").WithOutputs("o1"),
				component.NewComponent("c2").WithInputs("i1").WithOutputs("o1"),
			),
			initFM: func(fm *FMesh) {
				//Create a pipe
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				//All ports in all components are empty
				assert.False(t, fm.Components().ByName("c1").Inputs().AnyHasSignal())
				assert.False(t, fm.Components().ByName("c1").Outputs().AnyHasSignal())
				assert.False(t, fm.Components().ByName("c2").Inputs().AnyHasSignal())
				assert.False(t, fm.Components().ByName("c2").Outputs().AnyHasSignal())
			},
		},
		{
			name: "there are signals on output, but no pipes",
			fm: New("fm").WithComponents(
				component.NewComponent("c1").WithInputs("i1").WithOutputs("o1"),
				component.NewComponent("c2").WithInputs("i1").WithOutputs("o1"),
			),
			initFM: func(fm *FMesh) {
				//Both components have signals on their outputs
				fm.Components().ByName("c1").Outputs().ByName("o1").PutSignal(signal.New(1))
				fm.Components().ByName("c2").Outputs().ByName("o1").PutSignal(signal.New(1))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				//Output signals are still there
				assert.True(t, fm.Components().ByName("c1").Outputs().ByName("o1").HasSignal())
				assert.True(t, fm.Components().ByName("c2").Outputs().ByName("o1").HasSignal())

				//Inputs are clear
				assert.False(t, fm.Components().ByName("c1").Inputs().ByName("i1").HasSignal())
				assert.False(t, fm.Components().ByName("c2").Inputs().ByName("i1").HasSignal())
			},
		},
		{
			name: "happy path",
			fm: New("fm").WithComponents(
				component.NewComponent("c1").WithInputs("i1").WithOutputs("o1"),
				component.NewComponent("c2").WithInputs("i1").WithOutputs("o1"),
			),
			initFM: func(fm *FMesh) {
				//Create a pipe
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))

				//c1 has a signal which must go to c2.i1 after drain
				c1.Outputs().ByName("o1").PutSignal(signal.New(123))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")

				assert.True(t, c2.Inputs().ByName("i1").HasSignal())                    //Signal is transferred to destination port
				assert.False(t, c1.Outputs().ByName("o1").HasSignal())                  //Source port is cleaned up
				assert.Equal(t, c2.Inputs().ByName("i1").Signal().Payload().(int), 123) //The signal is correct
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFM != nil {
				tt.initFM(tt.fm)
			}
			tt.fm.drainComponents()
			tt.assertionsAfterDrain(t, tt.fm)
		})
	}
}
