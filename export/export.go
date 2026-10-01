// Package export defines how a mesh is exported, and provides the JSON
// exporter. Diagram formats (DOT, Mermaid, D2, PlantUML) implement the same
// interface in github.com/hovsep/fmesh-export.
package export

import (
	"errors"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/cycle"
)

// ErrNilCycle is returned by ExportCycle when the cycle is nil.
var ErrNilCycle = errors.New("export: cycle is nil")

// Exporter renders a mesh in one format. An exporter only reads the mesh, and
// its output is deterministic: equal meshes export byte-identical documents.
type Exporter interface {
	// Export renders the mesh structure: components, ports and pipes.
	Export(fm *fmesh.FMesh) ([]byte, error)

	// ExportCycle renders the structure together with the results of one
	// cycle of a run, e.g. one from RuntimeInfo.Cycles or from an AfterCycle
	// hook. A component with no result in the cycle had no input. A nil cycle
	// is ErrNilCycle.
	ExportCycle(fm *fmesh.FMesh, c *cycle.Cycle) ([]byte, error)
}
