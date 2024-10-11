package export

import (
	"bytes"
	"fmt"
	"github.com/hovsep/fmesh"
	fmeshcomponent "github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/port"
	"github.com/lucasepe/dot"
)

type dotExporter struct {
}

const (
	nodeIDLabel    = "export/dot/id"
	portKindInput  = "input"
	portKindOutput = "output"
)

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

// ExportWithCycles returns multiple graphs showing the state of the given f-mesh in each activation cycle
func (d *dotExporter) ExportWithCycles(fm *fmesh.FMesh, cycles cycle.Collection) ([][]byte, error) {
	if len(fm.Components()) == 0 {
		return nil, nil
	}

	if len(cycles) == 0 {
		return nil, nil
	}

	results := make([][]byte, len(cycles))

	for cycleNumber, c := range cycles {
		graphForCycle, err := buildGraphForCycle(fm, c, cycleNumber)
		if err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		graphForCycle.Write(buf)

		results[cycleNumber] = buf.Bytes()
	}

	return results, nil
}

// buildGraph returns a graph representing the given f-mesh
func buildGraph(fm *fmesh.FMesh) (*dot.Graph, error) {
	mainGraph := getMainGraph(fm)

	addComponents(mainGraph, fm.Components(), nil)

	err := addPipes(mainGraph, fm.Components())
	if err != nil {
		return nil, err
	}
	return mainGraph, nil
}

func buildGraphForCycle(fm *fmesh.FMesh, activationCycle *cycle.Cycle, cycleNumber int) (*dot.Graph, error) {
	mainGraph := getMainGraph(fm)

	addCycleInfo(mainGraph, activationCycle, cycleNumber)

	addComponents(mainGraph, fm.Components(), activationCycle.ActivationResults())

	err := addPipes(mainGraph, fm.Components())
	if err != nil {
		return nil, err
	}
	return mainGraph, nil
}

// addPipes adds pipes representation to the graph
func addPipes(graph *dot.Graph, components fmeshcomponent.Collection) error {
	for _, c := range components {
		for _, srcPort := range c.Outputs() {
			for _, destPort := range srcPort.Pipes() {
				// Any destination port in any pipe is input port, but we do not know in which component
				// so we use the label we added earlier
				destPortID, err := destPort.Label(nodeIDLabel)
				if err != nil {
					return fmt.Errorf("failed to add pipe to port: %s : %w", destPort.Name(), err)
				}
				// Delete label, as it is not needed anymore
				destPort.DeleteLabel(nodeIDLabel)

				// Any source port in any pipe is always output port, so we can build its node ID
				srcPortNode := graph.FindNodeByID(getPortID(c.Name(), portKindOutput, srcPort.Name()))
				destPortNode := graph.FindNodeByID(destPortID)
				graph.Edge(srcPortNode, destPortNode).Attr("minlen", "3")
			}
		}
	}
	return nil
}

// addComponents adds components representation to the graph
func addComponents(graph *dot.Graph, components fmeshcomponent.Collection, activationResults fmeshcomponent.ActivationResultCollection) {
	for _, c := range components {
		// Component
		var activationResult *fmeshcomponent.ActivationResult
		if activationResults != nil {
			activationResult = activationResults.ByComponentName(c.Name())
		}
		componentSubgraph := getComponentSubgraph(graph, c, activationResult)
		componentNode := getComponentNode(componentSubgraph, c, activationResult)

		// Input ports
		for _, p := range c.Inputs() {
			portNode := getPortNode(c, p, portKindInput, componentSubgraph)
			componentSubgraph.Edge(portNode, componentNode)
		}

		// Output ports
		for _, p := range c.Outputs() {
			portNode := getPortNode(c, p, portKindOutput, componentSubgraph)
			componentSubgraph.Edge(componentNode, portNode)
		}
	}
}

