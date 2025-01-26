package integration_tests

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
)

func Test_Math(t *testing.T) {
	tests := []struct {
		name       string
		setupFM    func() *fmesh.FMesh
		setInputs  func(fm *fmesh.FMesh)
		assertions func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error)
	}{
		{
			name: "add and multiply",
			setupFM: func() *fmesh.FMesh {
				c1 := component.New("c1").
					WithDescription("adds 2 to the input").
					WithInputs("num").
					WithOutputs("res").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
						num := inputs.ByName("num").FirstSignalPayloadOrNil()
						outputs.ByName("res").PutSignals(signal.New(num.(int) + 2))
						return nil
					})

				c2 := component.New("c2").
					WithDescription("multiplies by 3").
					WithInputs("num").
					WithOutputs("res").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
						num := inputs.ByName("num").FirstSignalPayloadOrDefault(0)
						outputs.ByName("res").PutSignals(signal.New(num.(int) * 3))
						return nil
					})

				c1.OutputByName("res").PipeTo(c2.InputByName("num"))
				return fmesh.New("fm").WithComponents(c1, c2).WithConfig(fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10,
				})
			},
			setInputs: func(fm *fmesh.FMesh) {
				fm.Components().ByName("c1").InputByName("num").PutSignals(signal.New(32))
			},
			assertions: func(t *testing.T, fm *fmesh.FMesh, cycles cycle.Cycles, err error) {
				assert.NoError(t, err)
				assert.Len(t, cycles, 3)

				resultSignals := fm.Components().ByName("c2").OutputByName("res").Buffer()
				sig, err := resultSignals.FirstPayload()
				assert.NoError(t, err)
				assert.Len(t, resultSignals.SignalsOrNil(), 1)
				assert.Equal(t, 102, sig.(int))
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

func Test_Readme(t *testing.T) {
	t.Run("readme test", func(t *testing.T) {
		fm := fmesh.New("hello world").
			WithComponents(
				component.New("concat").
					WithInputs("i1", "i2").
					WithOutputs("res").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
						word1 := inputs.ByName("i1").FirstSignalPayloadOrDefault("").(string)
						word2 := inputs.ByName("i2").FirstSignalPayloadOrDefault("").(string)

						outputs.ByName("res").PutSignals(signal.New(word1 + word2))
						return nil
					}),
				component.New("case").
					WithInputs("i1").
					WithOutputs("res").
					WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
						inputString := inputs.ByName("i1").FirstSignalPayloadOrDefault("").(string)

						outputs.ByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
						return nil
					})).
			WithConfig(fmesh.Config{
				ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
				CyclesLimit:           10,
			})

		fm.Components().ByName("concat").Outputs().ByName("res").PipeTo(
			fm.Components().ByName("case").Inputs().ByName("i1"),
		)

		// Init inputs
		fm.Components().ByName("concat").Inputs().ByName("i1").PutSignals(signal.New("hello "))
		fm.Components().ByName("concat").Inputs().ByName("i2").PutSignals(signal.New("world !"))

		// Run the mesh
		_, err := fm.Run()

		// Check for errors
		if err != nil {
			fmt.Println("F-Mesh returned an error")
			os.Exit(1)
		}

		//Extract results
		results := fm.Components().ByName("case").Outputs().ByName("res").FirstSignalPayloadOrNil()
		fmt.Printf("Result is :%v", results)
		assert.Equal(t, "HELLO WORLD !", results)
	})
}
