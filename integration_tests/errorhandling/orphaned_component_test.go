package errorhandling

import (
	"testing"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"github.com/stretchr/testify/require"
)

func Test_AllComponentsMustBeRegistered(t *testing.T) {
	t.Run("orphaned component", func(t *testing.T) {
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
		fm := fmesh.New("fm").WithComponents(c1) // Oops, we forgot to add c2
		_, err := fm.Run()
		require.Error(t, err)
	})
}
