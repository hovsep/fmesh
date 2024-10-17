package fmesh

import (
	"errors"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
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
				config:     defaultConfig,
			},
		},
		{
			name: "with name",
			args: args{
				name: "fm1",
			},
			want: &FMesh{
				NamedEntity: common.NewNamedEntity("fm1"),
				components:  component.Collection{},
				config:      defaultConfig,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.name))
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
				NamedEntity:     common.NewNamedEntity("fm1"),
				DescribedEntity: common.NewDescribedEntity(""),
				components:      component.Collection{},
				config:          defaultConfig,
			},
		},
		{
			name: "with description",
			fm:   New("fm1"),
			args: args{
				description: "descr",
			},
			want: &FMesh{
				NamedEntity:     common.NewNamedEntity("fm1"),
				DescribedEntity: common.NewDescribedEntity("descr"),
				components:      component.Collection{},
				config:          defaultConfig,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fm.WithDescription(tt.args.description))
		})
	}
}

func TestFMesh_WithConfig(t *testing.T) {
	type args struct {
		config Config
	}
	tests := []struct {
		name string
		fm   *FMesh
		args args
		want *FMesh
	}{
		{
			name: "custom config",
			fm:   New("fm1"),
			args: args{
				config: Config{
					ErrorHandlingStrategy: IgnoreAll,
					CyclesLimit:           9999,
				},
			},
			want: &FMesh{
				NamedEntity: common.NewNamedEntity("fm1"),
				components:  component.Collection{},
				config: Config{
					ErrorHandlingStrategy: IgnoreAll,
					CyclesLimit:           9999,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fm.WithConfig(tt.args.config))
		})
	}
}

func TestFMesh_WithComponents(t *testing.T) {
	type args struct {
		components []*component.Component
	}
	tests := []struct {
		name           string
		fm             *FMesh
		args           args
		wantComponents component.Collection
	}{
		{
			name: "no components",
			fm:   New("fm1"),
			args: args{
				components: nil,
			},
			wantComponents: component.Collection{},
		},
		{
			name: "with single component",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.New("c1"),
				},
			},
			wantComponents: component.Collection{
				"c1": component.New("c1"),
			},
		},
		{
			name: "with multiple components",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.New("c1"),
					component.New("c2"),
				},
			},
			wantComponents: component.Collection{
				"c1": component.New("c1"),
				"c2": component.New("c2"),
			},
		},
		{
			name: "components with duplicating name are collapsed",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.New("c1").WithDescription("descr1"),
					component.New("c2").WithDescription("descr2"),
					component.New("c2").WithDescription("descr3"), //This will overwrite the previous one
					component.New("c4").WithDescription("descr4"),
				},
			},
			wantComponents: component.Collection{
				"c1": component.New("c1").WithDescription("descr1"),
				"c2": component.New("c2").WithDescription("descr3"),
				"c4": component.New("c4").WithDescription("descr4"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantComponents, tt.fm.WithComponents(tt.args.components...).Components())
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
			want:    cycle.NewCollection().With(cycle.New()),
			wantErr: false,
		},
		{
			name: "unsupported error handling strategy",
			fm: New("fm").WithConfig(Config{
				ErrorHandlingStrategy: 100,
				CyclesLimit:           0,
			}).
				WithComponents(
					component.New("c1").
						WithDescription("This component simply puts a constant on o1").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							outputs.ByName("o1").PutSignals(signal.New(77))
							return nil
						}),
				),
			initFM: func(fm *FMesh) {
				//Fire the mesh
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignals(signal.New("start c1"))
			},
			want: cycle.NewCollection().With(
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
				WithConfig(Config{
					ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
				}).
				WithComponents(
					component.New("c1").
						WithDescription("This component just returns an unexpected error").
						WithInputs("i1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							return errors.New("boom")
						})),
			initFM: func(fm *FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignals(signal.New("start"))
			},
			want: cycle.NewCollection().With(
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
				WithConfig(Config{
					ErrorHandlingStrategy: StopOnFirstPanic,
				}).
				WithComponents(
					component.New("c1").
						WithDescription("This component just sends a number to c2").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							outputs.ByName("o1").PutSignals(signal.New(10))
							return nil
						}),
					component.New("c2").
						WithDescription("This component receives a number from c1 and passes it to c4").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),
					component.New("c3").
						WithDescription("This component returns an error, but the mesh is configured to ignore errors").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							return errors.New("boom")
						}),
					component.New("c4").
						WithDescription("This component receives a number from c2 and panics").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
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
				c1.Inputs().ByName("i1").PutSignals(signal.New("start c1"))
				c3.Inputs().ByName("i1").PutSignals(signal.New("start c3"))
			},
			want: cycle.NewCollection().With(
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
				WithConfig(Config{
					ErrorHandlingStrategy: IgnoreAll,
				}).
				WithComponents(
					component.New("c1").
						WithDescription("This component just sends a number to c2").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							outputs.ByName("o1").PutSignals(signal.New(10))
							return nil
						}),
					component.New("c2").
						WithDescription("This component receives a number from c1 and passes it to c4").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),
					component.New("c3").
						WithDescription("This component returns an error, but the mesh is configured to ignore errors").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							return errors.New("boom")
						}),
					component.New("c4").
						WithDescription("This component receives a number from c2 and panics, but the mesh is configured to ignore even panics").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))

							// Even component panicked, it managed to set some data on output "o1"
							// so that data will be available in next cycle
							panic("no way")
							return nil
						}),
					component.New("c5").
						WithDescription("This component receives a number from c4").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
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
				c1.Inputs().ByName("i1").PutSignals(signal.New("start c1"))
				c3.Inputs().ByName("i1").PutSignals(signal.New("start c3"))
			},
			want: cycle.NewCollection().With(
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

					if tt.want[i].ActivationResults()[componentName].IsError() {
						assert.EqualError(t, tt.want[i].ActivationResults()[componentName].Error(), gotActivationResult.Error().Error())
					} else {
						assert.False(t, gotActivationResult.IsError())
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
			name: "all components activated in one cycle (concurrently)",
			fm: New("test").WithComponents(
				component.New("c1").
					WithDescription("").
					WithInputs("i1").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						// No output
						return nil
					}),
				component.New("c2").
					WithDescription("").
					WithInputs("i1").
					WithOutputs("o1", "o2").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						// Sets output
						outputs.ByName("o1").PutSignals(signal.New(1))

						outputs.ByName("o2").PutSignals(signal.NewGroup(2, 3, 4, 5).SignalsOrNil()...)
						return nil
					}),
				component.New("c3").
					WithDescription("").
					WithInputs("i1").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						// No output
						return nil
					}),
			),
			initFM: func(fm *FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignals(signal.New(1))
				fm.Components().ByName("c2").Inputs().ByName("i1").PutSignals(signal.New(2))
				fm.Components().ByName("c3").Inputs().ByName("i1").PutSignals(signal.New(3))
			},
			want: cycle.New().WithActivationResults(
				component.NewActivationResult("c1").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c2").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c3").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFM != nil {
				tt.initFM(tt.fm)
			}
			assert.Equal(t, tt.want, tt.fm.runCycle())
		})
	}
}

