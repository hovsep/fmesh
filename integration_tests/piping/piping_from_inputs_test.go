package integration_tests

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_PipingFromInput(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error)
	}{
		{
			name: "observer pattern",
			setupFM: func() *fmesh.FMesh {
				adder := component.New("adder").
					WithDescription("adds i1 and i2").
					WithInputsIndexed("i", 1, 2).
					WithOutputs("out").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						i1, i2 := inputs.ByName("i1").Signals().FirstPayload().(int), inputs.ByName("i2").Signals().FirstPayload().(int)
						outputs.ByName("out").PutSignals(signal.New(i1 + i2))
						return nil
					})

				multiplier := component.New("multiplier").
					WithDescription("multiplies i1 by 10").
					WithInputs("i1").
					WithOutputs("out").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						i1 := inputs.ByName("i1").Signals().FirstPayload().(int)
						outputs.ByName("out").PutSignals(signal.New(i1 * 10))
						return nil
					})

				logger := component.New("logger").
					WithDescription("logs all input signals").
					WithInputs("in").
					WithOutputs("log").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						for _, sig := range inputs.ByName("in").Signals() {
							outputs.ByName("log").PutSignals(signal.New(fmt.Sprintf("LOGGED SIGNAL: %v", sig.Payload())))
						}
						return nil
					})

				fm := fmesh.New("fm with observer").
					WithDescription("In this f-mesh adder receives 2 numbers, adds them and passes to multiplier. "+
						"The logger component is connected to adder's inputs, so it can observe them"+
						"The cool thing is logger does not need multiple input ports to observe multiple ports of other component").
					WithComponents(adder, multiplier, logger)

				adder.Outputs().ByName("out").PipeTo(multiplier.Inputs().ByName("i1"))
				adder.Inputs().ByNames("i1", "i2").PipeTo(logger.Inputs().ByName("in"))
				multiplier.Inputs().ByName("i1").PipeTo(logger.Inputs().ByName("in"))

				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("adder").Inputs().ByName("i1").PutSignals(signal.New(4))
				fm.Components().ByName("adder").Inputs().ByName("i2").PutSignals(signal.New(5))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)
				m := fm.Components().ByName("multiplier")
				l := fm.Components().ByName("logger")
				assert.True(t, m.Outputs().ByName("out").HasSignals())
				assert.Equal(t, 90, m.Outputs().ByName("out").Signals().FirstPayload().(int))

				assert.True(t, fm.Components().ByName("logger").Outputs().ByName("log").HasSignals())
				assert.Len(t, l.Outputs().ByName("log").Signals(), 3)
			},
		},
		{
			name: "observing component which waits for inputs",
			setupFM: func() *fmesh.FMesh {
				starter := component.New("starter").
					WithDescription("This component just starts the whole f-mesh").
					WithInputs("start").
					WithOutputsIndexed("o", 1, 2).
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						//Activate downstream components
						outputs.PutSignals(inputs.ByName("start").Signals().First())
						return nil
					})

				incr1 := component.New("incr1").
					WithDescription("Increments the input").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.PutSignals(signal.New(1 + inputs.ByName("i1").Signals().FirstPayload().(int)))
						return nil
					})

				incr2 := component.New("incr2").
					WithDescription("Increments the input").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.PutSignals(signal.New(1 + inputs.ByName("i1").Signals().FirstPayload().(int)))
						return nil
					})

				doubler := component.New("doubler").
					WithDescription("Doubles the input").
					WithInputs("i1").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.PutSignals(signal.New(2 * inputs.ByName("i1").Signals().FirstPayload().(int)))
						return nil
					})

				agg := component.New("result_aggregator").
					WithDescription("Adds 2 inputs (only when both are available)").
					WithInputsIndexed("i", 1, 2).
					WithOutputs("result").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						if !inputs.ByNames("i1", "i2").AllHaveSignals() {
							return component.NewErrWaitForInputs(true)
						}
						i1 := inputs.ByName("i1").Signals().FirstPayload().(int)
						i2 := inputs.ByName("i2").Signals().FirstPayload().(int)
						outputs.PutSignals(signal.New(i1 + i2))
						return nil
					})

				observer := component.New("obsrv").
					WithDescription("Observes inputs of result aggregator").
					WithInputsIndexed("i", 1, 2).
					WithOutputs("log").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						outputs.ByName("log").PutSignals(inputs.ByNames("i1", "i2").Signals()...)
						return nil
					})

				fm := fmesh.New("observer").WithComponents(starter, incr1, incr2, doubler, agg, observer)

				starter.Outputs().ByName("o1").PipeTo(incr1.Inputs().ByName("i1"))
				starter.Outputs().ByName("o2").PipeTo(incr2.Inputs().ByName("i1"))
				incr1.Outputs().ByName("o1").PipeTo(doubler.Inputs().ByName("i1"))
				doubler.Outputs().ByName("o1").PipeTo(agg.Inputs().ByName("i1"))
				incr2.Outputs().ByName("o1").PipeTo(agg.Inputs().ByName("i2"))
				agg.Inputs().ByName("i1").PipeTo(observer.Inputs().ByName("i1"))
				agg.Inputs().ByName("i2").PipeTo(observer.Inputs().ByName("i2"))
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("starter").Inputs().PutSignals(signal.New(10))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)

				//Multiplier result
				assert.Equal(t, 33, fm.Components().ByName("result_aggregator").Outputs().ByName("result").Signals().FirstPayload())

				//Observed signals
				assert.Len(t, fm.Components().ByName("obsrv").Outputs().ByName("log").Signals(), 2)
				assert.Contains(t, fm.Components().ByName("obsrv").Outputs().ByName("log").Signals().AllPayloads(), 11)
				assert.Contains(t, fm.Components().ByName("obsrv").Outputs().ByName("log").Signals().AllPayloads(), 22)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.setupFM()
			tt.setInputs(fm)
			cycles, err := fm.Run()
			tt.assertions(t, fm, cycles, err)
		})
	}
}
