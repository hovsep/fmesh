package dot

import fmeshcomponent "github.com/hovsep/fmesh/component"

type attributesMap map[string]string

// ComponentConfig defines the configuration for the component visualization
type ComponentConfig struct {
	Subgraph                                 attributesMap
	SubgraphNodeBaseAttrs                    attributesMap
	Node                                     attributesMap
	NodeDefaultLabel                         string
	ErrorNode                                attributesMap
	SubgraphAttributesByActivationResultCode map[fmeshcomponent.ActivationResultCode]attributesMap
}

// PortConfig defines the configuration for the port visualization
type PortConfig struct {
	Node attributesMap
}

// LegendConfig defines the configuration for the legend visualization
type LegendConfig struct {
	Subgraph attributesMap
	Node     attributesMap
}

// PipeConfig defines the configuration for the pipe visualization
type PipeConfig struct {
	Edge attributesMap
}

// Config defines the configuration for the dot exporter
type Config struct {
	MainGraph attributesMap
	Component ComponentConfig
	Port      PortConfig
	Pipe      PipeConfig
	Legend    LegendConfig
}

var (
	defaultConfig = &Config{
		MainGraph: attributesMap{
			"layout":  "dot",
			"splines": "ortho",
		},
		Component: ComponentConfig{
			Subgraph: attributesMap{
				"cluster":  "true",
				"style":    "rounded",
				"color":    "black",
				"margin":   "20",
				"penwidth": "5",
			},
			SubgraphNodeBaseAttrs: attributesMap{
				"fontname": "Courier New",
				"width":    "1.0",
				"height":   "1.0",
				"penwidth": "2.5",
				"style":    "filled",
			},
			Node: attributesMap{
				"shape": "rect",
				"color": "#9dddea",
				"style": "filled",
			},
			NodeDefaultLabel: "𝑓",
			ErrorNode:        nil,
			SubgraphAttributesByActivationResultCode: map[fmeshcomponent.ActivationResultCode]attributesMap{
				fmeshcomponent.ActivationCodeOK: {
					"color": "green",
				},
				fmeshcomponent.ActivationCodeNoInput: {
					"color": "yellow",
				},
				fmeshcomponent.ActivationCodeNoFunction: {
					"color": "gray",
				},
				fmeshcomponent.ActivationCodeReturnedError: {
					"color": "red",
				},
				fmeshcomponent.ActivationCodePanicked: {
					"color": "pink",
				},
				fmeshcomponent.ActivationCodeWaitingForInputsClear: {
					"color": "blue",
				},
				fmeshcomponent.ActivationCodeWaitingForInputsKeep: {
					"color": "purple",
				},
			},
		},
		Port: PortConfig{
			Node: attributesMap{
				"shape": "circle",
			},
		},
		Pipe: PipeConfig{
			Edge: attributesMap{
				"minlen":   "3",
				"penwidth": "2",
				"color":    "#e437ea",
			},
		},
		Legend: LegendConfig{
			Subgraph: attributesMap{
				"style":     "dashed,filled",
				"fillcolor": "#e2c6fc",
			},
			Node: attributesMap{
				"shape":    "plaintext",
				"color":    "green",
				"fontname": "Courier New",
			},
		},
	}

	legendTemplate = `
	<table border="0" cellborder="0" cellspacing="10">
			{{ if .meshDescription }}
			<tr>
				<td>Description:</td><td>{{ .meshDescription }}</td>
			</tr>
			{{ end }}
		
			{{ if .cycleNumber }}
			<tr>
				<td>Cycle:</td><td>{{ .cycleNumber }}</td>
			</tr>
			{{ end }}

			{{ if .stats }}
				{{ range .stats }}
				<tr>
					<td>{{ .Name }}:</td><td>{{ .Value }}</td>
				</tr>
				{{ end }}
			{{ end }}
	</table>
	`
)
