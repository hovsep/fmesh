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

var noOpActivationFunc = func(this *component.Component) error { return nil }

func mustNewFMesh(name string, opts ...Option) *FMesh {
	fm, err := New(name, opts...)
	if err != nil {
		panic(err)
	}
	return fm
}

func mustNewComponent(name string, opts ...component.Option) *component.Component {
	c, err := component.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func mustAddInputs(c *component.Component, names ...string) *component.Component {
	if err := c.AddInputs(names...); err != nil {
		panic(err)
	}
	return c
}

func mustAddOutputs(c *component.Component, names ...string) *component.Component {
	if err := c.AddOutputs(names...); err != nil {
		panic(err)
	}
	return c
}

func mustPipeTo(src *port.Port, dsts ...*port.Port) {
	if err := src.PipeTo(dsts...); err != nil {
		panic(err)
	}
}

func mustPutSignals(p *port.Port, signals ...*signal.Signal) {
	if err := p.PutSignals(signals...); err != nil {
		panic(err)
	}
}
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
			got := mustNewFMesh(tt.fmName)
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestFMesh_WithDescription(t *testing.T) {
	t.Run("empty description", func(t *testing.T) {
		fm := mustNewFMesh("fm1")
		assert.Empty(t, fm.Description())
	})

	t.Run("with description", func(t *testing.T) {
		fm := mustNewFMesh("fm1", WithDescription("descr"))
		assert.Equal(t, "descr", fm.Description())
	})

	t.Run("WithDescription replaces previous value", func(t *testing.T) {
		fm := mustNewFMesh("fm1", WithDescription("first"), WithDescription("second"))
		assert.Equal(t, "second", fm.Description())
	})
}

func TestFMesh_WithConfig(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name: "custom config",
			config: Config{
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
			got := mustNewFMesh("fm1", WithConfig(tt.config))
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestWithErrorHandlingStrategy(t *testing.T) {
	tests := []struct {
		name       string
		strategy   ErrorHandlingStrategy
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name:     "sets ignore all",
			strategy: IgnoreAll,
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, IgnoreAll, fm.config.ErrorHandlingStrategy)
			},
		},
		{
			name:     "sets stop on first error or panic",
			strategy: StopOnFirstErrorOrPanic,
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, StopOnFirstErrorOrPanic, fm.config.ErrorHandlingStrategy)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustNewFMesh("fm1", WithErrorHandlingStrategy(tt.strategy))
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestWithCyclesLimit(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		wantErr    bool
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name:  "sets custom limit",
			limit: 42,
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, 42, fm.config.CyclesLimit)
			},
		},
		{
			name:    "zero is rejected",
			limit:   0,
			wantErr: true,
		},
		{
			name:    "negative is rejected",
			limit:   -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, err := New("fm1", WithCyclesLimit(tt.limit))
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, fm)
				return
			}
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, fm)
			}
		})
	}
}

func TestWithUnlimitedCycles(t *testing.T) {
	fm := mustNewFMesh("fm1", WithUnlimitedCycles())
	assert.Equal(t, 0, fm.config.CyclesLimit)
}

func TestWithTimeLimit(t *testing.T) {
	tests := []struct {
		name       string
		limit      time.Duration
		wantErr    bool
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name:  "sets custom time limit",
			limit: 5 * time.Second,
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, 5*time.Second, fm.config.TimeLimit)
			},
		},
		{
			name:    "zero is rejected",
			limit:   0,
			wantErr: true,
		},
		{
			name:    "negative is rejected",
			limit:   -time.Second,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, err := New("fm1", WithTimeLimit(tt.limit))
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, fm)
				return
			}
			require.NoError(t, err)
			if tt.assertions != nil {
				tt.assertions(t, fm)
			}
		})
	}
}

func TestWithUnlimitedTime(t *testing.T) {
	fm := mustNewFMesh("fm1", WithUnlimitedTime())
	assert.Equal(t, time.Duration(0), fm.config.TimeLimit)
}

func TestWithDebug(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name:    "enables debug",
			enabled: true,
			assertions: func(t *testing.T, fm *FMesh) {
				assert.True(t, fm.config.Debug)
			},
		},
		{
			name:    "disables debug",
			enabled: false,
			assertions: func(t *testing.T, fm *FMesh) {
				assert.False(t, fm.config.Debug)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustNewFMesh("fm1", WithDebug(tt.enabled))
			if tt.assertions != nil {
				tt.assertions(t, got)
			}
		})
	}
}