// getPortNode creates and returns a node representing one port
func getPortNode(c *fmeshcomponent.Component, port *port.Port, portKind string, componentSubgraph *dot.Graph) *dot.Node {
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
func getComponentSubgraph(graph *dot.Graph, component *fmeshcomponent.Component, activationResult *fmeshcomponent.ActivationResult) *dot.Graph {
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

	// In cycle
	if activationResult != nil {
		switch activationResult.Code() {
		case fmeshcomponent.ActivationCodeOK:
			componentSubgraph.Attr("bgcolor", "green")
		case fmeshcomponent.ActivationCodeNoInput:
			componentSubgraph.Attr("bgcolor", "yellow")
		case fmeshcomponent.ActivationCodeNoFunction:
			componentSubgraph.Attr("bgcolor", "gray")
		case fmeshcomponent.ActivationCodeReturnedError:
			componentSubgraph.Attr("bgcolor", "red")
		case fmeshcomponent.ActivationCodePanicked:
			componentSubgraph.Attr("bgcolor", "pink")
		case fmeshcomponent.ActivationCodeWaitingForInputsClear:
			componentSubgraph.Attr("bgcolor", "blue")
		case fmeshcomponent.ActivationCodeWaitingForInputsKeep:
			componentSubgraph.Attr("bgcolor", "purple")
		default:
		}
	}

	return componentSubgraph
}

// getComponentNode creates component node and returns it
func getComponentNode(componentSubgraph *dot.Graph, component *fmeshcomponent.Component, activationResult *fmeshcomponent.ActivationResult) *dot.Node {
	componentNode := componentSubgraph.Node()
	label := "𝑓"

	if component.Description() != "" {
		label = component.Description()
	}

	if activationResult != nil {

		if activationResult.Error() != nil {
			errorNode := componentSubgraph.Node()
			errorNode.
				Attr("shape", "note").
				Attr("label", activationResult.Error().Error())
			componentSubgraph.Edge(componentNode, errorNode)
		}
	}

	componentNode.
		Attr("label", label).
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

// addDescription adds f-mesh description to graph
func addDescription(graph *dot.Graph, description string) {
	subgraph := graph.NewSubgraph()
	subgraph.
		Attr("label", "Description:").
		Attr("color", "green").
		Attr("fontcolor", "green").
		Attr("style", "dashed")
	node := subgraph.Node()
	node.
		Attr("shape", "plaintext").
		Attr("color", "green").
		Attr("fontcolor", "green").
		Attr("label", description)
}

// addCycleInfo adds useful insights about current cycle
func addCycleInfo(graph *dot.Graph, activationCycle *cycle.Cycle, cycleNumber int) {
	subgraph := graph.NewSubgraph()
	subgraph.
		Attr("label", "Cycle info:").
		Attr("style", "dashed")
	subgraph.NodeBaseAttrs().
		Attr("shape", "plaintext")

	// Current cycle number
	cycleNumberNode := subgraph.Node()
	cycleNumberNode.Attr("label", fmt.Sprintf("Current cycle: %d", cycleNumber))

	// Stats
	stats := getCycleStats(activationCycle)
	statNode := subgraph.Node()
	tableRows := dot.HTML("<table border=\"0\" cellborder=\"1\" cellspacing=\"0\">")
	for statName, statValue := range stats {
		//@TODO: keep order
		tableRows = tableRows + dot.HTML(fmt.Sprintf("<tr><td>%s : %d</td></tr>", statName, statValue))
	}
	tableRows = tableRows + "</table>"
	statNode.Attr("label", tableRows)
}

// getCycleStats returns basic cycle stats
func getCycleStats(activationCycle *cycle.Cycle) map[string]int {
	stats := make(map[string]int)
	for _, ar := range activationCycle.ActivationResults() {
		if ar.Activated() {
			stats["Activated"]++
		}

		stats[ar.Code().String()]++
	}
	return stats
}

// getPortID returns unique ID used to locate ports while building pipe edges
func getPortID(componentName string, portKind string, portName string) string {
	return fmt.Sprintf("component/%s/%s/%s", componentName, portKind, portName)
}
