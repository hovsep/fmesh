package jsonexport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/internal/testutil"
	"github.com/hovsep/fmesh/port"
)

func noop(context.Context, *component.Component) error { return nil }

func TestPlugin_Export(t *testing.T) {
	t.Run("exports structure, descriptions and metadata", func(t *testing.T) {
		exporter := New()
		fm := testutil.MustFMesh("pricing",
			fmesh.WithDescription("computes prices"),
			fmesh.WithMeta("env", "prod"),
			fmesh.WithPlugins(exporter))

		src := testutil.MustComponent("src",
			component.WithDescription("emits prices"),
			component.WithMeta("weight", 2.0),
			component.WithOutputs("out"),
			component.WithActivationFunc(noop))
		in := testutil.MustInputPort("in", port.WithDescription("prices in"))
		dst := testutil.MustComponent("dst", component.WithActivationFunc(noop))
		require.NoError(t, dst.AttachInputPorts(in))
		require.NoError(t, fm.AddComponents(dst, src))
		require.NoError(t, src.OutputByName("out").PipeTo(dst.InputByName("in")))

		got, err := exporter.Export()
		require.NoError(t, err)

		assert.JSONEq(t, `{
			"name": "pricing",
			"description": "computes prices",
			"meta": {"env": "prod"},
			"components": [
				{"name": "dst", "inputs": [{"name": "in", "description": "prices in"}], "outputs": []},
				{"name": "src", "description": "emits prices", "meta": {"weight": 2},
				 "inputs": [], "outputs": [{"name": "out"}]}
			],
			"pipes": [
				{"from": {"component": "src", "port": "out"}, "to": {"component": "dst", "port": "in"}}
			]
		}`, string(got))
	})

	t.Run("empty mesh exports empty lists", func(t *testing.T) {
		exporter := New()
		testutil.MustFMesh("empty", fmesh.WithPlugins(exporter))

		got, err := exporter.Export()
		require.NoError(t, err)
		assert.JSONEq(t, `{"name": "empty", "components": [], "pipes": []}`, string(got))
	})

	t.Run("not attached", func(t *testing.T) {
		_, err := New().Export()
		require.ErrorIs(t, err, ErrNotAttached)
	})
}

func TestPlugin_Init(t *testing.T) {
	// One instance keeps one mesh; a second mesh would silently steal it.
	exporter := New()
	testutil.MustFMesh("first", fmesh.WithPlugins(exporter))

	_, err := fmesh.New("second", fmesh.WithPlugins(exporter))
	require.ErrorContains(t, err, "already attached")
}
