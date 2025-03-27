package dot

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"

	"github.com/hovsep/fmesh"
	fmeshcomponent "github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/cycle"
	"github.com/hovsep/fmesh/export"
	"github.com/hovsep/fmesh/port"
	"github.com/lucasepe/dot"
)

type statEntry struct {
	Name  string
	Value int
}

type dotExporter struct {
	config *Config
}

const (
	nodeIDLabel = "export/dot/id"
)

// NewDotExporter returns exporter with default configuration
func NewDotExporter() export.Exporter {
	return NewDotExporterWithConfig(defaultConfig)
}

// NewDotExporterWithConfig returns exporter with custom configuration
func NewDotExporterWithConfig(config *Config) export.Exporter {
	return &dotExporter{
		config: config,
	}
}

// Export returns the f-mesh as DOT-graph
func (d *dotExporter) Export(fm *fmesh.FMesh) ([]byte, error) {
	if fm.Components().Len() == 0 {
		return nil, nil
	}

	graph, err := d.buildGraph(fm, nil)

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	graph.Write(buf)

	return buf.Bytes(), nil
}

// ExportWithCycles returns multiple graphs showing the state of the given f-mesh in each activation cycle
func (d *dotExporter) ExportWithCycles(fm *fmesh.FMesh, activationCycles cycle.Cycles) ([][]byte, error) {
	if fm.Components().Len() == 0 {
		return nil, nil
	}

	if len(activationCycles) == 0 {
		return nil, nil
	}

	results := make([][]byte, len(activationCycles))

	for _, activationCycle := range activationCycles {
		graphForCycle, err := d.buildGraph(fm, activationCycle)
		if err != nil {
			return nil, err
		}

		buf := new(bytes.Buffer)
		graphForCycle.Write(buf)

		results[activationCycle.Number()-1] = buf.Bytes()
	}

	return results, nil
}

// buildGraph returns f-mesh as a graph
// activationCycle may be passed optionally to get a representation of f-mesh in a given activation cycle
func (d *dotExporter) buildGraph(fm *fmesh.FMesh, activationCycle *cycle.Cycle) (*dot.Graph, error) {
	mainGraph, err := d.getMainGraph(fm, activationCycle)
	if err != nil {
		return nil, fmt.Errorf("failed to get main graph: %w", err)
	}

	components, err := fm.Components().Components()
	if err != nil {
		return nil, fmt.Errorf("failed to get components: %w", err)
	}

	err = d.addComponents(mainGraph, components, activationCycle)
	if err != nil {
		return nil, fmt.Errorf("failed to add components: %w", err)
	}

	err = d.addPipes(mainGraph, components)
	if err != nil {
		return nil, fmt.Errorf("failed to add pipes: %w", err)
	}
	return mainGraph, nil
}

// getMainGraph creates and returns the main (root) graph
func (d *dotExporter) getMainGraph(fm *fmesh.FMesh, activationCycle *cycle.Cycle) (*dot.Graph, error) {
	graph := dot.NewGraph(dot.Directed)

	setAttrMap(&graph.AttributesMap, d.config.MainGraph)

	err := d.addLegend(graph, fm, activationCycle)
	if err != nil {
		return nil, fmt.Errorf("failed to build main graph: %w", err)
	}

	return graph, nil
}

// addPipes adds pipes representation to the graph
func (d *dotExporter) addPipes(graph *dot.Graph, components fmeshcomponent.Map) error {
	for _, c := range components {
		srcPorts, err := c.Outputs().Ports()
		if err != nil {
			return err
		}

		for _, srcPort := range srcPorts {
			destPorts, err := srcPort.Pipes().Ports()
			if err != nil {
				return err
			}
			for _, destPort := range destPorts {
				// Any destination port in any pipe is input port, but we do not know in which component
				// so we use the label we added earlier
				destPortID, err := destPort.Label(nodeIDLabel)
				if err != nil {
					return fmt.Errorf("failed to add pipe to port: %s : %w", destPort.Name(), err)
				}
				// Delete label, as it is not needed anymore
				// destPort.DeleteLabel(nodeIDLabel)

				// Any source port in any pipe is always output port, so we can build its node ID
				srcPortNode := graph.FindNodeByID(getPortID(c.Name(), port.DirectionOut, srcPort.Name()))
				destPortNode := graph.FindNodeByID(destPortID)

				graph.Edge(srcPortNode, destPortNode, func(a *dot.AttributesMap) {
					setAttrMap(a, d.config.Pipe.Edge)
				})
			}
		}
	}
	return nil
}

// addComponents adds components representation to the graph
func (d *dotExporter) addComponents(graph *dot.Graph, components fmeshcomponent.Map, activationCycle *cycle.Cycle) error {
	for _, c := range components {
		// Component
		var activationResult *fmeshcomponent.ActivationResult
		if activationCycle != nil {
			activationResult = activationCycle.ActivationResults().ByComponentName(c.Name())
		}
		componentSubgraph := d.getComponentSubgraph(graph, c, activationResult)
		componentNode := d.getComponentNode(componentSubgraph, c, activationResult)

		// Input ports
		inputPorts, err := c.Inputs().Ports()
		if err != nil {
			return err
		}
		for _, p := range inputPorts {
			portNode := d.getPortNode(c, p, componentSubgraph)
			componentSubgraph.Edge(portNode, componentNode)
		}

		// Output ports
		outputPorts, err := c.Outputs().Ports()
		if err != nil {
			return err
		}
		for _, p := range outputPorts {
			portNode := d.getPortNode(c, p, componentSubgraph)
			componentSubgraph.Edge(componentNode, portNode)
		}
	}
	return nil
}

