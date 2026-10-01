package export

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
)

func noop(context.Context, *component.Component) error { return nil }

// pricingMesh is src -> dst with descriptions and metadata on every level.
func pricingMesh(t *testing.T, dstFunc component.ActivationFunc) *fmesh.FMesh {
	t.Helper()
	fm := testutil.MustFMesh("pricing",
		fmesh.WithDescription("computes prices"),
		fmesh.WithMeta("env", "prod"),
		fmesh.WithErrorHandlingStrategy(fmesh.IgnoreAll))
	src := testutil.MustComponent("src",
		component.WithDescription("emits prices"),
		component.WithMeta("weight", 2.0),
		component.WithInputs("start"),
		component.WithOutputs("out"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			return this.OutputByName("out").PutPayloads(10)
		}))
	in := testutil.MustInputPort("in", port.WithDescription("prices in"))
	dst := testutil.MustComponent("dst", component.WithActivationFunc(dstFunc))
	require.NoError(t, dst.AttachInputPorts(in))
	require.NoError(t, fm.AddComponents(dst, src))
	require.NoError(t, src.OutputByName("out").PipeTo(dst.InputByName("in")))
	return fm
}

const pricingStructure = `{
	"name": "pricing",
	"description": "computes prices",
	"meta": {"env": "prod"},
	"components": [
		{"name": "dst", "inputs": [{"name": "in", "description": "prices in"}], "outputs": []},
		{"name": "src", "description": "emits prices", "meta": {"weight": 2},
		 "inputs": [{"name": "start"}], "outputs": [{"name": "out"}]}
	],
	"pipes": [
		{"from": {"component": "src", "port": "out"}, "to": {"component": "dst", "port": "in"}}
	]
}`

func TestJSON_Export(t *testing.T) {
	t.Run("structure, descriptions and metadata", func(t *testing.T) {
		got, err := JSON().Export(pricingMesh(t, noop))
		require.NoError(t, err)
		assert.JSONEq(t, pricingStructure, string(got))
	})

	t.Run("empty mesh exports empty lists", func(t *testing.T) {
		got, err := JSON().Export(testutil.MustFMesh("empty"))
		require.NoError(t, err)
		assert.JSONEq(t, `{"name": "empty", "components": [], "pipes": []}`, string(got))
	})

	t.Run("deterministic", func(t *testing.T) {
		a, err := JSON().Export(pricingMesh(t, noop))
		require.NoError(t, err)
		b, err := JSON().Export(pricingMesh(t, noop))
		require.NoError(t, err)
		assert.Equal(t, string(a), string(b))
	})
}

func TestJSON_ExportCycle(t *testing.T) {
	fm := pricingMesh(t, func(context.Context, *component.Component) error { return errors.New("boom") })
	require.NoError(t, fm.ComponentByName("src").InputByName("start").PutSignals(signal.New(1)))
	ri, err := fm.Run(t.Context())
	require.NoError(t, err)
	// Cycle 1: src runs. Cycle 2: dst fails. Cycle 3: nothing runs, the mesh stops.
	require.Equal(t, 3, ri.Cycles.Len())

	t.Run("results per component, a missing one as no input", func(t *testing.T) {
		got, err := JSON().ExportCycle(fm, ri.Cycles.All()[0])
		require.NoError(t, err)
		var doc JSONCycle
		require.NoError(t, json.Unmarshal(got, &doc))

		assert.Equal(t, 1, doc.Number)
		assert.Equal(t, "pricing", doc.Mesh.Name)
		assert.Equal(t, []JSONResult{
			{Component: "dst", Code: component.ActivationCodeNoInput.String()},
			{Component: "src", Code: component.ActivationCodeOK.String(), Activated: true},
		}, doc.Results)
	})

	t.Run("errors are listed", func(t *testing.T) {
		got, err := JSON().ExportCycle(fm, ri.Cycles.All()[1])
		require.NoError(t, err)
		var doc JSONCycle
		require.NoError(t, json.Unmarshal(got, &doc))

		require.Len(t, doc.Results, 2)
		dst := doc.Results[0]
		assert.Equal(t, component.ActivationCodeReturnedError.String(), dst.Code)
		assert.True(t, dst.Activated)
		require.Len(t, dst.Errors, 1)
		assert.Contains(t, dst.Errors[0], "boom")
	})

	t.Run("a nil cycle is an error", func(t *testing.T) {
		_, err := JSON().ExportCycle(fm, nil)
		require.ErrorIs(t, err, ErrNilCycle)
	})

	t.Run("works from an AfterCycle hook", func(t *testing.T) {
		// Streaming: export each cycle as it ends, with no history kept.
		var frames [][]byte
		live := pricingMesh(t, noop)
		live.SetupHooks(func(h *fmesh.Hooks) {
			h.AfterCycle(func(_ context.Context, cc *fmesh.CycleContext) error {
				frame, err := JSON().ExportCycle(cc.FMesh, cc.Cycle)
				frames = append(frames, frame)
				return err
			})
		})
		require.NoError(t, live.ComponentByName("src").InputByName("start").PutSignals(signal.New(1)))
		ri, err := live.Run(t.Context())
		require.NoError(t, err)
		assert.Len(t, frames, ri.Cycles.Len())
	})
}
