package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/common"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"os"
)

const (
	portIn = "in"
)

// This demo demonstrates F-Mesh's signal filtering and routing capabilities.
// It showcases how components can filter and route signals based on conditions
func main() {
	filter := getFilter("pop-filter", common.LabelsCollection{"genre": "pop"})
	printer1 := getPrinter("dropped-printer")
	printer2 := getPrinter("passed-printer")

	filter.OutputByName("dropped").PipeTo(printer1.InputByName(portIn))
	filter.OutputByName("passed").PipeTo(printer2.InputByName(portIn))

	fm := fmesh.New("demo-filter").
		WithComponents(filter, printer1, printer2)

	// Init with data
	signalsToFilter := getSignals()
	filter.InputByName(portIn).PutSignals(signalsToFilter.SignalsOrNil()...)

	_, err := fm.Run()
	if err != nil {
		fmt.Println("Pipeline finished with error:", err)
		os.Exit(1)
	}

	fmt.Println("Filtering finished successfully")
}

func getPrinter(name string) *component.Component {
	return component.New(name).
		WithDescription("Simple stdout printer").
		WithInputs(portIn).
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName(portIn).AllSignalsOrNil() {
				fmt.Printf("%s: %v \n", this.Name(), sig.PayloadOrDefault("no payload"))
			}
			return nil
		})
}

func getFilter(name string, disallowedLabels common.LabelsCollection) *component.Component {
	return component.New(name).
		WithDescription("Simple filter").
		WithInputs(portIn).
		WithOutputs("dropped", "passed").
		WithActivationFunc(func(this *component.Component) error {
			for _, sig := range this.InputByName(portIn).AllSignalsOrNil() {
				if sig.HasAnyLabelWithValue(disallowedLabels) {
					this.OutputByName("dropped").PutSignals(sig)
				} else {
					this.OutputByName("passed").PutSignals(sig)
				}
			}
			return nil
		})
}

func getSignals() *signal.Group {
	return signal.NewGroup().With(
		signal.New("Justice").WithLabels(common.LabelsCollection{
			"genre":  "pop",
			"artist": "Justin Bieber",
			"year":   "2021",
		}),
		signal.New("Dysania").WithLabels(common.LabelsCollection{
			"genre":  "rock",
			"artist": "Elita",
			"year":   "2023",
		}),
		signal.New("After Hours").WithLabels(common.LabelsCollection{
			"genre":  "pop",
			"artist": "The Weeknd",
			"year":   "2020",
		}),
		signal.New("Random Access Memories").WithLabels(common.LabelsCollection{
			"genre":  "electronic",
			"artist": "Daft Punk",
			"year":   "2013",
		}),
		signal.New("Evermore").WithLabels(common.LabelsCollection{
			"genre":  "pop",
			"artist": "Taylor Swift",
			"year":   "2020",
		}),
		signal.New("1989").WithLabels(common.LabelsCollection{
			"genre":  "pop",
			"artist": "Taylor Swift",
			"year":   "2014",
		}),
		signal.New("To Pimp a Butterfly").WithLabels(common.LabelsCollection{
			"genre":  "hip-hop",
			"artist": "Kendrick Lamar",
			"year":   "2015",
		}),
		signal.New("Ghost Stories").WithLabels(common.LabelsCollection{
			"genre":  "alternative",
			"artist": "Coldplay",
			"year":   "2014",
		}),
		signal.New("Future Nostalgia").WithLabels(common.LabelsCollection{
			"genre":  "pop",
			"artist": "Dua Lipa",
			"year":   "2020",
		}))
}
