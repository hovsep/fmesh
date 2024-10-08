package ports

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_WaitingForInputs(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error)
	}{
		{
			name: "waiting for longer chain",
			setupFM: func() *fmesh.FMesh {
				getDoubler := func(name string) *component.Component {
					return component.New(name).
						WithDescription("This component just doubles the input").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							inputNum := inputs.ByName("i1").Signals().FirstPayload().(int)
							outputs.ByName("o1").PutSignals(signal.New(inputNum * 2))
							return nil
						})
				}

				d1 := getDoubler("d1")
				d2 := getDoubler("d2")
				d3 := getDoubler("d3")
				d4 := getDoubler("d4")
				d5 := getDoubler("d5")

				s := component.New("sum").
					WithDescription("This component just sums 2 inputs").
					WithInputs("i1", "i2").
					WithOutputs("o1").
					WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
						if !inputs.ByNames("i1", "i2").AllHaveSignals() {
							return component.NewErrWaitForInputs(true)
						}

						inputNum1 := inputs.ByName("i1").Signals().FirstPayload().(int)
						inputNum2 := inputs.ByName("i2").Signals().FirstPayload().(int)
						outputs.ByName("o1").PutSignals(signal.New(inputNum1 + inputNum2))
						return nil
					})

				//This chain consist of 3 components: d1->d2->d3
				d1.Outputs().ByName("o1").PipeTo(d2.Inputs().ByName("i1"))
				d2.Outputs().ByName("o1").PipeTo(d3.Inputs().ByName("i1"))

				//This chain has only 2: d4->d5
				d4.Outputs().ByName("o1").PipeTo(d5.Inputs().ByName("i1"))

				//Both chains go into summator
				d3.Outputs().ByName("o1").PipeTo(s.Inputs().ByName("i1"))
				d5.Outputs().ByName("o1").PipeTo(s.Inputs().ByName("i2"))

				return fmesh.New("fm").
					WithComponents(d1, d2, d3, d4, d5, s).
					WithConfig(fmesh.Config{
						ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
						CyclesLimit:           5,
					})

			},
			setInputs: func(fm *fmesh.FMesh) {
				//Put 1 signal to each chain so they start in the same cycle
				fm.Components().ByName("d1").Inputs().ByName("i1").PutSignals(signal.New(1))
				fm.Components().ByName("d4").Inputs().ByName("i1").PutSignals(signal.New(2))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Collection, err error) {
				assert.NoError(t, err)
				result := fm.Components().ByName("sum").Outputs().ByName("o1").Signals().FirstPayload().(int)
				assert.Equal(t, 16, result)
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
