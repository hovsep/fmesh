package export

import "github.com/hovsep/fmesh"

type Exporter interface {
	Export(fm *fmesh.FMesh) []byte
}
