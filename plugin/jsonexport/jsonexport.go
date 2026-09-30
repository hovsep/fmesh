package jsonexport

import (
	"encoding/json"
	"errors"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/port"
)

// ErrNotAttached is returned by Export before the plugin is attached to a mesh.
var ErrNotAttached = errors.New("jsonexport: plugin is not attached to a mesh")

// Plugin exports the structure of the mesh it is attached to. One instance
// serves one mesh.
type Plugin struct {
	fm *fmesh.FMesh
}

// New returns a plugin to attach with fmesh.WithPlugins.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string { return "jsonexport" }

// Init attaches the plugin to the mesh.
func (p *Plugin) Init(fm *fmesh.FMesh) error {
	if p.fm != nil && p.fm != fm {
		return errors.New("jsonexport: plugin is already attached to another mesh")
	}
	p.fm = fm
	return nil
}

// Mesh is the exported document.
type Mesh struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
	Components  []Component    `json:"components"`
	Pipes       []Pipe         `json:"pipes"`
}

// Component is one component with its ports.
type Component struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
	Inputs      []Port         `json:"inputs"`
	Outputs     []Port         `json:"outputs"`
}

// Port is one input or output port.
type Port struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// Pipe connects an output port to an input port.
type Pipe struct {
	From Endpoint `json:"from"`
	To   Endpoint `json:"to"`
}

// Endpoint names a port and the component it belongs to.
type Endpoint struct {
	Component string `json:"component"`
	Port      string `json:"port"`
}

// Export returns the structure of the mesh the plugin is attached to as
// indented JSON. Its shape is [Mesh].
func (p *Plugin) Export() ([]byte, error) {
	if p.fm == nil {
		return nil, ErrNotAttached
	}
	return Export(p.fm)
}

// Export returns the structure of fm as indented JSON, for a mesh built without
// the plugin. Its shape is [Mesh].
func Export(fm *fmesh.FMesh) ([]byte, error) {
	b := &builder{}
	if err := fm.Walk(b); err != nil {
		return nil, err
	}
	return json.MarshalIndent(b.mesh, "", "  ")
}

// builder assembles a [Mesh] from a walk.
type builder struct {
	mesh Mesh
}

func (b *builder) VisitMesh(fm *fmesh.FMesh) error {
	b.mesh = Mesh{
		Name:        fm.Name(),
		Description: fm.Description(),
		Meta:        metaOf(fm.Meta()),
		Components:  []Component{},
		Pipes:       []Pipe{},
	}
	return nil
}

func (b *builder) VisitComponent(c *component.Component) error {
	b.mesh.Components = append(b.mesh.Components, Component{
		Name:        c.Name(),
		Description: c.Description(),
		Meta:        metaOf(c.Meta()),
		Inputs:      []Port{},
		Outputs:     []Port{},
	})
	return nil
}

// VisitPort adds the port to the component visited last: Walk visits a
// component's ports right after the component.
func (b *builder) VisitPort(_ *component.Component, p *port.Port) error {
	c := &b.mesh.Components[len(b.mesh.Components)-1]
	exported := Port{Name: p.Name(), Description: p.Description(), Meta: metaOf(p.Meta())}
	if p.IsInput() {
		c.Inputs = append(c.Inputs, exported)
	} else {
		c.Outputs = append(c.Outputs, exported)
	}
	return nil
}

func (b *builder) VisitPipe(from, to *port.Port) error {
	b.mesh.Pipes = append(b.mesh.Pipes, Pipe{From: endpointOf(from), To: endpointOf(to)})
	return nil
}

func endpointOf(p *port.Port) Endpoint {
	e := Endpoint{Port: p.Name()}
	if parent := p.ParentComponent(); parent != nil {
		e.Component = parent.Name()
	}
	return e
}

// metaOf returns nil for an empty store, so omitempty drops it.
func metaOf(m *meta.Meta) map[string]any {
	if m.IsEmpty() {
		return nil
	}
	return m.All()
}
