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
		c1, err := component.New("c1",
			component.WithInputs("num"),
			component.WithOutputs("res"),
			component.WithDescription("adds 2 to the input"),
			component.WithActivationFunc(func(this *component.Component) error {
				num := this.InputByName("num").Signals().FirstPayloadOrNil()
				return this.OutputByName("res").PutSignals(signal.New(num.(int) + 2))
			}),
		)
		require.NoError(t, err)

		c2, err := component.New("c2",
			component.WithInputs("num"),
			component.WithOutputs("res"),
			component.WithDescription("multiplies by 3"),
			component.WithActivationFunc(func(this *component.Component) error {
				num := this.InputByName("num").Signals().FirstPayloadOrDefault(0)
				return this.OutputByName("res").PutSignals(signal.New(num.(int) * 3))
			}),
		)
		require.NoError(t, err)

		require.NoError(t, c1.OutputByName("res").PipeTo(c2.InputByName("num")))

		fm, err := fmesh.New("fm")
		require.NoError(t, err)
		require.NoError(t, fm.AddComponents(c1)) // Oops, we forgot to add c2
		_, err = fm.Run()
		require.Error(t, err)
	})
}
