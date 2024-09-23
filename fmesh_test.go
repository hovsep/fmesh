package fmesh

import (
	"errors"
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
			assert.Equal(t, tt.want, tt.fm.WithDescription(tt.args.description))
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
				errorHandlingStrategy: StopOnFirstErrorOrPanic,
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
			assert.Equal(t, tt.want, tt.fm.WithErrorHandlingStrategy(tt.args.strategy))
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
					component.New("c1"),
				},
			},
			want: &FMesh{
				name: "fm1",
				components: component.Collection{
					"c1": component.New("c1"),
				},
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
			want: &FMesh{
				name: "fm1",
				components: component.Collection{
					"c1": component.New("c1"),
					"c2": component.New("c2"),
				},
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
			want: &FMesh{
				name: "fm1",
				components: component.Collection{
					"c1": component.New("c1").WithDescription("descr1"),
					"c2": component.New("c2").WithDescription("descr3"),
					"c4": component.New("c4").WithDescription("descr4"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fm.WithComponents(tt.args.components...))
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
			assert.Equal(t, tt.want, tt.fm.Name())
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
			assert.Equal(t, tt.want, tt.fm.Description())
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
					component.New("c1").
						WithDescription("This component simply puts a constant on o1").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignals(signal.New(77))
							return nil
						}),
				),
			initFM: func(fm *FMesh) {
				//Fire the mesh
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignals(signal.New("start c1"))
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
				WithErrorHandlingStrategy(StopOnFirstErrorOrPanic).
				WithComponents(
					component.New("c1").
						WithDescription("This component just returns an unexpected error").
						WithInputs("i1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							return errors.New("boom")
						})),
			initFM: func(fm *FMesh) {
				fm.Components().ByName("c1").Inputs().ByName("i1").PutSignals(signal.New("start"))
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
					component.New("c1").
						WithDescription("This component just sends a number to c2").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignals(signal.New(10))
							return nil
						}),
					component.New("c2").
						WithDescription("This component receives a number from c1 and passes it to c4").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),
					component.New("c3").
						WithDescription("This component returns an error, but the mesh is configured to ignore errors").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							return errors.New("boom")
						}),
					component.New("c4").
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
				c1.Inputs().ByName("i1").PutSignals(signal.New("start c1"))
				c3.Inputs().ByName("i1").PutSignals(signal.New("start c3"))
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
					component.New("c1").
						WithDescription("This component just sends a number to c2").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							outputs.ByName("o1").PutSignals(signal.New(10))
							return nil
						}),
					component.New("c2").
						WithDescription("This component receives a number from c1 and passes it to c4").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							port.ForwardSignals(inputs.ByName("i1"), outputs.ByName("o1"))
							return nil
						}),
					component.New("c3").
						WithDescription("This component returns an error, but the mesh is configured to ignore errors").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							return errors.New("boom")
						}),
					component.New("c4").
						WithDescription("This component receives a number from c2 and panics, but the mesh is configured to ignore even panics").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
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
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
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
				component.New("c1").
					WithDescription("I do not have any input signal set, hence I will never be activated").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("o1").PutSignals(signal.New("this signal will never be sent"))
						return nil
					}),

				component.New("c2").
					WithDescription("I do not have activation func set").
					WithInputs("i1").
					WithOutputs("o1"),

				component.New("c3").
					WithDescription("I'm waiting for specific input").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						if !inputs.ByNames("i1", "i2").AllHaveSignals() {
							return component.NewErrWaitForInputs(true)
						}
						return nil
					}),
				component.New("c4").
					WithDescription("I'm waiting for specific input").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						if !inputs.ByNames("i1", "i2").AllHaveSignals() {
							return component.NewErrWaitForInputs(false)
						}
						return nil
					}),
			),
			initFM: func(fm *FMesh) {
				//Only i1 is set, while component is waiting for both i1 and i2 to be set
				fm.Components().ByName("c3").Inputs().ByName("i1").PutSignals(signal.New(123))
				//Same for c4
				fm.Components().ByName("c4").Inputs().ByName("i1").PutSignals(signal.New(456))
			},
			want: cycle.New().
				WithActivationResults(
					component.NewActivationResult("c1").
						SetActivated(false).
						WithActivationCode(component.ActivationCodeNoInput).
						WithStateBefore(component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 0,
								},
							})).
						WithStateAfter(component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 0,
								},
							})),
					component.NewActivationResult("c2").
						SetActivated(false).
						WithActivationCode(component.ActivationCodeNoFunction).
						WithStateBefore(
							component.NewStateSnapshot().
								WithInputPortsMetadata(port.MetadataMap{
									"i1": &port.Metadata{
										SignalBufferLen: 0,
									},
								}).
								WithOutputPortsMetadata(port.MetadataMap{
									"o1": &port.Metadata{
										SignalBufferLen: 0,
									},
								})).
						WithStateAfter(
							component.NewStateSnapshot().
								WithInputPortsMetadata(port.MetadataMap{
									"i1": &port.Metadata{
										SignalBufferLen: 0,
									},
								}).
								WithOutputPortsMetadata(port.MetadataMap{
									"o1": &port.Metadata{
										SignalBufferLen: 0,
									},
								})),
					component.NewActivationResult("c3").
						SetActivated(false).
						WithActivationCode(component.ActivationCodeWaitingForInput).
						WithStateBefore(
							component.NewStateSnapshot().
								WithInputPortsMetadata(port.MetadataMap{
									"i1": &port.Metadata{
										SignalBufferLen: 1,
									},
									"i2": &port.Metadata{
										SignalBufferLen: 0,
									},
								}).
								WithOutputPortsMetadata(port.MetadataMap{
									"o1": &port.Metadata{
										SignalBufferLen: 0,
									},
								})).
						WithStateAfter(
							component.NewStateSnapshot().
								WithInputPortsMetadata(port.MetadataMap{
									"i1": &port.Metadata{
										SignalBufferLen: 1,
									},
									"i2": &port.Metadata{
										SignalBufferLen: 0,
									},
								}).
								WithOutputPortsMetadata(port.MetadataMap{
									"o1": &port.Metadata{
										SignalBufferLen: 0,
									},
								})),
					component.NewActivationResult("c4").
						SetActivated(false).
						WithActivationCode(component.ActivationCodeWaitingForInput).
						WithStateBefore(
							component.NewStateSnapshot().
								WithInputPortsMetadata(port.MetadataMap{
									"i1": &port.Metadata{
										SignalBufferLen: 1,
									},
									"i2": &port.Metadata{
										SignalBufferLen: 0,
									},
								}).
								WithOutputPortsMetadata(port.MetadataMap{
									"o1": &port.Metadata{
										SignalBufferLen: 0,
									},
								})).
						WithStateAfter(
							component.NewStateSnapshot().
								WithInputPortsMetadata(port.MetadataMap{
									"i1": &port.Metadata{
										SignalBufferLen: 1,
									},
									"i2": &port.Metadata{
										SignalBufferLen: 0,
									},
								}).
								WithOutputPortsMetadata(port.MetadataMap{
									"o1": &port.Metadata{
										SignalBufferLen: 0,
									},
								}))),
		},
		{
			name: "all components activated in one cycle (concurrently)",
			fm: New("test").WithComponents(
				component.New("c1").
					WithDescription("").
					WithInputs("i1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						// No output
						return nil
					}),
				component.New("c2").
					WithDescription("").
					WithInputs("i1").
					WithOutputs("o1", "o2").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						// Sets output
						outputs.ByName("o1").PutSignals(signal.New(1))

						outputs.ByName("o2").PutSignals(signal.NewGroup(2, 3, 4, 5)...)
						return nil
					}),
				component.New("c3").
					WithDescription("").
					WithInputs("i1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
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
					WithActivationCode(component.ActivationCodeOK).
					WithStateBefore(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 1,
								},
							})).
					WithStateAfter(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 1,
								},
							})),
				component.NewActivationResult("c2").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK).
					WithStateBefore(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 1,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 0,
								},
								"o2": &port.Metadata{
									SignalBufferLen: 0,
								},
							})).
					WithStateAfter(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 1,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 1,
								},
								"o2": &port.Metadata{
									SignalBufferLen: 4,
								},
							})),
				component.NewActivationResult("c3").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK).
					WithStateBefore(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 1,
								},
							})).
					WithStateAfter(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 1,
								},
							})),
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

