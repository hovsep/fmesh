package export

import (
	"encoding/json"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/meta"
	"github.com/hovsep/fmesh/port"
)

var _ Exporter = (*JSONExporter)(nil)

// JSONExporter exports a mesh as indented JSON. Unmarshal its output into
// JSONMesh or JSONCycle to read it back.
type JSONExporter struct{}

// JSON returns the JSON exporter.
func JSON() *JSONExporter {
	return &JSONExporter{}
}

// JSONMesh is the document Export produces.
type JSONMesh struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Meta        map[string]any  `json:"meta,omitempty"`
	Components  []JSONComponent `json:"components"`
	Pipes       []JSONPipe      `json:"pipes"`
}

// JSONComponent is one component with its ports.
type JSONComponent struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
	Inputs      []JSONPort     `json:"inputs"`
	Outputs     []JSONPort     `json:"outputs"`
}

// JSONPort is one input or output port.
type JSONPort struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// JSONPipe connects an output port to an input port.
type JSONPipe struct {
	From JSONEndpoint `json:"from"`
	To   JSONEndpoint `json:"to"`
}

// JSONEndpoint names a port and the component it belongs to.
type JSONEndpoint struct {
	Component string `json:"component"`
	Port      string `json:"port"`
}

// JSONCycle is the document ExportCycle produces: the structure plus one
// result per component, in component name order.
type JSONCycle struct {
	Number  int          `json:"number"`
	Mesh    JSONMesh     `json:"mesh"`
	Results []JSONResult `json:"results"`
}

// JSONResult is how one component's activation ended in a cycle.
type JSONResult struct {
	Component string   `json:"component"`
	Code      string   `json:"code"`
	Activated bool     `json:"activated"`
	Errors    []string `json:"errors,omitempty"`
}

// Export returns the mesh structure as indented JSON, shaped as JSONMesh.
func (e *JSONExporter) Export(fm *fmesh.FMesh) ([]byte, error) {
	mesh, err := structureOf(fm)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(mesh, "", "  ")
}

// ExportCycle returns the structure and the results of c as indented JSON,
// shaped as JSONCycle.
func (e *JSONExporter) ExportCycle(fm *fmesh.FMesh, c *cycle.Cycle) ([]byte, error) {
	if c == nil {
		return nil, ErrNilCycle
	}
	mesh, err := structureOf(fm)
	if err != nil {
		return nil, err
	}
	doc := JSONCycle{Number: c.Number(), Mesh: mesh, Results: make([]JSONResult, 0, len(mesh.Components))}
	for _, comp := range mesh.Components {
		doc.Results = append(doc.Results, resultOf(comp.Name, c.ActivationResults().ByName(comp.Name)))
	}
	return json.MarshalIndent(doc, "", "  ")
}

// resultOf treats a missing result as NoInput, which is what it means.
func resultOf(name string, ar *component.ActivationResult) JSONResult {
	if ar == nil {
		return JSONResult{Component: name, Code: component.ActivationCodeNoInput.String()}
	}
	r := JSONResult{Component: name, Code: ar.Code().String(), Activated: ar.Activated()}
	for _, err := range ar.ActivationErrors() {
		r.Errors = append(r.Errors, err.Error())
	}
	return r
}

func structureOf(fm *fmesh.FMesh) (JSONMesh, error) {
	b := &jsonBuilder{}
	if err := fm.Walk(b); err != nil {
		return JSONMesh{}, err
	}
	return b.mesh, nil
}

// jsonBuilder assembles a JSONMesh from a walk.
type jsonBuilder struct {
	mesh JSONMesh
}

func (b *jsonBuilder) VisitMesh(fm *fmesh.FMesh) error {
	b.mesh = JSONMesh{
		Name:        fm.Name(),
		Description: fm.Description(),
		Meta:        metaOf(fm.Meta()),
		Components:  []JSONComponent{},
		Pipes:       []JSONPipe{},
	}
	return nil
}

func (b *jsonBuilder) VisitComponent(c *component.Component) error {
	b.mesh.Components = append(b.mesh.Components, JSONComponent{
		Name:        c.Name(),
		Description: c.Description(),
		Meta:        metaOf(c.Meta()),
		Inputs:      []JSONPort{},
		Outputs:     []JSONPort{},
	})
	return nil
}

// VisitPort adds the port to the component visited last: Walk visits a
// component's ports right after the component.
func (b *jsonBuilder) VisitPort(_ *component.Component, p *port.Port) error {
	c := &b.mesh.Components[len(b.mesh.Components)-1]
	exported := JSONPort{Name: p.Name(), Description: p.Description(), Meta: metaOf(p.Meta())}
	if p.IsInput() {
		c.Inputs = append(c.Inputs, exported)
	} else {
		c.Outputs = append(c.Outputs, exported)
	}
	return nil
}

func (b *jsonBuilder) VisitPipe(from, to *port.Port) error {
	b.mesh.Pipes = append(b.mesh.Pipes, JSONPipe{From: endpointOf(from), To: endpointOf(to)})
	return nil
}

func endpointOf(p *port.Port) JSONEndpoint {
	e := JSONEndpoint{Port: p.Name()}
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
