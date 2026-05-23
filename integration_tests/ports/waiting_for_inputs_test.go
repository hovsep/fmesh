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
					return mustComponent(name,
						component.WithInputs("i1"),
						component.WithOutputs("o1"),
						component.WithActivationFunc(func(this *component.Component) error {
							inputNum := this.InputByName("i1").Signals().FirstPayloadOrDefault(0)
							return this.OutputByName("o1").PutSignals(signal.New(inputNum.(int) * 2))
						}),
					).WithDescription("This component just doubles the input")
				}

				d1 := getDoubler("d1")
				d2 := getDoubler("d2")
				d3 := getDoubler("d3")
				d4 := getDoubler("d4")
				d5 := getDoubler("d5")

				s := mustComponent("sum",
					component.WithInputs("i1", "i2"),
					component.WithOutputs("o1"),
					component.WithActivationFunc(func(this *component.Component) error {
						if !this.Inputs().ByNames("i1", "i2").AllHaveSignals() {
							return component.ErrWaitingForInputsKeep
						}

						inputNum1 := this.InputByName("i1").Signals().FirstPayloadOrDefault(0)
						inputNum2 := this.InputByName("i2").Signals().FirstPayloadOrDefault(0)

						return this.OutputByName("o1").PutSignals(signal.New(inputNum1.(int) + inputNum2.(int)))
					}),
				).WithDescription("This component just sums 2 inputs")

				// This chain consists of 3 components: d1->d2->d3
				if err := d1.OutputByName("o1").PipeTo(d2.InputByName("i1")); err != nil {
					panic(err)
				}
				if err := d2.OutputByName("o1").PipeTo(d3.InputByName("i1")); err != nil {
					panic(err)
				}

				// This chain has only 2: d4->d5
				if err := d4.OutputByName("o1").PipeTo(d5.InputByName("i1")); err != nil {
					panic(err)
				}

				// Both chains go into summator
				if err := d3.OutputByName("o1").PipeTo(s.InputByName("i1")); err != nil {
					panic(err)
				}
				if err := d5.OutputByName("o1").PipeTo(s.InputByName("i2")); err != nil {
					panic(err)
				}

				fm := mustFMesh("fm", fmesh.WithConfig(&fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           5,
				}))
				if err := fm.AddComponents(d1, d2, d3, d4, d5, s); err != nil {
					panic(err)
				}
				return fm
			},
			setInputs: func(fm *fmesh.FMesh) {
				// Put 1 signal to each chain so they start in the same cycle
				if err := fm.Components().ByName("d1").InputByName("i1").PutSignals(signal.New(1)); err != nil {
					panic(err)
				}
				if err := fm.Components().ByName("d4").InputByName("i1").PutSignals(signal.New(2)); err != nil {
					panic(err)
				}
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error) {
				require.NoError(t, err)
				result, err := fm.Components().ByName("sum").OutputByName("o1").Signals().FirstPayload()
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
			cycles, cycleErr := runResult.Cycles.All()
			require.NoError(t, cycleErr)
			tt.assertions(t, fm, cycles, err)
		})
	}
}
