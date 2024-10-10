package export

import (
	"bytes"
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
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
	if len(fm.Components()) == 0 {
		return nil, nil
	}

	graph, err := buildGraph(fm)

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	graph.Write(buf)

	return buf.Bytes(), nil
}

// buildGraph returns a graph representing the given f-mesh
func buildGraph(fm *fmesh.FMesh) (*dot.Graph, error) {
	mainGraph := getMainGraph(fm)

	addComponents(mainGraph, fm.Components())

	err := addPipes(mainGraph, fm.Components())
	if err != nil {
		return nil, err
	}
	return mainGraph, nil
}

// addPipes adds pipes representation to the graph
func addPipes(graph *dot.Graph, components component.Collection) error {
	for _, c := range components {
		for _, srcPort := range c.Outputs() {
			for _, destPort := range srcPort.Pipes() {
				// Any destination port in any pipe is input port, but we do not know in which component
				// so we use the label we added earlier
				destPortID, err := destPort.Label(nodeIDLabel)
				if err != nil {
					return fmt.Errorf("failed to add pipe: %w", err)
				}
				// Clean up and leave the f-mesh as it was before export
				destPort.DeleteLabel(nodeIDLabel)

				// Any source port in any pipe is always output port, so we can build its node ID
				srcPortNode := graph.FindNodeByID(getPortID(c.Name(), "output", srcPort.Name()))
				destPortNode := graph.FindNodeByID(destPortID)
				graph.Edge(srcPortNode, destPortNode).Attr("minlen", 3)
			}
		}
	}
	return nil
}

// addComponents adds components representation to the graph
func addComponents(graph *dot.Graph, components component.Collection) {
	for _, c := range components {
		// Component
		componentSubgraph := getComponentSubgraph(graph, c)
		componentNode := getComponentNode(componentSubgraph, c)

		// Input ports
		for _, p := range c.Inputs() {
			portNode := getPortNode(c, p, "input", componentSubgraph)
			componentSubgraph.Edge(portNode, componentNode)
		}

		// Output ports
		for _, p := range c.Outputs() {
			portNode := getPortNode(c, p, "output", componentSubgraph)
			componentSubgraph.Edge(componentNode, portNode)
		}
	}
}

// getPortNode creates and returns a node representing one port
func getPortNode(c *component.Component, port *port.Port, portKind string, componentSubgraph *dot.Graph) *dot.Node {
	portID := getPortID(c.Name(), portKind, port.Name())

	//Mark ports to be able to find their respective nodes later when adding pipes
	port.AddLabel(nodeIDLabel, portID)

	portNode := componentSubgraph.NodeWithID(portID)
	portNode.
		Attr("label", port.Name()).
		Attr("shape", "circle").
		Attr("group", c.Name())
	return portNode
}

// getComponentSubgraph creates component subgraph and returns it
func getComponentSubgraph(graph *dot.Graph, component *component.Component) *dot.Graph {
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

	return componentSubgraph
}

// getComponentNodeCreate creates component node and returns it
func getComponentNode(componentSubgraph *dot.Graph, component *component.Component) *dot.Node {
	componentNode := componentSubgraph.Node()
	componentNode.Attr("label", "𝑓")
	if component.Description() != "" {
		componentNode.Attr("label", component.Description())
	}
	componentNode.
		Attr("color", "blue").
		Attr("shape", "rect").
		Attr("group", component.Name())
	return componentNode
}

// getMainGraph creates and returns the main (root) graph
func getMainGraph(fm *fmesh.FMesh) *dot.Graph {
	graph := dot.NewGraph(dot.Directed)
	graph.
		Attr("layout", "dot").
		Attr("splines", "ortho")

	if fm.Description() != "" {
		addDescription(graph, fm.Description())
	}

	return graph
}

func addDescription(graph *dot.Graph, description string) {
	descriptionSubgraph := graph.NewSubgraph()
	descriptionSubgraph.
		Attr("label", "Description:").
		Attr("color", "green").
		Attr("fontcolor", "green").
		Attr("style", "dashed")
	descriptionNode := descriptionSubgraph.Node()
	descriptionNode.
		Attr("shape", "plaintext").
		Attr("color", "green").
		Attr("fontcolor", "green").
		Attr("label", description)
}

// getPortID returns unique ID used to locate ports while building pipe edges
func getPortID(componentName string, portKind string, portName string) string {
	return fmt.Sprintf("component/%s/%s/%s", componentName, portKind, portName)
}
