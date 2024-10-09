package export

import "github.com/hovsep/fmesh"

// Exporter is the common interface for all formats
type Exporter interface {
	Export(fm *fmesh.FMesh) ([]byte, error)
}
