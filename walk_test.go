package fmesh

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
)

// recordingVisitor logs every visit and fails on the visit named in failOn.
type recordingVisitor struct {
	events []string
	failOn string
}

func (r *recordingVisitor) record(event string) error {
	r.events = append(r.events, event)
	if event == r.failOn {
		return errors.New("stop")
	}
	return nil
}

func (r *recordingVisitor) VisitMesh(fm *FMesh) error { return r.record("mesh " + fm.Name()) }

func (r *recordingVisitor) VisitComponent(c *component.Component) error {
	return r.record("component " + c.Name())
}

func (r *recordingVisitor) VisitPort(c *component.Component, p *port.Port) error {
	return r.record("port " + c.Name() + "." + p.Name())
}

func (r *recordingVisitor) VisitPipe(from, to *port.Port) error {
	return r.record("pipe " + from.ParentComponent().Name() + "." + from.Name() +
		" -> " + to.ParentComponent().Name() + "." + to.Name())
}

func walkTestMesh(t *testing.T) *FMesh {
	t.Helper()
	noop := component.WithActivationFunc(func(context.Context, *component.Component) error { return nil })
	// Added out of name order: the walk must not depend on insertion order.
	b := mustNewComponent("b", component.WithInputs("in"), noop)
	a := mustNewComponent("a", component.WithInputs("z", "y"), component.WithOutputs("out"), noop)
	fm := mustNewFMesh("m")
	require.NoError(t, fm.AddComponents(b, a))
	require.NoError(t, a.OutputByName("out").PipeTo(b.InputByName("in"), a.InputByName("y")))
	return fm
}

func TestFMesh_Walk(t *testing.T) {
	t.Run("visits in a fixed order", func(t *testing.T) {
		v := &recordingVisitor{}
		require.NoError(t, walkTestMesh(t).Walk(v))

		assert.Equal(t, []string{
			"mesh m",
			"component a", "port a.y", "port a.z", "port a.out",
			"component b", "port b.in",
			"pipe a.out -> b.in", "pipe a.out -> a.y",
		}, v.events)
	})

	t.Run("stops at the first error", func(t *testing.T) {
		v := &recordingVisitor{failOn: "port a.z"}
		require.Error(t, walkTestMesh(t).Walk(v))
		assert.Equal(t, "port a.z", v.events[len(v.events)-1])
	})

	t.Run("empty mesh visits only the mesh", func(t *testing.T) {
		v := &recordingVisitor{}
		require.NoError(t, mustNewFMesh("empty").Walk(v))
		assert.Equal(t, []string{"mesh empty"}, v.events)
	})
}
