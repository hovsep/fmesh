package export

import (
	"bytes"
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/lucasepe/dot"
)

type dotExporter struct {
}

const nodeIDLabel = "export/dot/id"

func NewDotExporter() Exporter {
	return &dotExporter{}
}

// Export returns the f-mesh represented as digraph in DOT language
func (d *dotExporter) Export(fm *fmesh.FMesh) ([]byte, error) {
	// Setup main graph
	graph := dot.NewGraph(dot.Directed)
	graph.
		Attr("layout", "dot").
		Attr("splines", "ortho")

	for _, component := range fm.Components() {
		// Component subgraph (wrapper)
		componentSubgraph := graph.NewSubgraph()
		componentSubgraph.
			NodeBaseAttrs().
			Attr("width", "1.0").Attr("height", "1.0")
		componentSubgraph.
			Attr("label", component.Name()).
			Attr("cluster", "true").
			Attr("style", "rounded").
			Attr("color", "black").
			Attr("bgcolor", "lightgrey").
			Attr("margin", "20")

		// Create component node and subgraph (cluster)
		componentNode := componentSubgraph.Node()
		componentNode.Attr("label", "𝑓")
		if component.Description() != "" {
			componentNode.Attr("label", component.Description())
		}
		componentNode.
			Attr("color", "blue").
			Attr("shape", "rect").
			Attr("group", component.Name())

		// Create nodes for input ports
		for _, port := range component.Inputs() {
			portID := getPortID(component.Name(), "input", port.Name())

			//Mark input ports to be able to find their respective  nodes later when adding pipes
			port.AddLabel(nodeIDLabel, portID)

			portNode := componentSubgraph.NodeWithID(portID)
			portNode.
				Attr("label", port.Name()).
				Attr("shape", "circle").
				Attr("group", component.Name())

			componentSubgraph.Edge(portNode, componentNode)
		}

		// Create nodes for output ports
		for _, port := range component.Outputs() {
			portID := getPortID(component.Name(), "output", port.Name())
			portNode := componentSubgraph.NodeWithID(portID)
			portNode.
				Attr("label", port.Name()).
				Attr("shape", "circle").
				Attr("group", component.Name())

			componentSubgraph.Edge(componentNode, portNode)
		}
	}

	// Create edges representing pipes (all ports must exist at this point)
	for _, component := range fm.Components() {
		for _, srcPort := range component.Outputs() {
			for _, destPort := range srcPort.Pipes() {
				// Any destination port in any pipe is input port, but we do not know in which component
				// so we use the label we added earlier
				destPortID, err := destPort.Label(nodeIDLabel)
				if err != nil {
					return nil, err
				}
				// Clean up and leave the f-mesh as it was before export
				destPort.DeleteLabel(nodeIDLabel)

				// Any source port in any pipe is always output port, so we can build its node ID
				srcPortNode := graph.FindNodeByID(getPortID(component.Name(), "output", srcPort.Name()))
				destPortNode := graph.FindNodeByID(destPortID)
				graph.Edge(srcPortNode, destPortNode)
			}
		}
	}

	buf := new(bytes.Buffer)
	graph.Write(buf)

	return buf.Bytes(), nil
}

func getPortID(componentName string, portKind string, portName string) string {
	return fmt.Sprintf("component/%s/%s/%s", componentName, portKind, portName)
}
