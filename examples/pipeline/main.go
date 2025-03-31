package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
	"os"
	"strconv"
	"time"
)

const (
	PortIn  = "in"
	PortOut = "out"
)

// This example demonstrates simple pipeline implementation
func main() {
	stdInReader := getStdInReader("Please input some text")
	fileReader := getFileReader()
	fileWriter := getFileWriter()

	fm := buildPipeline(stdInReader, fileWriter, fileReader)

	//Init
	fm.ComponentByName("stdin-reader").InputByName(PortIn).PutSignals(signal.New("start"))

	_, err := fm.Run()
	if err != nil {
		fmt.Println("Pipeline finished with error:", err)
	}
}

func getFileReader() *component.Component {
	return component.New("file-reader").
		WithDescription("read file in current directory").
		WithActivationFunc(func(this *component.Component) error {
			// We expect exactly one signal with file name
			filename := this.InputByName(PortIn).FirstSignalPayloadOrDefault("").(string)
			if filename == "" {
				return errors.New("no input filename")
			}

			root, err := os.OpenRoot(".")
			if err != nil {
				return err
			}

			file, err := root.Open(filename)
			buf := make([]byte, 0)

			_, err = file.Read(buf)
			if err != nil {
				return err
			}
			this.OutputByName(PortOut).PutSignals(signal.New(string(buf)))
			return nil
		})
}

func getFileWriter() *component.Component {
	return component.New("file-writer").
		WithDescription("write to a file in current directory").
		WithActivationFunc(func(this *component.Component) error {
			// We expect exactly one signal with file contents
			data := this.InputByName(PortIn).FirstSignalPayloadOrDefault("").(string)
			if data == "" {
				return errors.New("no file contents")
			}

			root, err := os.OpenRoot(".")
			if err != nil {
				return err
			}
			fileName := time.Now().String()
			file, err := root.Create(fileName)
			defer root.Close()

			_, err = file.WriteString(data)
			if err != nil {
				return err
			}
			this.OutputByName(PortOut).PutSignals(signal.New(fileName))
			return nil
		})
}

func getStdInReader(prompt string) *component.Component {
	return component.New("stdin-reader").
		WithDescription("read a line from stdin").
		WithActivationFunc(func(this *component.Component) error {
			scanner := bufio.NewScanner(os.Stdin)

			fmt.Println(prompt)
			scanner.Scan()
			input := scanner.Text()

			if input == "" {
				return errors.New("no input typed")
			}

			this.OutputByName(PortOut).PutSignals(signal.New(scanner.Text()))

			return nil
		})
}

func buildPipeline(components ...*component.Component) *fmesh.FMesh {
	stageIndex := 0
	fm := fmesh.New("pipeline")

	for _, c := range components {
		// We can add custom labels
		c.AddLabel("stage", strconv.Itoa(stageIndex))

		fm = fm.WithComponents(withPipelineInterface(c))

		// Connect stages with pipes
		if stageIndex > 0 {
			// We can find previous component by name, but also we can use our custom labels
			previousStageComponent := fm.Components().ByLabelValue("stage", strconv.Itoa(stageIndex-1)).First()
			previousStageComponent.OutputByName(PortOut).PipeTo(c.InputByName(PortIn))
		}
		stageIndex++
	}

	return fm
}

// This helper function allows us to define the common interface shared by all components
// as we are building a pipeline each component will have one input and one output
// not required, just nice and easy way to avoid code duplication
func withPipelineInterface(c *component.Component) *component.Component {
	return c.WithInputs(PortIn).WithOutputs(PortOut)
}