func TestWithLogger(t *testing.T) {
	customLogger := log.New(io.Discard, "test", log.LstdFlags)
	tests := []struct {
		name       string
		logger     *log.Logger
		assertions func(t *testing.T, fm *FMesh, err error)
	}{
		{
			name:   "sets custom logger",
			logger: customLogger,
			assertions: func(t *testing.T, fm *FMesh, err error) {
				require.NoError(t, err)
				assert.NotNil(t, fm)
				assert.Equal(t, customLogger, fm.Logger())
			},
		},
		{
			name:   "logger must not be nil",
			logger: nil,
			assertions: func(t *testing.T, fm *FMesh, err error) {
				require.Error(t, err)
				assert.Nil(t, fm)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New("fm1", WithLogger(tt.logger))
			if tt.assertions != nil {
				tt.assertions(t, got, err)
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
		wantErr    bool
		wantErrMsg string
		assertions func(t *testing.T, fm *FMesh)
	}{
		{
			name: "no components",
			fm:   mustNewFMesh("fm1"),
			args: args{
				components: nil,
			},
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Zero(t, fm.Components().Len())
			},
		},
		{
			name: "with single component",
			fm:   mustNewFMesh("fm1"),
			args: args{
				components: []*component.Component{
					mustNewComponent("c1", component.WithActivationFunc(noOpActivationFunc)),
				},
			},
			assertions: func(t *testing.T, fm *FMesh) {
				assert.Equal(t, 1, fm.Components().Len())
				assert.NotNil(t, fm.Components().ByName("c1"))
			},
		},
		{
			name: "with multiple components",
			fm:   mustNewFMesh("fm1"),
			args: args{
				components: []*component.Component{
					mustNewComponent("c1", component.WithActivationFunc(noOpActivationFunc)),
					mustNewComponent("c2", component.WithActivationFunc(noOpActivationFunc)),
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
			fm:   mustNewFMesh("fm1"),
			args: args{
				components: []*component.Component{
					mustNewComponent("c1", component.WithActivationFunc(noOpActivationFunc)),
					mustNewComponent("c1", component.WithActivationFunc(noOpActivationFunc)),
				},
			},
			wantErr:    true,
			wantErrMsg: `component with name "c1" already exists`,
		},
		{
			name: "adding invalid component",
			fm:   mustNewFMesh("fm1"),
			args: args{
				components: []*component.Component{
					mustNewComponent("c1", component.WithDescription("No AF")),
				},
			},
			wantErr:    true,
			wantErrMsg: `failed to add component "c1"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fm.AddComponents(tt.args.components...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				if tt.assertions != nil {
					tt.assertions(t, tt.fm)
				}
			}
		})
	}
}

func TestFMesh_Run(t *testing.T) {
	tests := []struct {
		name       string
		getFM      func() *FMesh
		initFM     func(fm *FMesh)
		wantCycles *cycle.Group
		wantErr    bool
	}{
		{
			name:       "empty mesh stops after first cycle",
			getFM:      func() *FMesh { return mustNewFMesh("fm") },
			wantCycles: cycle.NewGroup().Add(cycle.New().SetNumber(1)),
			wantErr:    true,
		},
		{
			name: "unsupported error handling strategy",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm", WithConfig(Config{
					ErrorHandlingStrategy: 100,
					CyclesLimit:           0,
				}))
				require.NoError(t, fm.AddComponents(
					mustNewComponent("c1",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component simply puts a constant on o1"),
						component.WithActivationFunc(func(this *component.Component) error {
							return this.OutputByName("o1").PutSignals(signal.New(77))
						}),
					),
				))
				return fm
			},
			initFM: func(fm *FMesh) {
				// Fire the mesh
				mustPutSignals(fm.Components().ByName("c1").InputByName("i1"), signal.New("start c1"))
			},
			wantCycles: cycle.NewGroup().Add(
				cycle.New().
					AddActivationResults(component.NewActivationResult("c1").
						SetActivated(true).
						SetActivationCode(component.ActivationCodeOK)),
			),
			wantErr: true,
		},
		{
			name: "stop on first error on first cycle",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm", WithConfig(Config{
					ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
				}))
				require.NoError(t, fm.AddComponents(
					mustNewComponent("c1",
						component.WithInputs("i1"),
						component.WithDescription("This component just returns an unexpected error"),
						component.WithActivationFunc(func(this *component.Component) error {
							return errors.New("boom")
						}),
					),
				))
				return fm
			},
			initFM: func(fm *FMesh) {
				mustPutSignals(fm.Components().ByName("c1").InputByName("i1"), signal.New("start"))
			},
			wantCycles: cycle.NewGroup().Add(
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeReturnedError).
							WithActivationError(errors.New("component returned an error: boom")),
					),
			),
			wantErr: true,
		},
		{
			name: "ErrWaitingForInputs does not stop mesh with StopOnFirstErrorOrPanic strategy",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm", WithConfig(Config{
					ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
				}))
				require.NoError(t, fm.AddComponents(
					mustNewComponent("c1",
						component.WithInputs("i1", "i2"),
						component.WithOutputs("o1"),
						component.WithDescription("This component waits until it gets signals on both inputs"),
						component.WithActivationFunc(func(this *component.Component) error {
							if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
								return component.ErrWaitingForInputs
							}
							return this.OutputByName("o1").PutSignals(signal.New("done"))
						})),
				))
				return fm
			},
			initFM: func(fm *FMesh) {
				// Only feed i1 first; c1 will wait for i2
				mustPutSignals(fm.Components().ByName("c1").InputByName("i1"), signal.New("first"))
			},
			wantCycles: cycle.NewGroup().Add(
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeWaitingForInputsClear).
							WithActivationError(component.ErrWaitingForInputs),
					),
				// Mesh stops naturally in the next cycle because nothing is activated
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
					),
			),
			wantErr: false,
		},
		{
			// c1 loops back into itself 4 times (counts 1-4), sending odd counts to c2.in1
			// and even counts to c2.in2. c2 uses ErrWaitingForInputsKeep until it has both
			// inputs, then multiplies them. Two pairs (1×2, 3×4) are produced and the mesh
			// stops naturally — proving ErrWaitingForInputsKeep never triggers StopOnFirstErrorOrPanic.
			name: "ErrWaitingForInputsKeep does not stop mesh with StopOnFirstErrorOrPanic strategy",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm", WithConfig(Config{
					ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
				}))
				require.NoError(t, fm.AddComponents(
					mustNewComponent("c1",
						component.WithInputs("trigger", "loop_in"),
						component.WithOutputs("loop_out", "num1", "num2"),
						component.WithDescription("Loops back into itself, routing odd counts to num1 and even counts to num2"),
						component.WithActivationFunc(func(this *component.Component) error {
							count := this.InputByName("loop_in").Signals().FirstPayloadOrDefault(0).(int)
							count++

							if count%2 != 0 {
								if err := this.OutputByName("num1").PutPayloads(count); err != nil {
									return err
								}
							} else {
								if err := this.OutputByName("num2").PutPayloads(count); err != nil {
									return err
								}
							}
							// Stop recursion after 4 activations (2 odd + 2 even = 2 balanced pairs for c2)
							if count < 4 {
								if err := this.OutputByName("loop_out").PutPayloads(count); err != nil {
									return err
								}
							}
							return nil
						})),
					mustNewComponent("c2",
						component.WithInputs("in1", "in2"),
						component.WithOutputs("result"),
						component.WithDescription("Waits for both inputs (keeping signals) then multiplies them"),
						component.WithActivationFunc(func(this *component.Component) error {
							if !this.Inputs().ByNames("in1", "in2").AllHaveSignals() {
								return component.ErrWaitingForInputsKeep
							}
							a := this.InputByName("in1").Signals().FirstPayloadOrDefault(0).(int)
							b := this.InputByName("in2").Signals().FirstPayloadOrDefault(0).(int)
							return this.OutputByName("result").PutPayloads(a * b)
						})),
				))
				return fm
			},
			initFM: func(fm *FMesh) {
				c1 := fm.Components().ByName("c1")
				c2 := fm.Components().ByName("c2")
				mustPipeTo(c1.OutputByName("loop_out"), c1.InputByName("loop_in"))
				mustPipeTo(c1.OutputByName("num1"), c2.InputByName("in1"))
				mustPipeTo(c1.OutputByName("num2"), c2.InputByName("in2"))
				mustPutSignals(c1.InputByName("trigger"), signal.New("start"))
			},
			wantCycles: cycle.NewGroup().Add(
				cycle.New().AddActivationResults(
					component.NewActivationResult("c1").SetActivated(true).SetActivationCode(component.ActivationCodeOK),
					component.NewActivationResult("c2").SetActivated(false).SetActivationCode(component.ActivationCodeNoInput),
				),
				cycle.New().AddActivationResults(
					component.NewActivationResult("c1").SetActivated(true).SetActivationCode(component.ActivationCodeOK),
					component.NewActivationResult("c2").SetActivated(true).SetActivationCode(component.ActivationCodeWaitingForInputsKeep).WithActivationError(component.ErrWaitingForInputsKeep),
				),
				cycle.New().AddActivationResults(
					component.NewActivationResult("c1").SetActivated(true).SetActivationCode(component.ActivationCodeOK),
					component.NewActivationResult("c2").SetActivated(true).SetActivationCode(component.ActivationCodeOK),
				),
				cycle.New().AddActivationResults(
					component.NewActivationResult("c1").SetActivated(true).SetActivationCode(component.ActivationCodeOK),
					component.NewActivationResult("c2").SetActivated(true).SetActivationCode(component.ActivationCodeWaitingForInputsKeep).WithActivationError(component.ErrWaitingForInputsKeep),
				),
				cycle.New().AddActivationResults(
					component.NewActivationResult("c1").SetActivated(false).SetActivationCode(component.ActivationCodeNoInput),
					component.NewActivationResult("c2").SetActivated(true).SetActivationCode(component.ActivationCodeOK),
				),
				cycle.New().AddActivationResults(
					component.NewActivationResult("c1").SetActivated(false).SetActivationCode(component.ActivationCodeNoInput),
					component.NewActivationResult("c2").SetActivated(false).SetActivationCode(component.ActivationCodeNoInput),
				),
			),
			wantErr: false,
		},
		{
			name: "stop on first panic on cycle 3",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm", WithConfig(Config{
					ErrorHandlingStrategy: StopOnFirstPanic,
				}))
				require.NoError(t, fm.AddComponents(
					mustNewComponent("c1",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component just sends a number to c2"),
						component.WithActivationFunc(func(this *component.Component) error {
							return this.OutputByName("o1").PutSignals(signal.New(10))
						})),
					mustNewComponent("c2",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component receives a number from c1 and passes it to c4"),
						component.WithActivationFunc(func(this *component.Component) error {
							return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
						})),
					mustNewComponent("c3",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component returns an error, but the mesh is configured to ignore errors"),
						component.WithActivationFunc(func(this *component.Component) error {
							return errors.New("boom")
						})),
					mustNewComponent("c4",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component receives a number from c2 and panics"),
						component.WithActivationFunc(func(this *component.Component) error {
							panic("no way")
						})),
				))
				return fm
			},
			initFM: func(fm *FMesh) {
				c1, c2, c3, c4 := fm.Components().ByName("c1"), fm.Components().ByName("c2"), fm.Components().ByName("c3"), fm.Components().ByName("c4")
				// Piping
				mustPipeTo(c1.OutputByName("o1"), c2.InputByName("i1"))
				mustPipeTo(c2.OutputByName("o1"), c4.InputByName("i1"))

				// Input data
				mustPutSignals(c1.InputByName("i1"), signal.New("start c1"))
				mustPutSignals(c3.InputByName("i1"), signal.New("start c3"))
			},
			wantCycles: cycle.NewGroup().Add(
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c2").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeReturnedError).
							WithActivationError(errors.New("component returned an error: boom")),
						component.NewActivationResult("c4").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
					),
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c3").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
					),
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(true).
							SetActivationCode(component.ActivationCodePanicked).
							WithActivationError(errors.New("panicked with: no way")),
					),
			),
			wantErr: true,
		},
		{
			name: "all errors and panics are ignored",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm", WithConfig(Config{
					ErrorHandlingStrategy: IgnoreAll,
				}))
				require.NoError(t, fm.AddComponents(
					mustNewComponent("c1",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component just sends a number to c2"),
						component.WithActivationFunc(func(this *component.Component) error {
							return this.OutputByName("o1").PutSignals(signal.New(10))
						})),
					mustNewComponent("c2",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component receives a number from c1 and passes it to c4"),
						component.WithActivationFunc(func(this *component.Component) error {
							_ = port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
							return nil
						})),
					mustNewComponent("c3",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component returns an error, but the mesh is configured to ignore errors"),
						component.WithActivationFunc(func(this *component.Component) error {
							return errors.New("boom")
						})),
					mustNewComponent("c4",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component receives a number from c2 and panics, but the mesh is configured to ignore even panics"),
						component.WithActivationFunc(func(this *component.Component) error {
							_ = port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))

							// Even component panicked, it managed to set some data on output "o1"
							// so that data will be available in next cycle
							panic("no way")
						})),
					mustNewComponent("c5",
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithDescription("This component receives a number from c4"),
						component.WithActivationFunc(func(this *component.Component) error {
							return port.ForwardSignals(this.InputByName("i1"), this.OutputByName("o1"))
						})),
				))
				return fm
			},
			initFM: func(fm *FMesh) {
				c1, c2, c3, c4, c5 := fm.Components().ByName("c1"), fm.Components().ByName("c2"), fm.Components().ByName("c3"), fm.Components().ByName("c4"), fm.Components().ByName("c5")
				// Piping
				mustPipeTo(c1.OutputByName("o1"), c2.InputByName("i1"))
				mustPipeTo(c2.OutputByName("o1"), c4.InputByName("i1"))
				mustPipeTo(c4.OutputByName("o1"), c5.InputByName("i1"))

				// Input data
				mustPutSignals(c1.InputByName("i1"), signal.New("start c1"))
				mustPutSignals(c3.InputByName("i1"), signal.New("start c3"))
			},
			wantCycles: cycle.NewGroup().Add(
				// c1 and c3 activated, c3 finishes with error
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c2").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeReturnedError).
							WithActivationError(errors.New("component returned an error: boom")),
						component.NewActivationResult("c4").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
					),
				// Only c2 is activated
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeOK),
						component.NewActivationResult("c3").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
					),
				// Only c4 is activated and panicked
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(true).
							SetActivationCode(component.ActivationCodePanicked).
							WithActivationError(errors.New("panicked with: no way")),
						component.NewActivationResult("c5").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
					),
				// Only c5 is activated (after c4 panicked in the previous cycle)
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(true).
							SetActivationCode(component.ActivationCodeOK),
					),
				// Last (control) cycle, no component activated, so f-mesh stops naturally
				cycle.New().
					AddActivationResults(
						component.NewActivationResult("c1").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c2").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c3").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c4").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
						component.NewActivationResult("c5").
							SetActivated(false).
							SetActivationCode(component.ActivationCodeNoInput),
					),
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.getFM()
			if tt.initFM != nil {
				tt.initFM(fm)
			}
			got, err := fm.Run()
			assert.Equal(t, tt.wantCycles.Len(), got.Cycles.Len())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Compare cycle results one by one
			wantCycles := tt.wantCycles.All()
			gotCycles := got.Cycles.All()

			for i := range got.Cycles.Len() {
				wantCycle := wantCycles[i]
				gotCycle := gotCycles[i]
				assert.Equal(t, wantCycle.ActivationResults().Len(), gotCycle.ActivationResults().Len(), "ActivationResultCollection len mismatch")

				// Compare activation results
				gotActivationResults := gotCycle.ActivationResults().All()
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
		getFM     func() *FMesh
		initFM    func(fm *FMesh)
		want      *cycle.Cycle
		wantError bool
	}{
		{
			name:      "empty mesh",
			getFM:     func() *FMesh { return mustNewFMesh("empty mesh") },
			want:      nil,
			wantError: true,
		},
		{
			name: "all components activated in one cycle (concurrently)",
			getFM: func() *FMesh {
				fm := mustNewFMesh("test")
				require.NoError(t, fm.AddComponents(
					mustNewComponent("c1",
						component.WithInputs("i1"),
						component.WithActivationFunc(func(this *component.Component) error {
							// No output
							return nil
						})),
					mustNewComponent("c2",
						component.WithInputs("i1"),
						component.WithOutputs("o1", "o2"),
						component.WithActivationFunc(func(this *component.Component) error {
							// Sets output
							if err := this.OutputByName("o1").PutSignals(signal.New(1)); err != nil {
								return err
							}

							signals := signal.NewGroup(2, 3, 4, 5).All()
							return this.OutputByName("o2").PutSignals(signals...)
						})),
					mustNewComponent("c3",
						component.WithInputs("i1"),
						component.WithActivationFunc(func(this *component.Component) error {
							// No output
							return nil
						})),
				))
				return fm
			},
			initFM: func(fm *FMesh) {
				mustPutSignals(fm.Components().ByName("c1").InputByName("i1"), signal.New(1))
				mustPutSignals(fm.Components().ByName("c2").InputByName("i1"), signal.New(2))
				mustPutSignals(fm.Components().ByName("c3").InputByName("i1"), signal.New(3))
			},
			want: cycle.New().AddActivationResults(
				component.NewActivationResult("c1").
					SetActivated(true).
					SetActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c2").
					SetActivated(true).
					SetActivationCode(component.ActivationCodeOK),
				component.NewActivationResult("c3").
					SetActivated(true).
					SetActivationCode(component.ActivationCodeOK),
			).SetNumber(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.getFM()
			if tt.initFM != nil {
				tt.initFM(fm)
			}
			fm.runtimeInfo.MarkStarted()
			cycleErr := fm.runCycle()
			gotCycleResult := fm.runtimeInfo.Cycles.Last()
			if tt.wantError {
				require.Error(t, cycleErr)
			} else {
				require.NoError(t, cycleErr)
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
				fm := mustNewFMesh("fm")

				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						SetActivationCode(component.ActivationCodeOK),
				).SetNumber(5)

				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "with default config, reached max cycles",
			getFMesh: func() *FMesh {
				fm := mustNewFMesh("fm")
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						SetActivationCode(component.ActivationCodeOK),
				).SetNumber(1001)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    true,
			wantErr: ErrReachedMaxAllowedCycles,
		},
		{
			name: "mesh finished naturally and must stop",
			getFMesh: func() *FMesh {
				fm := mustNewFMesh("fm")
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(false).
						SetActivationCode(component.ActivationCodeNoInput),
				).SetNumber(5)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "mesh hit an error",
			getFMesh: func() *FMesh {
				fm := mustNewFMesh("fm", WithConfig(Config{
					ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
					CyclesLimit:           0,
				}))
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						SetActivationCode(component.ActivationCodeReturnedError).
						WithActivationError(errors.New("c1 activation finished with error")),
				).SetNumber(5)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    true,
			wantErr: fmt.Errorf("%w, cycle # %d", ErrHitAnErrorOrPanic, 5),
		},
		{
			name: "mesh hit a panic",
			getFMesh: func() *FMesh {
				fm := mustNewFMesh("fm", WithConfig(Config{
					ErrorHandlingStrategy: StopOnFirstPanic,
				}))
				c := cycle.New().AddActivationResults(
					component.NewActivationResult("c1").
						SetActivated(true).
						SetActivationCode(component.ActivationCodePanicked).
						WithActivationError(errors.New("c1 panicked")),
				).SetNumber(5)
				fm.runtimeInfo.Cycles = fm.runtimeInfo.Cycles.Add(c)
				return fm
			},
			want:    true,
			wantErr: ErrHitAPanic,
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

func TestFMesh_validateMeshStructure(t *testing.T) {
	tests := []struct {
		name    string
		getFM   func() *FMesh
		wantErr string
	}{
		{
			name: "valid mesh",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm")
				require.NoError(t, fm.AddComponents(
					mustNewComponent("c1",
						component.WithInputs("in"),
						component.WithOutputs("out"),
						component.WithActivationFunc(noOpActivationFunc)),
				))
				return fm
			},
			wantErr: "",
		},
		{
			name: "component not registered in mesh",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm")
				c := mustNewComponent("c1", component.WithInputs("in"), component.WithOutputs("out"))
				// Add component directly without using AddComponents (no parent mesh set)
				require.NoError(t, fm.components.Add(c))
				return fm
			},
			wantErr: "wrong parent mesh",
		},
		{
			name: "component has invalid parent mesh",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm")
				otherFm := mustNewFMesh("other")
				c := mustNewComponent("c1",
					component.WithInputs("in"),
					component.WithOutputs("out")).
					SetParentMesh(otherFm)
				require.NoError(t, fm.components.Add(c))
				return fm
			},
			wantErr: "wrong parent mesh",
		},
		{
			name: "pipe leads to unregistered port",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm")
				c1 := mustNewComponent("c1",
					component.WithInputs("in"),
					component.WithOutputs("out"),
					component.WithActivationFunc(noOpActivationFunc))

				// Create pipe but don't register destination component
				unregisteredPort, err := port.NewInput("orphan")
				require.NoError(t, err)
				require.NoError(t, c1.OutputByName("out").PipeTo(unregisteredPort))

				require.NoError(t, fm.AddComponents(c1))
				return fm
			},
			wantErr: "parent component",
		},
		{
			name: "pipe leads to absent component",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm")
				c1 := mustNewComponent("c1",
					component.WithInputs("in"),
					component.WithOutputs("out"),
					component.WithActivationFunc(noOpActivationFunc))

				c2 := mustNewComponent("c2",
					component.WithInputs("in"),
					component.WithActivationFunc(noOpActivationFunc))

				// Create a pipe to a component that won't be added to mesh
				require.NoError(t, c1.OutputByName("out").PipeTo(c2.InputByName("in")))

				require.NoError(t, fm.AddComponents(c1)) // Only add c1, not c2
				return fm
			},
			wantErr: "different mesh",
		},
		{
			name: "pipe leads to component with invalid parent mesh",
			getFM: func() *FMesh {
				fm := mustNewFMesh("fm")
				otherFm := mustNewFMesh("other")
				c1 := mustNewComponent("c1",
					component.WithInputs("in"),
					component.WithOutputs("out"),
					component.WithActivationFunc(noOpActivationFunc))

				c2 := mustNewComponent("c2",
					component.WithInputs("in"),
					component.WithActivationFunc(noOpActivationFunc)).
					SetParentMesh(otherFm) // Wrong parent mesh

				// Pipe between components
				require.NoError(t, c1.OutputByName("out").PipeTo(c2.InputByName("in")))

				// Add c1 properly
				require.NoError(t, fm.AddComponents(c1))

				// Add c2 but with wrong parent mesh
				require.NoError(t, fm.components.Add(c2))

				return fm
			},
			wantErr: "different mesh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.getFM()
			_, err := fm.Run()

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
	t.Run("beforeRun hook error causes Run to return error", func(t *testing.T) {
		fm := mustNewFMesh("test fm")
		fm.SetupHooks(func(h *Hooks) {
			h.BeforeRun(func(fm *FMesh) error {
				return errors.New("beforeRun hook failed")
			})
		})
		require.NoError(t, fm.AddComponents(
			mustNewComponent("simple",
				component.WithInputs("in"),
				component.WithActivationFunc(func(this *component.Component) error {
					return nil
				})),
		))

		require.NoError(t, fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(1)))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		assert.Contains(t, err.Error(), "beforeRun hook failed")
	})

	t.Run("cycle limit error causes Run to return ErrReachedMaxAllowedCycles", func(t *testing.T) {
		fm := mustNewFMesh("test fm", WithConfig(Config{
			CyclesLimit:           2,
			ErrorHandlingStrategy: IgnoreAll,
		}))
		require.NoError(t, fm.AddComponents(
			mustNewComponent("looper",
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(this *component.Component) error {
					return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
				})),
		))

		require.NoError(t, fm.ComponentByName("looper").OutputByName("out").
			PipeTo(fm.ComponentByName("looper").InputByName("in")))

		require.NoError(t, fm.ComponentByName("looper").InputByName("in").PutSignals(signal.New(1)))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrReachedMaxAllowedCycles)
	})

	t.Run("time limit error causes Run to return ErrTimeLimitExceeded", func(t *testing.T) {
		fm := mustNewFMesh("test fm", WithConfig(Config{
			TimeLimit:             10 * time.Millisecond,
			ErrorHandlingStrategy: IgnoreAll,
		}))
		require.NoError(t, fm.AddComponents(
			mustNewComponent("sleeper",
				component.WithInputs("in"),
				component.WithOutputs("out"),
				component.WithActivationFunc(func(this *component.Component) error {
					time.Sleep(50 * time.Millisecond)
					return port.ForwardSignals(this.InputByName("in"), this.OutputByName("out"))
				})),
		))

		require.NoError(t, fm.ComponentByName("sleeper").OutputByName("out").
			PipeTo(fm.ComponentByName("sleeper").InputByName("in")))

		require.NoError(t, fm.ComponentByName("sleeper").InputByName("in").PutSignals(signal.New(1)))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrTimeLimitExceeded)
	})

	t.Run("component error causes Run to return ErrHitAnErrorOrPanic", func(t *testing.T) {
		fm := mustNewFMesh("test fm", WithConfig(Config{
			ErrorHandlingStrategy: StopOnFirstErrorOrPanic,
		}))
		require.NoError(t, fm.AddComponents(
			mustNewComponent("faulty",
				component.WithInputs("in"),
				component.WithActivationFunc(func(this *component.Component) error {
					return errors.New("component failed")
				})),
		))

		require.NoError(t, fm.ComponentByName("faulty").InputByName("in").PutSignals(signal.New(1)))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrHitAnErrorOrPanic)
	})

	t.Run("component panic causes Run to return ErrHitAPanic", func(t *testing.T) {
		fm := mustNewFMesh("test fm", WithConfig(Config{
			ErrorHandlingStrategy: StopOnFirstPanic,
		}))
		require.NoError(t, fm.AddComponents(
			mustNewComponent("panicky",
				component.WithInputs("in"),
				component.WithActivationFunc(func(this *component.Component) error {
					panic("component panicked")
				})),
		))

		require.NoError(t, fm.ComponentByName("panicky").InputByName("in").PutSignals(signal.New(1)))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrHitAPanic)
	})

	t.Run("unsupported error handling strategy causes Run to return ErrUnsupportedErrorHandlingStrategy", func(t *testing.T) {
		fm := mustNewFMesh("test fm", WithConfig(Config{
			ErrorHandlingStrategy: 999,
		}))
		require.NoError(t, fm.AddComponents(
			mustNewComponent("simple",
				component.WithInputs("in"),
				component.WithActivationFunc(func(this *component.Component) error {
					return nil
				})),
		))

		require.NoError(t, fm.ComponentByName("simple").InputByName("in").PutSignals(signal.New(1)))
		_, err := fm.Run()

		require.Error(t, err, "Run should return error")
		require.ErrorIs(t, err, ErrUnsupportedErrorHandlingStrategy)
	})
}

func TestFMesh_Run_ComponentHookFailuresSurface(t *testing.T) {
	newMeshWithFailingHook := func(strategy ErrorHandlingStrategy, configureHooks func(*component.Hooks)) *FMesh {
		fm := mustNewFMesh("hook-failures", WithErrorHandlingStrategy(strategy))
		c := mustNewComponent("c1",
			component.WithInputs("in"),
			component.WithOutputs("out"),
			component.WithActivationFunc(func(this *component.Component) error {
				return this.OutputByName("out").PutSignals(signal.New("done"))
			}),
			component.WithHooks(configureHooks),
		)
		require.NoError(t, fm.AddComponents(c))
		require.NoError(t, c.InputByName("in").PutSignals(signal.New("start")))
		return fm
	}

	errHookFailed := errors.New("hook failed on purpose")

	t.Run("failing BeforeActivation hook surfaces in Run error", func(t *testing.T) {
		fm := newMeshWithFailingHook(StopOnFirstErrorOrPanic, func(h *component.Hooks) {
			h.BeforeActivation(func(*component.Component) error {
				return errHookFailed
			})
		})

		_, err := fm.Run()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrHitAnErrorOrPanic)
		require.ErrorIs(t, err, errHookFailed)
	})

	t.Run("failing OnSuccess hook surfaces in Run error", func(t *testing.T) {
		fm := newMeshWithFailingHook(StopOnFirstErrorOrPanic, func(h *component.Hooks) {
			h.OnSuccess(func(*component.ActivationContext) error {
				return errHookFailed
			})
		})

		_, err := fm.Run()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrHitAnErrorOrPanic)
		require.ErrorIs(t, err, errHookFailed)
	})

	t.Run("IgnoreAll strategy still ignores hook failures", func(t *testing.T) {
		fm := newMeshWithFailingHook(IgnoreAll, func(h *component.Hooks) {
			h.OnSuccess(func(*component.ActivationContext) error {
				return errHookFailed
			})
		})

		_, err := fm.Run()
		require.NoError(t, err)
	})
}