func TestFMesh_mustStop(t *testing.T) {
	type args struct {
		cycleResult *cycle.Cycle
	}
	tests := []struct {
		name    string
		fmesh   *FMesh
		args    args
		want    bool
		wantErr error
	}{
		{
			name:  "with default config, no time to stop",
			fmesh: New("fm"),
			args: args{
				cycleResult: cycle.New().WithActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeOK),
				).WithNumber(5),
			},
			want:    false,
			wantErr: nil,
		},
		{
			name:  "with default config, reached max cycles",
			fmesh: New("fm"),
			args: args{
				cycleResult: cycle.New().WithActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeOK),
				).WithNumber(1001),
			},
			want:    true,
			wantErr: ErrReachedMaxAllowedCycles,
		},
		{
			name:  "mesh finished naturally and must stop",
			fmesh: New("fm"),
			args: args{
				cycleResult: cycle.New().WithActivationResults(
					component.NewActivationResult("c1").
						SetActivated(false).
						WithActivationCode(component.ActivationCodeNoInput),
				).WithNumber(5),
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "mesh hit an error",
			fmesh: New("fm").WithConfig(Config{
				ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
				CyclesLimit:           UnlimitedCycles,
			}),
			args: args{
				cycleResult: cycle.New().WithActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeReturnedError).
						WithError(errors.New("c1 activation finished with error")),
				).WithNumber(5),
			},
			want:    true,
			wantErr: ErrHitAnErrorOrPanic,
		},
		{
			name: "mesh hit a panic",
			fmesh: New("fm").WithConfig(Config{
				ErrorHandlingStrategy: StopOnFirstPanic,
			}),
			args: args{
				cycleResult: cycle.New().WithActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodePanicked).
						WithError(errors.New("c1 panicked")),
				).WithNumber(5),
			},
			want:    true,
			wantErr: ErrHitAPanic,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fmesh.mustStop(tt.args.cycleResult)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFMesh_drainComponents(t *testing.T) {
	type args struct {
		cycle *cycle.Cycle
	}
	tests := []struct {
		name       string
		getFM      func() *FMesh
		args       args
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name: "component not activated",
			getFM: func() *FMesh {
				fm := New("fm").
					WithComponents(
						component.New("c1").
							WithDescription("This component has no activation function").
							WithInputs("i1").
							WithOutputs("o1"))

				fm.Components().
					ByName("c1").
					Inputs().
					ByName("i1").
					PutSignals(signal.New("input signal"))

				return fm
			},
			args: args{
				cycle: cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput)),
			},
			assertions: func(t *testing.T, fm *FMesh) {
				// Assert that inputs are not cleared
				assert.True(t, fm.Components().ByName("c1").Inputs().ByName("i1").HasSignals())
			},
		},
		{
			name: "component fully drained",
			getFM: func() *FMesh {
				c1 := component.New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						return nil
					})

				c2 := component.New("c2").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						return nil
					})

				// Pipe
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))

				// Simulate activation of c1
				c1.Outputs().ByName("o1").PutSignals(signal.New("this signal is generated by c1"))

				return New("fm").WithComponents(c1, c2)
			},
			args: args{
				cycle: cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput)),
			},
			assertions: func(t *testing.T, fm *FMesh) {
				// c1 output is cleared
				assert.False(t, fm.Components().ByName("c1").Outputs().ByName("o1").HasSignals())

				// c2 input received flushed signal
				assert.True(t, fm.Components().ByName("c2").Inputs().ByName("i1").HasSignals())
				sig, err := fm.Components().ByName("c2").Inputs().ByName("i1").Buffer().FirstPayload()
				assert.NoError(t, err)
				assert.Equal(t, "this signal is generated by c1", sig.(string))
			},
		},
		{
			name: "component is waiting for inputs",
			getFM: func() *FMesh {

				c1 := component.New("c1").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						return nil
					})

				c2 := component.New("c2").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						return nil
					})

				// Pipe
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))

				// Simulate activation of c1
				// NOTE: normally component should not create any output signal if it is waiting for inputs
				// but technically there is no limitation to do that and then return the special error to wait for inputs.
				// F-mesh just never flushes components waiting for inputs, so this test checks that
				c1.Outputs().ByName("o1").PutSignals(signal.New("this signal is generated by c1"))

				// Also simulate input signal on one port
				c1.Inputs().ByName("i1").PutSignals(signal.New("this is input signal for c1"))

				return New("fm").WithComponents(c1, c2)
			},
			args: args{
				cycle: cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeWaitingForInputsClear).
							WithError(component.NewErrWaitForInputs(false)),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
			},
			assertions: func(t *testing.T, fm *FMesh) {
				// As c1 is waiting for inputs it's outputs must not be flushed
				assert.False(t, fm.Components().ByName("c2").Inputs().ByName("i1").HasSignals())

				// The inputs must be cleared
				assert.False(t, fm.Components().ByName("c1").Inputs().AnyHasSignals())
			},
		},
		{
			name: "component is waiting for inputs and wants to keep input signals",
			getFM: func() *FMesh {

				c1 := component.New("c1").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						return nil
					})

				c2 := component.New("c2").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection) error {
						return nil
					})

				// Pipe
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))

				// Simulate activation of c1
				// NOTE: normally component should not create any output signal if it is waiting for inputs
				// but technically there is no limitation to do that and then return the special error to wait for inputs.
				// F-mesh just never flushes components waiting for inputs, so this test checks that
				c1.Outputs().ByName("o1").PutSignals(signal.New("this signal is generated by c1"))

				// Also simulate input signal on one port
				c1.Inputs().ByName("i1").PutSignals(signal.New("this is input signal for c1"))

				return New("fm").WithComponents(c1, c2)
			},
			args: args{
				cycle: cycle.New().
					WithActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeWaitingForInputsKeep).
							WithError(component.NewErrWaitForInputs(true)),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
			},
			assertions: func(t *testing.T, fm *FMesh) {
				// As c1 is waiting for inputs it's outputs must not be flushed
				assert.False(t, fm.Components().ByName("c2").Inputs().ByName("i1").HasSignals())

				// The inputs must NOT be cleared
				assert.True(t, fm.Components().ByName("c1").Inputs().AnyHasSignals())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.getFM()
			fm.drainComponents(tt.args.cycle)
			if tt.assertions != nil {
				tt.assertions(t, fm)
			}
		})
	}
}
