package export

import (
	"bytes"
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/lucasepe/dot"
)

type dotExporter struct {
}

func NewDotExporter() Exporter {
	return &dotExporter{}
}

// Export returns the f-mesh represented as digraph in DOT language
func (d *dotExporter) Export(fm *fmesh.FMesh) []byte {
	// Setup graph
	graph := dot.NewGraph(dot.Directed)
	graph.
		Attr("layout", "dot").
		Attr("splines", "ortho")

	for componentName, c := range fm.Components() {
		// Component subgraph (wrapper)
		componentSubgraph := graph.NewSubgraph()
		componentSubgraph.
			NodeBaseAttrs().
			Attr("width", "1.0").Attr("height", "1.0")
		componentSubgraph.
			Attr("label", componentName).
			Attr("cluster", "true").
			Attr("style", "rounded").
			Attr("color", "black").
			Attr("bgcolor", "lightgrey").
			Attr("margin", "20")

		// Component node
		componentNode := componentSubgraph.Node()
		componentNode.Attr("label", "𝑓")
		if c.Description() != "" {
			componentNode.Attr("label", c.Description())
		}
		componentNode.
			Attr("color", "blue").
			Attr("shape", "rect").
			Attr("group", componentName)

		// Input ports
		for portName := range c.Inputs() {
			portNode := componentSubgraph.NodeWithID(fmt.Sprintf("%s.inputs.%s", componentName, portName))
			portNode.
				Attr("label", portName).
				Attr("shape", "circle").
				Attr("group", componentName)

			componentSubgraph.Edge(portNode, componentNode)
		}

		// Output ports
		for portName, port := range c.Outputs() {
			portNode := componentSubgraph.NodeWithID(fmt.Sprintf("%s.inputs.%s", componentName, portName))
			portNode.
				Attr("label", portName).
				Attr("shape", "circle").
				Attr("group", componentName)

			componentSubgraph.Edge(componentNode, portNode)

			// Pipes
			//@TODO
		}
	}

	buf := new(bytes.Buffer)
	graph.Write(buf)

	return buf.Bytes()
}
