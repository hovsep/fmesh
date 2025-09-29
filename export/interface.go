package export

import (
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/cycle"
)

// Exporter defines an interface for exporting an FMesh in different formats.
type Exporter interface {
	// Export serializes the FMesh into a specific format.
	Export(fm *fmesh.FMesh) ([]byte, error)

	// ExportWithCycles serializes the FMesh state at each activation cycle.
	ExportWithCycles(fm *fmesh.FMesh, activationCycles cycle.Cycles) ([][]byte, error)
}
