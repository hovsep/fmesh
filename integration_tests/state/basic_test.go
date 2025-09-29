package state

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func Test_State(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error)
	}{
		{
			name: "stateful counter",
			setupFM: func() *fmesh.FMesh {
				producer := component.New("producer").
					WithDescription("produces some signals").
					WithInputs("demand_rate").
					WithOutputs("signal_out").
					WithActivationFunc(func(this *component.Component) error {
						demandRate := this.InputByName("demand_rate").FirstSignalPayloadOrDefault(1).(int)
						this.Logger().Println("demand rate= ", demandRate)

						for i := 0; i < demandRate; i++ {
							this.OutputByName("signal_out").PutSignals(signal.New(rand.Int()))
						}
						return nil
					})

				counter := component.New("stateful_counter").
					WithDescription("counts all observed signals and bypasses them down the stream").
					WithInputs("bypass_in").
					WithOutputs("bypass_out").
					WithInitialState(func(state component.State) {
						state.Set("observed_signals_count", 0)
					}).
					WithActivationFunc(func(this *component.Component) error {
						count := this.State().Get("observed_signals_count").(int)

						defer func() {
							this.State().Set("observed_signals_count", count)
						}()

						count += this.InputByName("bypass_in").Buffer().Len()
						this.Logger().Println("so far signals observed ", count)

						_ = port.ForwardSignals(this.InputByName("bypass_in"), this.OutputByName("bypass_out"))

						return nil
					})

				consumer := component.New("consumer").
					WithDescription("consumes signals").
					WithInputs("signal_in", "start").
					WithOutputs("consumed_signals", "demand_rate").
					WithInitialState(func(state component.State) {
						// Simulate uneven demand
						state.Set("demand_shape", []int{3, 70, 22, 1350})
					}).
					WithActivationFunc(func(this *component.Component) error {
						demandShape := this.State().Get("demand_shape").([]int)
						defer func() {
							this.State().Set("demand_shape", demandShape)
						}()

						if len(demandShape) > 0 {
							// Pop demand rate
							demandRate := demandShape[0]
							demandShape = demandShape[1:]

							this.OutputByName("demand_rate").PutSignals(signal.New(demandRate))
						}

						// Consume signals
						return port.ForwardSignals(this.InputByName("signal_in"), this.OutputByName("consumed_signals"))
					})

				producer.OutputByName("signal_out").PipeTo(counter.InputByName("bypass_in"))
				counter.OutputByName("bypass_out").PipeTo(consumer.InputByName("signal_in"))
				consumer.OutputByName("demand_rate").PipeTo(producer.InputByName("demand_rate"))

				return fmesh.NewWithConfig("fm", &fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10000,
				}).
					WithComponents(producer, counter, consumer)
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("consumer").InputByName("start").PutSignals(signal.New("start demand"))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error) {
				require.NoError(t, err)

				consumedSignals := fm.Components().ByName("consumer").OutputByName("consumed_signals").Buffer()

				// All signals transferred from producer to consumer
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
			tt.assertions(t, fm, runResult.Cycles.CyclesOrNil(), err)
		})
	}
}
