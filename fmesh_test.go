package fmesh

import (
	"errors"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		fmName     string
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name:   "empty name is valid",
			fmName: "",
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Empty(t, fm.Name())
			},
		},
		{
			name:   "with name",
			fmName: "fm1",
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, "fm1", fm.Name())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.fmName)
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestFMesh_WithDescription(t *testing.T) {
	tests := []struct {
		name        string
		fm          *FMesh
		description string
		assertions  func(t *testing.T, fm *FMesh)
	}{
		{
			name:        "empty description",
			fm:          New("fm1"),
			description: "",
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Empty(t, fm.Description())
			},
		},
		{
			name:        "with description",
			fm:          New("fm1"),
			description: "descr",
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, "descr", fm.Description())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fm.WithDescription(tt.description)
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestFMesh_WithConfig(t *testing.T) {
	tests := []struct {
		name       string
		fm         *FMesh
		config     *Config
		want       *FMesh
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name: "custom config",
			fm:   New("fm1"),
			config: &Config{
				ErrorHandlingStrategy: IgnoreAll,
				CyclesLimit:           9999,
			},
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, IgnoreAll, fm.config.ErrorHandlingStrategy)
				assert.Equal(t, 9999, fm.config.CyclesLimit)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fm.withConfig(tt.config)
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestFMesh_AddComponents(t *testing.T) {
	type args struct {
		components []*component.Component
	}
	tests := []struct {
		name       string
		fm         *FMesh
		args       args
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name: "no components",
			fm:   New("fm1"),
			args: args{
				components: nil,
			},
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Zero(t, fm.Components().Len())
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
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, 1, fm.Components().Len())
				assert.NotNil(t, fm.Components().ByName("c1"))
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
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, 2, fm.Components().Len())
				assert.NotNil(t, fm.Components().ByName("c1"))
				assert.NotNil(t, fm.Components().ByName("c2"))
			},
		},
		{
			name: "adding components with same names leads to error",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.New("c1"),
					component.New("c1"),
				},
			},
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, 0, fm.Components().Len())
				require.Error(t, fm.ChainableErr())
				require.Contains(t, fm.ChainableErr().Error(), "component with name 'c1' already exists")
				assert.True(t, fm.HasChainableErr())
			},
		},

		{
			name: "components inherit logger from fmesh when custom one is not set",
			fm:   New("fm1"),
			args: args{
				components: []*component.Component{
					component.New("c1"), // Must get default logger
					component.New("c2").WithLogger(log.New(io.Discard, "custom", log.LstdFlags)), // Must not be overridden by fmesh
				},
			},
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, 2, fm.Components().Len())
				assert.NotNil(t, fm.ComponentByName("c1"))
				assert.NotNil(t, fm.ComponentByName("c2"))

				assert.Equal(t, "c1:  ", fm.ComponentByName("c1").Logger().Prefix())
				assert.NotEqual(t, io.Discard, fm.ComponentByName("c1").Logger().Writer())

				assert.Equal(t, "c2: custom ", fm.ComponentByName("c2").Logger().Prefix())
				assert.Equal(t, io.Discard, fm.ComponentByName("c2").Logger().Writer())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmAfter := tt.fm.AddComponents(tt.args.components...)
			if tt.assertions != nil {
				tt.assertions(t, fmAfter)
			}
		})
	}
}