func TestFMesh_drainComponentsAfterCycle(t *testing.T) {
	tests := []struct {
		name                 string
		cycle                *cycle.Cycle
		fm                   *FMesh
		initFM               func(fm *FMesh)
		assertionsAfterDrain func(t *testing.T, fm *FMesh)
	}{
		{
			name:  "no components",
			cycle: cycle.New(),
			fm:    New("empty_fm"),
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				assert.Empty(t, fm.Components())
			},
		},
		{
			name: "no signals to be drained",
			cycle: cycle.New().WithActivationResults(
				component.NewActivationResult("c1").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
				component.NewActivationResult("c2").SetActivated(false).WithActivationCode(component.ActivationCodeNoInput),
			),
			fm: New("fm").WithComponents(
				component.New("c1").WithInputs("i1").WithOutputs("o1"),
				component.New("c2").WithInputs("i1").WithOutputs("o1"),
			),
			initFM: func(fm *FMesh) {
				//Create a pipe
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				//All ports in all components are empty
				assert.False(t, fm.Components().ByName("c1").Inputs().AnyHasSignals())
				assert.False(t, fm.Components().ByName("c1").Outputs().AnyHasSignals())
				assert.False(t, fm.Components().ByName("c2").Inputs().AnyHasSignals())
				assert.False(t, fm.Components().ByName("c2").Outputs().AnyHasSignals())
			},
		},
		{
			name: "there are signals on output, but no pipes",
			cycle: cycle.New().WithActivationResults(
				component.NewActivationResult("c1").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK).
					WithStateBefore(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 0,
								},
							})).
					WithStateAfter(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 1,
								},
							})),
				component.NewActivationResult("c2").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK).
					WithStateBefore(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 0,
								},
							})).
					WithStateAfter(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 1,
								},
							}))),
			fm: New("fm").WithComponents(
				component.New("c1").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						return nil
					}),
				component.New("c2").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						return nil
					}),
			),
			initFM: func(fm *FMesh) {
				//Both components have signals on their outputs
				fm.Components().ByName("c1").Outputs().ByName("o1").PutSignals(signal.New(1))
				fm.Components().ByName("c2").Outputs().ByName("o1").PutSignals(signal.New(1))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				//Output signals are still there
				assert.True(t, fm.Components().ByName("c1").Outputs().ByName("o1").HasSignals())
				assert.True(t, fm.Components().ByName("c2").Outputs().ByName("o1").HasSignals())

				//Inputs are clear
				assert.False(t, fm.Components().ByName("c1").Inputs().ByName("i1").HasSignals())
				assert.False(t, fm.Components().ByName("c2").Inputs().ByName("i1").HasSignals())
			},
		},
		{
			name: "happy path",
			cycle: cycle.New().WithActivationResults(
				component.NewActivationResult("c1").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK).
					WithStateBefore(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 0,
								},
							})).
					WithStateAfter(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 1,
								},
							})),
				component.NewActivationResult("c2").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK).
					WithStateBefore(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 0,
								},
							})).
					WithStateAfter(
						component.NewStateSnapshot().
							WithInputPortsMetadata(port.MetadataMap{
								"i1": &port.Metadata{
									SignalBufferLen: 0,
								},
							}).
							WithOutputPortsMetadata(port.MetadataMap{
								"o1": &port.Metadata{
									SignalBufferLen: 0,
								},
							})),
			),
			fm: New("fm").WithComponents(
				component.New("c1").WithInputs("i1").WithOutputs("o1"),
				component.New("c2").WithInputs("i1").WithOutputs("o1"),
			),
			initFM: func(fm *FMesh) {
				//Create a pipe
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")
				c1.Outputs().ByName("o1").PipeTo(c2.Inputs().ByName("i1"))

				//c1 has a signal which must go to c2.i1 after drain
				c1.Outputs().ByName("o1").PutSignals(signal.New(123))
			},
			assertionsAfterDrain: func(t *testing.T, fm *FMesh) {
				c1, c2 := fm.Components().ByName("c1"), fm.Components().ByName("c2")

				assert.True(t, c2.Inputs().ByName("i1").HasSignals())                         //Signal is transferred to destination port
				assert.False(t, c1.Outputs().ByName("o1").HasSignals())                       //Source port is cleaned up
				assert.Equal(t, c2.Inputs().ByName("i1").Signals().FirstPayload().(int), 123) //The signal is correct
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFM != nil {
				tt.initFM(tt.fm)
			}
			tt.fm.drainComponentsAfterCycle(tt.cycle)
			tt.assertionsAfterDrain(t, tt.fm)
		})
	}
}
