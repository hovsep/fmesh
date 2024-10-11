package export

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/cycle"
)

// Exporter is the common interface for all formats
type Exporter interface {
	// Export returns f-mesh representation in some format
	Export(fm *fmesh.FMesh) ([]byte, error)

	// ExportWithCycles returns representations of f-mesh during multiple cycles
	ExportWithCycles(fm *fmesh.FMesh, cycles cycle.Collection) ([][]byte, error)
}