// getPortNode creates and returns a node representing one port
func (d *dotExporter) getPortNode(c *fmeshcomponent.Component, p *port.Port, componentSubgraph *dot.Graph) *dot.Node {
	portID := getPortID(c.Name(), p.LabelOrDefault(port.DirectionLabel, ""), p.Name())

	// Mark ports to be able to find their respective nodes later when adding pipes
	p.AddLabel(nodeIDLabel, portID)

	portNode := componentSubgraph.NodeWithID(portID, func(a *dot.AttributesMap) {
		setAttrMap(a, d.config.Port.Node)
		a.Attr("label", p.Name()).Attr("group", c.Name())
	})

	return portNode
}

// getComponentSubgraph creates component subgraph and returns it
func (d *dotExporter) getComponentSubgraph(graph *dot.Graph, component *fmeshcomponent.Component, activationResult *fmeshcomponent.ActivationResult) *dot.Graph {
	componentSubgraph := graph.NewSubgraph()

	setAttrMap(componentSubgraph.NodeBaseAttrs(), d.config.Component.SubgraphNodeBaseAttrs)
	setAttrMap(&componentSubgraph.AttributesMap, d.config.Component.Subgraph)

	// Set cycle specific attributes
	if activationResult != nil {
		if attributesByCode, ok := d.config.Component.SubgraphAttributesByActivationResultCode[activationResult.Code()]; ok {
			setAttrMap(&componentSubgraph.AttributesMap, attributesByCode)
		}
	}

	componentSubgraph.Attr("label", component.Name())

	return componentSubgraph
}

// getComponentNode creates component node and returns it
func (d *dotExporter) getComponentNode(componentSubgraph *dot.Graph, component *fmeshcomponent.Component, activationResult *fmeshcomponent.ActivationResult) *dot.Node {
	componentNode := componentSubgraph.Node(func(a *dot.AttributesMap) {
		setAttrMap(a, d.config.Component.Node)
	})

	label := d.config.Component.NodeDefaultLabel

	if component.Description() != "" {
		label = component.Description()
	}

	if activationResult != nil {
		if activationResult.ActivationError() != nil {
			errorNode := componentSubgraph.Node(func(a *dot.AttributesMap) {
				setAttrMap(a, d.config.Component.ErrorNode)
			})
			errorNode.
				Attr("label", activationResult.ActivationError().Error())
			componentSubgraph.Edge(componentNode, errorNode)
		}
	}

	componentNode.
		Attr("label", label).
		Attr("group", component.Name())
	return componentNode
}

// addLegend adds useful information about f-mesh and (optionally) current activation cycle
func (d *dotExporter) addLegend(graph *dot.Graph, fm *fmesh.FMesh, activationCycle *cycle.Cycle) error {
	subgraph := graph.NewSubgraph()

	setAttrMap(&subgraph.AttributesMap, d.config.Legend.Subgraph)
	subgraph.Attr("label", "Legend:")

	legendData := make(map[string]any)
	legendData["meshDescription"] = fmt.Sprintf("This mesh consist of %d components", fm.Components().Len())
	if fm.Description() != "" {
		legendData["meshDescription"] = fm.Description()
	}

	if activationCycle != nil {
		legendData["cycleNumber"] = activationCycle.Number()
		legendData["stats"] = getCycleStats(activationCycle)
	}

	legendHTML := new(bytes.Buffer)
	err := template.Must(
		template.New("legend").
			Parse(legendTemplate)).
		Execute(legendHTML, legendData)

	if err != nil {
		return fmt.Errorf("failed to render legend: %w", err)
	}

	subgraph.Node(func(a *dot.AttributesMap) {
		setAttrMap(a, d.config.Legend.Node)
		a.Attr("label", dot.HTML(legendHTML.String()))
	})

	return nil
}

// getCycleStats returns basic cycle stats
func getCycleStats(activationCycle *cycle.Cycle) []*statEntry {
	statsMap := map[string]*statEntry{
		// Number of activated must be shown always
		"activated": {
			Name:  "Activated",
			Value: 0,
		},
	}
	for _, ar := range activationCycle.ActivationResults().All() {
		if ar.Activated() {
			statsMap["activated"].Value++
		}

		if entryByCode, ok := statsMap[ar.Code().String()]; ok {
			entryByCode.Value++
		} else {
			statsMap[ar.Code().String()] = &statEntry{
				Name:  ar.Code().String(),
				Value: 1,
			}
		}
	}
	// Convert to slice to preserve keys order
	statsList := make([]*statEntry, 0)
	for _, entry := range statsMap {
		statsList = append(statsList, entry)
	}

	sort.Slice(statsList, func(i, j int) bool {
		return statsList[i].Name < statsList[j].Name
	})
	return statsList
}

// getPortID returns unique ID used to locate ports while building pipe edges
func getPortID(componentName, portDirection, portName string) string {
	return fmt.Sprintf("component/%s/%s/%s", componentName, portDirection, portName)
}

// setAttrMap sets all attributes to target
func setAttrMap(target *dot.AttributesMap, attributes attributesMap) {
	for attrName, attrValue := range attributes {
		target.Attr(attrName, attrValue)
	}
}
