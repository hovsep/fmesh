package export

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/cycle"
)

// Exporter is the common interface for all formats
type Exporter interface {
	// Export returns f-mesh representation in some format
	Export(fm *fmesh.FMesh) ([]byte, error)

	// ExportWithCycles returns the f-mesh state representation in each activation cycle
	ExportWithCycles(fm *fmesh.FMesh, activationCycles cycle.Cycles) ([][]byte, error)
}
