package computation

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/assert"
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
					WithActivationFunc(func(this *component.Component) error {
						num := this.InputByName("num").FirstSignalPayloadOrNil()
						this.OutputByName("res").PutSignals(signal.New(num.(int) + 2))
						return nil
					})

				c2 := component.New("c2").
					WithDescription("multiplies by 3").
					WithInputs("num").
					WithOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						num := this.InputByName("num").FirstSignalPayloadOrDefault(0)
						this.OutputByName("res").PutSignals(signal.New(num.(int) * 3))
						return nil
					})

				c1.OutputByName("res").PipeTo(c2.InputByName("num"))
				return fmesh.NewWithConfig("fm", &fmesh.Config{
					ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
					CyclesLimit:           10,
				}).WithComponents(c1, c2)
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
		fm := fmesh.NewWithConfig("hello world", &fmesh.Config{
			ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
			CyclesLimit:           10,
		}).
			WithComponents(
				component.New("concat").
					WithInputs("i1", "i2").
					WithOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						word1 := this.InputByName("i1").FirstSignalPayloadOrDefault("").(string)
						word2 := this.InputByName("i2").FirstSignalPayloadOrDefault("").(string)
						this.OutputByName("res").PutSignals(signal.New(word1 + word2))
						return nil
					}),
				component.New("case").
					WithInputs("i1").
					WithOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						inputString := this.InputByName("i1").FirstSignalPayloadOrDefault("").(string)
						this.OutputByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
						return nil
					}))

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
