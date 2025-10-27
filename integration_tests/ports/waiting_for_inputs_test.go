package ports

import (
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_WaitingForInputs(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error)
	}{
		{
			name: "waiting for longer chain",
			setupFM: func() *fmesh.FMesh {
				getDoubler := func(name string) *component.Component {
					return component.New(name).
						WithDescription("This component just doubles the input").
						WithInputs("i1").
						WithOutputs("o1").
						WithActivationFunc(func(this *component.Component) error {
							inputNum := this.InputByName("i1").FirstSignalPayloadOrDefault(0)

							this.OutputByName("o1").PutSignals(signal.New(inputNum.(int) * 2))
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
					WithActivationFunc(func(this *component.Component) error {
						if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
							return component.NewErrWaitForInputs(true)
						}

						inputNum1 := this.InputByName("i1").FirstSignalPayloadOrDefault(0)
						inputNum2 := this.InputByName("i2").FirstSignalPayloadOrDefault(0)

						this.OutputByName("o1").PutSignals(signal.New(inputNum1.(int) + inputNum2.(int)))
						return nil
					})

				// This chain consist of 3 components: d1->d2->d3
				d1.OutputByName("o1").PipeTo(d2.InputByName("i1"))
				d2.OutputByName("o1").PipeTo(d3.InputByName("i1"))

				// This chain has only 2: d4->d5
				d4.OutputByName("o1").PipeTo(d5.InputByName("i1"))

				// Both chains go into summator
				d3.OutputByName("o1").PipeTo(s.InputByName("i1"))
				d5.OutputByName("o1").PipeTo(s.InputByName("i2"))

				return fmesh.NewWithConfig("fm", &fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           5,
				}).
					WithComponents(d1, d2, d3, d4, d5, s)
			},
			setInputs: func(fm *fmesh.FMesh) {
				// Put 1 signal to each chain so they start in the same cycle
				fm.Components().ByName("d1").InputByName("i1").PutSignals(signal.New(1))
				fm.Components().ByName("d4").InputByName("i1").PutSignals(signal.New(2))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error) {
				require.NoError(t, err)
				result, err := fm.Components().ByName("sum").OutputByName("o1").FirstSignalPayload()
				require.NoError(t, err)
				assert.Equal(t, 16, result.(int))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := tt.setupFM()
			tt.setInputs(fm)
			runResult, err := fm.Run()
			tt.assertions(t, fm, runResult.Cycles.AllAsSliceOrNil(), err)
		})
	}
}
