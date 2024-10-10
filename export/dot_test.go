package export

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_dotExporter_Export(t *testing.T) {
	type args struct {
		fm *fmesh.FMesh
	}
	tests := []struct {
		name       string
		args       args
		assertions func(t *testing.T, data []byte, err error)
	}{
		{
			name: "empty f-mesh",
			args: args{
				fm: fmesh.New("fm"),
			},
			assertions: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				assert.Empty(t, data)
			},
		},
		{
			name: "happy path",
			args: args{
				fm: func() *fmesh.FMesh {
					adder := component.New("adder").
						WithDescription("This component adds 2 numbers").
						WithInputs("num1", "num2").
						WithOutputs("result").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							//The activation func can be even empty, does not affect export
							return nil
						})

					multiplier := component.New("multiplier").
						WithDescription("This component multiplies number by 3").
						WithInputs("num").
						WithOutputs("result").
						WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
							//The activation func can be even empty, does not affect export
							return nil
						})

					adder.Outputs().ByName("result").PipeTo(multiplier.Inputs().ByName("num"))

					fm := fmesh.New("fm").
						WithDescription("This f-mesh has just one component").
						WithComponents(adder, multiplier)
					return fm
				}(),
			},
			assertions: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, data)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := NewDotExporter()

			got, err := exporter.Export(tt.args.fm)
			if tt.assertions != nil {
				tt.assertions(t, got, err)
			}
		})
	}
}
