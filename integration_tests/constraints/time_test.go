package constraints

import (
	"testing"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustComponent(name string, opts ...component.Option) *component.Component {
	c, err := component.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func mustFMesh(name string, opts ...fmesh.Option) *fmesh.FMesh {
	fm, err := fmesh.New(name, opts...)
	if err != nil {
		panic(err)
	}
	return fm
}

func Test_TimeConstraint(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, runResult *fmesh.RuntimeInfo, err error)
	}{
		{
			name: "mesh stops by time constraint",
			setupFM: func() *fmesh.FMesh {
				ticker := mustComponent("ticker",
					component.WithInputs("tick_in", "start"),
					component.WithOutputs("tick_out"),
					component.WithDescription("simple clock ticking for 10 seconds"),
					component.WithActivationFunc(func(this *component.Component) error {
						ticksCount := this.InputByName("tick_in").Signals().FirstPayloadOrDefault(0).(int)

						if ticksCount == 10 {
							this.Logger().Println("Time is up")
							return nil
						}

						time.Sleep(1 * time.Second)
						this.Logger().Println("Tick #", ticksCount)

						return this.OutputByName("tick_out").PutSignals(signal.New(ticksCount + 1))
					}),
				)

				if err := ticker.LoopbackPipe("tick_out", "tick_in"); err != nil {
					panic(err)
				}

				fm := mustFMesh("fm", fmesh.WithConfig(fmesh.Config{
					Debug:     true,
					TimeLimit: 2 * time.Second,
				}), fmesh.WithDescription("this mesh ticks every second for 10 seconds"))
				if err := fm.AddComponents(ticker); err != nil {
					panic(err)
				}
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				if err := fm.ComponentByName("ticker").InputByName("start").PutSignals(signal.New("start")); err != nil {
					panic(err)
				}
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, runResult *fmesh.RuntimeInfo, err error) {
				require.Error(t, err)
				assert.GreaterOrEqual(t, runResult.Duration(), 2*time.Second)
				assert.LessOrEqual(t, runResult.Duration(), 3*time.Second)
			},
		},
		{
			name: "mesh stops naturally",
			setupFM: func() *fmesh.FMesh {
				ticker := mustComponent("ticker",
					component.WithInputs("tick_in", "start"),
					component.WithOutputs("tick_out"),
					component.WithDescription("simple clock ticking for 3 seconds"),
					component.WithActivationFunc(func(this *component.Component) error {
						ticksCount := this.InputByName("tick_in").Signals().FirstPayloadOrDefault(0).(int)

						if ticksCount == 3 {
							this.Logger().Println("Time is up")
							return nil
						}

						time.Sleep(1 * time.Second)
						this.Logger().Println("Tick #", ticksCount)

						return this.OutputByName("tick_out").PutSignals(signal.New(ticksCount + 1))
					}),
				)

				if err := ticker.LoopbackPipe("tick_out", "tick_in"); err != nil {
					panic(err)
				}

				fm := mustFMesh("fm", fmesh.WithConfig(fmesh.Config{
					Debug:     true,
					TimeLimit: 0,
				}), fmesh.WithDescription("this mesh ticks every second for 10 seconds"))
				if err := fm.AddComponents(ticker); err != nil {
					panic(err)
				}
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				if err := fm.ComponentByName("ticker").InputByName("start").PutSignals(signal.New("start")); err != nil {
					panic(err)
				}
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, runResult *fmesh.RuntimeInfo, err error) {
				require.NoError(t, err)
				assert.Greater(t, runResult.Duration(), 3*time.Second)
				assert.LessOrEqual(t, runResult.Duration(), 4*time.Second)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.setupFM()
			tt.setInputs(fm)
			runResult, err := fm.Run()
			tt.assertions(t, fm, runResult, err)
		})
	}
}
