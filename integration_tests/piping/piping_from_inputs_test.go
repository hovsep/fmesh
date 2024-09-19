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
				adder := component.NewComponent("adder").
					WithDescription("adds i1 and i2").
					WithInputs("i1", "i2").
					WithOutputs("out").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						i1, i2 := inputs.ByName("i1").Signals().FirstPayload().(int), inputs.ByName("i2").Signals().FirstPayload().(int)
						outputs.ByName("out").PutSignals(signal.New(i1 + i2))
						return nil
					})

				multiplier := component.NewComponent("multiplier").
					WithDescription("multiplies i1 by 10").
					WithInputs("i1").
					WithOutputs("out").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						i1 := inputs.ByName("i1").Signals().FirstPayload().(int)
						outputs.ByName("out").PutSignals(signal.New(i1 * 10))
						return nil
					})

				logger := component.NewComponent("logger").
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
