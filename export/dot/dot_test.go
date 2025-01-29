package dot

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
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
						WithActivationFunc(func(this *component.Component) error {
							//The activation func can be even empty, does not affect export
							return nil
						})

					multiplier := component.New("multiplier").
						WithDescription("This component multiplies number by 3").
						WithInputs("num").
						WithOutputs("result").
						WithActivationFunc(func(this *component.Component) error {
							//The activation func can be even empty, does not affect export
							return nil
						})

					adder.OutputByName("result").PipeTo(multiplier.InputByName("num"))

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

func Test_dotExporter_ExportWithCycles(t *testing.T) {
	type args struct {
		fm *fmesh.FMesh
	}
	tests := []struct {
		name       string
		args       args
		assertions func(t *testing.T, data [][]byte, err error)
	}{
		{
			name: "happy path",
			args: args{
				fm: func() *fmesh.FMesh {
					adder := component.New("adder").
						WithDescription("This component adds 2 numbers").
						WithInputs("num1", "num2").
						WithOutputs("result").
						WithActivationFunc(func(this *component.Component) error {
							num1, err := this.InputByName("num1").FirstSignalPayload()
							if err != nil {
								return err
							}

							num2, err := this.InputByName("num2").FirstSignalPayload()
							if err != nil {
								return err
							}

							this.OutputByName("result").PutSignals(signal.New(num1.(int) + num2.(int)))
							return nil
						})

					multiplier := component.New("multiplier").
						WithDescription("This component multiplies number by 3").
						WithInputs("num").
						WithOutputs("result").
						WithActivationFunc(func(this *component.Component) error {
							num, err := this.InputByName("num").FirstSignalPayload()
							if err != nil {
								return err
							}
							this.OutputByName("result").PutSignals(signal.New(num.(int) * 3))
							return nil
						})

					adder.OutputByName("result").PipeTo(multiplier.InputByName("num"))

					fm := fmesh.New("fm").
						WithDescription("This f-mesh has just one component").
						WithComponents(adder, multiplier)

					adder.InputByName("num1").PutSignals(signal.New(15))
					adder.InputByName("num2").PutSignals(signal.New(12))

					return fm
				}(),
			},
			assertions: func(t *testing.T, data [][]byte, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, data)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cycles, err := tt.args.fm.Run()
			assert.NoError(t, err)

			exporter := NewDotExporter()

			got, err := exporter.ExportWithCycles(tt.args.fm, cycles)
			if tt.assertions != nil {
				tt.assertions(t, got, err)
			}
		})
	}
}
