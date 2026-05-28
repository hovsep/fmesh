package state

import (
	"math/rand"
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
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

func Test_State(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles []*cycle.Cycle, err error)
	}{
		{
			name: "stateful counter",
			setupFM: func() *fmesh.FMesh {
				producer := mustComponent("producer",
					component.WithInputs("demand_rate"),
					component.WithOutputs("signal_out"),
					component.WithActivationFunc(func(this *component.Component) error {
						demandRate := this.InputByName("demand_rate").Signals().FirstPayloadOrDefault(1).(int)
						this.Logger().Println("demand rate= ", demandRate)

						for range demandRate {
							if err := this.OutputByName("signal_out").PutSignals(signal.New(rand.Int())); err != nil {
								return err
							}
						}
						return nil
					}),
				).WithDescription("produces some signals")

				counter := mustComponent("stateful_counter",
					component.WithInputs("bypass_in"),
					component.WithOutputs("bypass_out"),
					component.WithActivationFunc(func(this *component.Component) error {
						count := this.State().Get("observed_signals_count").(int)

						defer func() {
							this.State().Set("observed_signals_count", count)
						}()

						count += this.InputByName("bypass_in").Signals().Len()
						this.Logger().Println("so far signals observed ", count)

						_ = port.ForwardSignals(this.InputByName("bypass_in"), this.OutputByName("bypass_out"))

						return nil
					}),
				).WithDescription("counts all observed signals and bypasses them down the stream").
					WithInitialState(func(state component.State) {
						state.Set("observed_signals_count", 0)
					})

				consumer := mustComponent("consumer",
					component.WithInputs("signal_in", "start"),
					component.WithOutputs("consumed_signals", "demand_rate"),
					component.WithActivationFunc(func(this *component.Component) error {
						demandShape := this.State().Get("demand_shape").([]int)
						defer func() {
							this.State().Set("demand_shape", demandShape)
						}()

						if len(demandShape) > 0 {
							// Pop demand rate
							demandRate := demandShape[0]
							demandShape = demandShape[1:]

							if err := this.OutputByName("demand_rate").PutSignals(signal.New(demandRate)); err != nil {
								return err
							}
						}

						// Consume signals
						return port.ForwardSignals(this.InputByName("signal_in"), this.OutputByName("consumed_signals"))
					}),
				).WithDescription("consumes signals").
					WithInitialState(func(state component.State) {
						// Simulate uneven demand
						state.Set("demand_shape", []int{3, 70, 22, 1350})
					})

				if err := producer.OutputByName("signal_out").PipeTo(counter.InputByName("bypass_in")); err != nil {
					panic(err)
				}
				if err := counter.OutputByName("bypass_out").PipeTo(consumer.InputByName("signal_in")); err != nil {
					panic(err)
				}
				if err := consumer.OutputByName("demand_rate").PipeTo(producer.InputByName("demand_rate")); err != nil {
					panic(err)
				}

				fm := mustFMesh("fm", fmesh.WithConfig(fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10000,
				}))
				if err := fm.AddComponents(producer, counter, consumer); err != nil {
					panic(err)
				}
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				if err := fm.Components().ByName("consumer").InputByName("start").PutSignals(signal.New("start demand")); err != nil {
					panic(err)
				}
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles []*cycle.Cycle, err error) {
				require.NoError(t, err)

				consumedSignals := fm.Components().ByName("consumer").OutputByName("consumed_signals").Signals()

				// AllMatch signals transferred from producer to consumer
				assert.Equal(t, 3+70+22+1350, consumedSignals.Len())

				// Counter state reflects correct count
				assert.Equal(t, 3+70+22+1350, fm.Components().ByName("stateful_counter").State().Get("observed_signals_count"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.setupFM()
			tt.setInputs(fm)
			runResult, err := fm.Run()
			cycles := runResult.Cycles.All()
			tt.assertions(t, fm, cycles, err)
		})
	}
}