func TestFMesh_Run(t *testing.T) {
	tests := []struct {
		name       string
		fm         *FMesh
		initFM     func(fm *FMesh)
		wantCycles *cycle.Group
		wantErr    bool
	}{
		{
			name:       "empty mesh stops after first cycle",
			fm:         New("fm"),
			wantCycles: cycle.NewGroup().Add(cycle.New().WithNumber(1)),
			wantErr:    true,
		},
		{
			name: "unsupported error handling strategy",
			fm: NewWithConfig("fm", &Config{
				ErrorHandlingStrategy: 100,
				CyclesLimit:           0,
			}).
				AddComponents(
					component.New("c1").
						WithDescription("This component simply puts a constant on o1").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							this.OutputByName("o1").PutSignals(signal.New(77))
							return nil
						}),
				),
			initFM: func(fm *FMesh) {
				// Fire the mesh
				fm.Components().ByName("c1").InputByName("i1").PutSignals(signal.New("start c1"))
			},
			wantCycles: cycle.NewGroup().Add(
				cycle.New().
					AddActivationResults(component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeOK)),
			),
			wantErr: true,
		},
		{
			name: "stop on first error on first cycle",
			fm: NewWithConfig("fm", &Config{
				ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
			}).
				AddComponents(
					component.New("c1").
						WithDescription("This component just returns an unexpected error").
						AddInputs("i1").
						WithActivationFunc(func(this *component.Component) error {
							return errors.New("boom")
						})),
			initFM: func(fm *FMesh) {
				fm.Components().ByName("c1").InputByName("i1").PutSignals(signal.New("start"))
			},
			wantCycles: cycle.NewGroup().Add(
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeReturnedError).
							WithActivationError(errors.New("component returned an error: boom")),
					),
			),
			wantErr: true,
		},
		{
			name: "stop on first panic on cycle 3",
			fm: NewWithConfig("fm", &Config{
				ErrorHandlingStrategy: StopOnFirstPanic,
			}).
				AddComponents(
					component.New("c1").
						WithDescription("This component just sends a number to c2").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							this.OutputByName("o1").PutSignals(signal.New(10))
							return nil
						}),
					component.New("c2").
						WithDescription("This component receives a number from c1 and passes it to c4").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
						}),
					component.New("c3").
						WithDescription("This component returns an error, but the mesh is configured to ignore errors").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							return errors.New("boom")
						}),
					component.New("c4").
						WithDescription("This component receives a number from c2 and panics").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							panic("no way")
						}),
				),
			initFM: func(fm *FMesh) {
				c1, c2, c3, c4 := fm.Components().ByName("c1"), fm.Components().ByName("c2"), fm.Components().ByName("c3"), fm.Components().ByName("c4")
				// Piping
				c1.OutputByName("o1").PipeTo(c2.InputByName("i1"))
				c2.OutputByName("o1").PipeTo(c4.InputByName("i1"))

				// Input data
				c1.InputByName("i1").PutSignals(signal.New("start c1"))
				c3.InputByName("i1").PutSignals(signal.New("start c3"))
			},
			wantCycles: cycle.NewGroup().Add(
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeReturnedError).
							WithActivationError(errors.New("component returned an error: boom")),
						component.NewActivationResult("c4").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
				cycle.New().
					AddActivationResults(
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
					AddActivationResults(
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
							WithActivationError(errors.New("panicked with: no way")),
					),
			),
			wantErr: true,
		},
		{
			name: "all errors and panics are ignored",
			fm: NewWithConfig("fm", &Config{
				ErrorHandlingStrategy: IgnoreAll,
			}).
				AddComponents(
					component.New("c1").
						WithDescription("This component just sends a number to c2").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							this.OutputByName("o1").PutSignals(signal.New(10))
							return nil
						}),
					component.New("c2").
						WithDescription("This component receives a number from c1 and passes it to c4").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							_ = port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
							return nil
						}),
					component.New("c3").
						WithDescription("This component returns an error, but the mesh is configured to ignore errors").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							return errors.New("boom")
						}),
					component.New("c4").
						WithDescription("This component receives a number from c2 and panics, but the mesh is configured to ignore even panics").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							_ = port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))

							// Even component panicked, it managed to set some data on output "o1"
							// so that data will be available in next cycle
							panic("no way")
						}),
					component.New("c5").
						WithDescription("This component receives a number from c4").
						AddInputs("i1").
						AddOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
						}),
				),
			initFM: func(fm *FMesh) {
				c1, c2, c3, c4, c5 := fm.Components().ByName("c1"), fm.Components().ByName("c2"), fm.Components().ByName("c3"), fm.Components().ByName("c4"), fm.Components().ByName("c5")
				// Piping
				c1.OutputByName("o1").PipeTo(c2.InputByName("i1"))
				c2.OutputByName("o1").PipeTo(c4.InputByName("i1"))
				c4.OutputByName("o1").PipeTo(c5.InputByName("i1"))

				// Input data
				c1.InputByName("i1").PutSignals(signal.New("start c1"))
				c3.InputByName("i1").PutSignals(signal.New("start c3"))
			},
			wantCycles: cycle.NewGroup().Add(
				// c1 and c3 activated, c3 finishes with error
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c2").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(true).
							WithActivationCode(component.ActivationCodeReturnedError).
							WithActivationError(errors.New("component returned an error: boom")),
						component.NewActivationResult("c4").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
				// Only c2 is activated
				cycle.New().
					AddActivationResults(
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
				// Only c4 is activated and panicked
				cycle.New().
					AddActivationResults(
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
							WithActivationError(errors.New("panicked with: no way")),
						component.NewActivationResult("c5").
							SetActivated(false).
							WithActivationCode(component.ActivationCodeNoInput),
					),
				// Only c5 is activated (after c4 panicked in the previous cycle)
				cycle.New().
					AddActivationResults(
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
				// Last (control) cycle, no component activated, so f-mesh stops naturally
				cycle.New().
					AddActivationResults(
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
			assert.Equal(t, tt.wantCycles.Len(), got.Cycles.Len())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Compare cycle results one by one
			wantCycles, err := tt.wantCycles.All()
			require.NoError(t, err)
			gotCycles, err := got.Cycles.All()
			require.NoError(t, err)

			for i := 0; i < got.Cycles.Len(); i++ {
				wantCycle := wantCycles[i]
				gotCycle := gotCycles[i]
				assert.Equal(t, wantCycle.ActivationResults().Len(), gotCycle.ActivationResults().Len(), "ActivationResultCollection len mismatch")

				// Compare activation results
				gotActivationResults, err := gotCycle.ActivationResults().All()
				require.NoError(t, err)
				for componentName, gotActivationResult := range gotActivationResults {
					assert.Equal(t, wantCycle.ActivationResults().ByName(componentName).Activated(), gotActivationResult.Activated())
					assert.Equal(t, wantCycle.ActivationResults().ByName(componentName).ComponentName(), gotActivationResult.ComponentName())
					assert.Equal(t, wantCycle.ActivationResults().ByName(componentName).Code(), gotActivationResult.Code())

					if wantCycle.ActivationResults().ByName(componentName).IsError() {
						assert.EqualError(t, wantCycle.ActivationResults().ByName(componentName).ActivationError(), gotActivationResult.ActivationError().Error())
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
		name      string
		fm        *FMesh
		initFM    func(fm *FMesh)
		want      *cycle.Cycle
		wantError bool
	}{
		{
			name:      "empty mesh",
			fm:        New("empty mesh"),
			want:      nil,
			wantError: true,
		},
		{
			name: "all components activated in one cycle (concurrently)",
			fm: New("test").AddComponents(
				component.New("c1").
					WithDescription("").
					AddInputs("i1").
					WithActivationFunc(func(this *component.Component) error {
						// No output
						return nil
					}),
				component.New("c2").
					WithDescription("").
					AddInputs("i1").
					AddOutputs("o1", "o2").
					WithActivationFunc(func(this *component.Component) error {
						// Sets output
						this.OutputByName("o1").PutSignals(signal.New(1))

						signals, err := signal.NewGroup(2, 3, 4, 5).All()
						if err != nil {
							return err
						}
						this.OutputByName("o2").PutSignals(signals...)
						return nil
					}),
				component.New("c3").
					WithDescription("").
					AddInputs("i1").
					WithActivationFunc(func(this *component.Component) error {
						// No output
						return nil
					}),
			),
			initFM: func(fm *FMesh) {
				fm.Components().ByName("c1").InputByName("i1").PutSignals(signal.New(1))
				fm.Components().ByName("c2").InputByName("i1").PutSignals(signal.New(2))
				fm.Components().ByName("c3").InputByName("i1").PutSignals(signal.New(3))
			},
			want: cycle.New().AddActivationResults(
				component.NewActivationResult("c1").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c2").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c3").
					SetActivated(true).
					WithActivationCode(component.ActivationCodeOK),
			).WithNumber(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFM != nil {
				tt.initFM(tt.fm)
			}
			tt.fm.runtimeInfo.MarkStarted()
			tt.fm.runCycle()
			gotCycleResult := tt.fm.runtimeInfo.Cycles.Last()
			if tt.wantError {
				assert.True(t, gotCycleResult.HasChainableErr())
				assert.Error(t, gotCycleResult.ChainableErr())
			} else {
				assert.False(t, gotCycleResult.HasChainableErr())
				require.NoError(t, gotCycleResult.ChainableErr())
				assert.Equal(t, tt.want, gotCycleResult)
			}
		})
	}
}

func TestFMesh_mustStop(t *testing.T) {
	tests := []struct {
		name     string
		getFMesh func() *FMesh
		want     bool
		wantErr  error
	}{
		{
			name: "with default config, no time to stop",
			getFMesh: func() *FMesh {
				fm := New("fm")

				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeOK),
				).WithNumber(5)

				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "with default config, reached max cycles",
			getFMesh: func() *FMesh {
				fm := New("fm")
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeOK),
				).WithNumber(1001)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    true,
			wantErr: ErrReachedMaxAllowedCycles,
		},
		{
			name: "mesh finished naturally and must stop",
			getFMesh: func() *FMesh {
				fm := New("fm")
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(false).
						WithActivationCode(component.ActivationCodeNoInput),
				).WithNumber(5)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "mesh hit an error",
			getFMesh: func() *FMesh {
				fm := NewWithConfig("fm", &Config{
					ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
					CyclesLimit:           UnlimitedCycles,
				})
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeReturnedError).
						WithActivationError(errors.New("c1 activation finished with error")),
				).WithNumber(5)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    true,
			wantErr: fmt.Errorf("%w, cycle # %d", ErrHitAnErrorOrPanic, 5),
		},
		{
			name: "mesh hit a panic",
			getFMesh: func() *FMesh {
				fm := NewWithConfig("fm", &Config{
					ErrorHandlingStrategy: StopOnFirstPanic,
				})
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodePanicked).
						WithActivationError(errors.New("c1 panicked")),
				).WithNumber(5)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    true,
			wantErr: ErrHitAPanic,
		},
		{
			name: "mesh has chainable error (e.g., from runCycle)",
			getFMesh: func() *FMesh {
				fm := New("fm")
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						WithActivationCode(component.ActivationCodeOK),
				).WithNumber(5)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				// Simulate error set during runCycle or drainComponents
				fm.WithChainableErr(errors.New("some error during execution"))
				return fm
			},
			want:    true,
			wantErr: errors.New("some error during execution"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.getFMesh()
			got, stopErr := fm.mustStop()
			if tt.wantErr != nil {
				require.ErrorContains(t, stopErr, tt.wantErr.Error())
			} else {
				require.NoError(t, stopErr)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFMesh_validate(t *testing.T) {
	tests := []struct {
		name    string
		getFM   func() *FMesh
		wantErr string
	}{
		{
			name: "valid mesh",
			getFM: func() *FMesh {
				return New("fm").AddComponents(
					component.New("c1").
						AddInputs("in").
						AddOutputs("out"),
				)
			},
			wantErr: "",
		},
		{
			name: "mesh has chainable error",
			getFM: func() *FMesh {
				return New("fm").
					WithChainableErr(errors.New("mesh error")).
					AddComponents(
						component.New("c1").
							AddInputs("in").
							AddOutputs("out"),
					)
			},
			wantErr: "failed to validate fmesh",
		},
		{
			name: "component has chainable error",
			getFM: func() *FMesh {
				fm := New("fm")
				c := component.New("c1").
					AddInputs("in").
					AddOutputs("out")
				fm.AddComponents(c)
				// Set error after adding to mesh
				c.WithChainableErr(errors.New("component error"))
				return fm
			},
			wantErr: "failed to validate component c1",
		},
		{
			name: "component not registered in mesh",
			getFM: func() *FMesh {
				fm := New("fm")
				c := component.New("c1").AddInputs("in").AddOutputs("out")
				// Add component directly without using AddComponents
				fm.components = fm.components.Add(c)
				// Don't set parent mesh - this is invalid
				return fm
			},
			wantErr: "component c1 is not registered in the mesh",
		},
		{
			name: "component has invalid parent mesh",
			getFM: func() *FMesh {
				fm := New("fm")
				otherFm := New("other")
				c := component.New("c1").
					AddInputs("in").
					AddOutputs("out").
					WithParentMesh(otherFm)
				fm.components = fm.components.Add(c)
				return fm
			},
			wantErr: "component c1 has invalid parent mesh",
		},
		{
			name: "port has chainable error",
			getFM: func() *FMesh {
				fm := New("fm")
				c := component.New("c1").
					AddInputs("in").
					AddOutputs("out")
				fm.AddComponents(c)
				// Inject error into port after adding
				c.OutputByName("out").WithChainableErr(errors.New("port error"))
				return fm
			},
			wantErr: "failed to validate port out in component c1",
		},
		{
			name: "pipe leads to unregistered port",
			getFM: func() *FMesh {
				c1 := component.New("c1").
					AddInputs("in").
					AddOutputs("out")

				// Create pipe but don't register destination component
				unregisteredPort := port.NewInput("orphan")
				c1.OutputByName("out").PipeTo(unregisteredPort)

				return New("fm").AddComponents(c1)
			},
			wantErr: "pipe leads to unregistered port orphan in component c1",
		},
		{
			name: "pipe leads to absent component",
			getFM: func() *FMesh {
				c1 := component.New("c1").
					AddInputs("in").
					AddOutputs("out")
				c2 := component.New("c2").AddInputs("in")

				// Create a pipe to a component that won't be added to mesh
				c1.OutputByName("out").PipeTo(c2.InputByName("in"))

				return New("fm").AddComponents(c1) // Only add c1, not c2
			},
			wantErr: "pipe leads to absent component c2",
		},
		{
			// This test case is flaky because of undeterministic map iteration in .validate()
			name: "pipe leads to unregistered component (no parent mesh)",
			getFM: func() *FMesh {
				fm := New("fm")
				c1 := component.New("c1").
					AddInputs("in").
					AddOutputs("out")
				c2 := component.New("c2").AddInputs("in")

				// Pipe between components
				c1.OutputByName("out").PipeTo(c2.InputByName("in"))

				// Add c1 properly
				fm.AddComponents(c1)

				// Add c2 but without setting parent mesh
				fm.components = fm.components.Add(c2)

				return fm
			},
			wantErr: "pipe leads to unregistered component c2",
		},
		{
			name: "pipe leads to component with invalid parent mesh",
			getFM: func() *FMesh {
				fm := New("fm")
				otherFm := New("other")
				c1 := component.New("c1").
					AddInputs("in").
					AddOutputs("out")
				c2 := component.New("c2").
					AddInputs("in").
					WithParentMesh(otherFm) // Wrong parent mesh

				// Pipe between components
				c1.OutputByName("out").PipeTo(c2.InputByName("in"))

				// Add c1 properly
				fm.AddComponents(c1)

				// Add c2 but with wrong parent mesh
				fm.components = fm.components.Add(c2)

				return fm
			},
			// Flaky: Map iteration order determines which error is returned first
			// Either "component c2 has invalid parent mesh" (c2 checked first)
			// or "pipe leads to port in in component c1 that has invalid parent mesh" (c1's pipe checked first)
			wantErr: "invalid parent mesh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.getFM()
			err := fm.validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestFMesh_Run_ErrorHandlingConsistency(t *testing.T) {
	t.Run("beforeRun hook error is stored in chainable error", func(t *testing.T) {
		fm := New("test fm").
			SetupHooks(func(h *Hooks) {
				h.BeforeRun(func(fm *FMesh) error {
					return errors.New("beforeRun hook failed")
				})
			}).
			AddComponents(
				component.New("simple").
					AddInputs("in").
					WithActivationFunc(func(this *component.Component) error {
						return nil
					}))

		fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(1))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		assert.Contains(t, err.Error(), "beforeRun hook failed")
		assert.True(t, fm.HasChainableErr(), "Error should be stored in chainable error")
		assert.ErrorContains(t, fm.ChainableErr(), "beforeRun hook failed")
	})

	t.Run("cycle limit error is stored in chainable error", func(t *testing.T) {
		fm := NewWithConfig("test fm", &Config{
			CyclesLimit:           2,
			ErrorHandlingStrategy: IgnoreAll,
		}).AddComponents(
			component.New("looper").
				AddInputs("in").
				AddOutputs("out").
				WithActivationFunc(func(this *component.Component) error {
					return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
				}))

		fm.ComponentByName("looper").OutputByName("out").
			PipeTo(fm.ComponentByName("looper").InputByName("in"))

		fm.ComponentByName("looper").InputByName("in").PutSignals(signal.New(1))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrReachedMaxAllowedCycles)
		assert.True(t, fm.HasChainableErr(), "Error should be stored in chainable error")
		assert.ErrorIs(t, fm.ChainableErr(), ErrReachedMaxAllowedCycles)
	})

	t.Run("time limit error is stored in chainable error", func(t *testing.T) {
		fm := NewWithConfig("test fm", &Config{
			TimeLimit:             10 * time.Millisecond,
			ErrorHandlingStrategy: IgnoreAll,
		}).AddComponents(
			component.New("sleeper").
				AddInputs("in").
				AddOutputs("out").
				WithActivationFunc(func(this *component.Component) error {
					time.Sleep(50 * time.Millisecond)
					return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
				}))

		fm.ComponentByName("sleeper").OutputByName("out").
			PipeTo(fm.ComponentByName("sleeper").InputByName("in"))

		fm.ComponentByName("sleeper").InputByName("in").PutSignals(signal.New(1))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrTimeLimitExceeded)
		assert.True(t, fm.HasChainableErr(), "Error should be stored in chainable error")
		assert.ErrorIs(t, fm.ChainableErr(), ErrTimeLimitExceeded)
	})

	t.Run("component error is stored in chainable error", func(t *testing.T) {
		fm := NewWithConfig("test fm", &Config{
			ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
		}).AddComponents(
			component.New("faulty").
				AddInputs("in").
				WithActivationFunc(func(this *component.Component) error {
					return errors.New("component failed")
				}))

		fm.ComponentByName("faulty").InputByName("in").PutSignals(signal.New(1))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrHitAnErrorOrPanic)
		assert.True(t, fm.HasChainableErr(), "Error should be stored in chainable error")
		assert.ErrorIs(t, fm.ChainableErr(), ErrHitAnErrorOrPanic)
	})

	t.Run("component panic is stored in chainable error", func(t *testing.T) {
		fm := NewWithConfig("test fm", &Config{
			ErrorHandlingStrategy: StopOnFirstPanic,
		}).AddComponents(
			component.New("panicky").
				AddInputs("in").
				WithActivationFunc(func(this *component.Component) error {
					panic("component panicked")
				}))

		fm.ComponentByName("panicky").InputByName("in").PutSignals(signal.New(1))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrHitAPanic)
		assert.True(t, fm.HasChainableErr(), "Error should be stored in chainable error")
		assert.ErrorIs(t, fm.ChainableErr(), ErrHitAPanic)
	})

	t.Run("unsupported error handling strategy is stored in chainable error", func(t *testing.T) {
		fm := NewWithConfig("test fm", &Config{
			ErrorHandlingStrategy: 999,
		}).AddComponents(
			component.New("simple").
				AddInputs("in").
				WithActivationFunc(func(this *component.Component) error {
					return nil
				}))

		fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(1))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrUnsupportedErrorHandlingStrategy)
		assert.True(t, fm.HasChainableErr(), "Error should be stored in chainable error")
		require.ErrorIs(t, fm.ChainableErr(), ErrUnsupportedErrorHandlingStrategy)
	})
}
