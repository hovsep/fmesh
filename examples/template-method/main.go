package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Template Method Pattern Example
// This example demonstrates how to implement the template method pattern using FMesh components.
// It shows how to:
// 1. Define a template algorithm in a base component
// 2. Allow specific steps to be overridden by variants
// 3. Maintain consistent process flow
// 4. Reuse common functionality
// The pattern is useful for:
// - Implementing standardized workflows
// - Enforcing process consistency
// - Allowing controlled variations
// - Avoiding code duplication
// Common use cases include:
// - Document processing
// - Data transformation pipelines
// - Build systems
// - Report generation
// - Service initialization
func main() {
	// Create document generator
	generator := component.New("generator").
		WithDescription("Generates documents to be processed").
		WithInputs("start").
		WithOutputs("document").
		WithActivationFunc(func(this *component.Component) error {
			documents := []string{
				"PDF:Report Q1 2024:confidential",
				"TXT:Meeting Notes:internal",
				"HTML:Blog Post:public",
			}

			for _, doc := range documents {
				this.OutputByName("document").PutSignals(signal.New(doc))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})

	// Create PDF processor component
	pdfProcessor := component.New("pdf-processor").
		WithDescription("Processes PDF documents").
		WithInputs("document").
		WithOutputs("processed").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("document").AllSignalsOrNil() {
				doc := s.PayloadOrNil().(string)
				parts := strings.Split(doc, ":")
				if parts[0] != "PDF" {
					continue
				}

				// Template method steps
				fmt.Printf("PDF: Loading document: %s\n", parts[1])
				fmt.Printf("PDF: Extracting text and metadata\n")

				if parts[2] == "confidential" {
					fmt.Printf("PDF: Applying watermark: CONFIDENTIAL\n")
				}

				fmt.Printf("PDF: Converting to searchable format\n")
				fmt.Printf("PDF: Saving processed document\n")

				this.OutputByName("processed").PutSignals(signal.New(fmt.Sprintf("PDF processed: %s", parts[1])))
			}
			return nil
		})

	// Create TXT processor component
	txtProcessor := component.New("txt-processor").
		WithDescription("Processes TXT documents").
		WithInputs("document").
		WithOutputs("processed").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("document").AllSignalsOrNil() {
				doc := s.PayloadOrNil().(string)
				parts := strings.Split(doc, ":")
				if parts[0] != "TXT" {
					continue
				}

				// Template method steps
				fmt.Printf("TXT: Loading document: %s\n", parts[1])
				fmt.Printf("TXT: Spell checking\n")

				if parts[2] == "internal" {
					fmt.Printf("TXT: Adding internal use disclaimer\n")
				}

				fmt.Printf("TXT: Formatting text\n")
				fmt.Printf("TXT: Saving processed document\n")

				this.OutputByName("processed").PutSignals(signal.New(fmt.Sprintf("TXT processed: %s", parts[1])))
			}
			return nil
		})

	// Create HTML processor component
	htmlProcessor := component.New("html-processor").
		WithDescription("Processes HTML documents").
		WithInputs("document").
		WithOutputs("processed").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("document").AllSignalsOrNil() {
				doc := s.PayloadOrNil().(string)
				parts := strings.Split(doc, ":")
				if parts[0] != "HTML" {
					continue
				}

				// Template method steps
				fmt.Printf("HTML: Loading document: %s\n", parts[1])
				fmt.Printf("HTML: Validating markup\n")

				if parts[2] == "public" {
					fmt.Printf("HTML: Adding SEO metadata\n")
				}

				fmt.Printf("HTML: Minifying content\n")
				fmt.Printf("HTML: Saving processed document\n")

				this.OutputByName("processed").PutSignals(signal.New(fmt.Sprintf("HTML processed: %s", parts[1])))
			}
			return nil
		})

	// Create result collector
	collector := component.New("collector").
		WithDescription("Collects processing results").
		WithInputs("result").
		WithActivationFunc(func(this *component.Component) error {
			for _, s := range this.InputByName("result").AllSignalsOrNil() {
				fmt.Printf("Collector: %v\n", s.PayloadOrNil())
			}
			return nil
		})

	// Create and run mesh
	mesh := fmesh.New("template-method-example").
		WithDescription("Demonstrates template method pattern").
		WithComponents(generator, pdfProcessor, txtProcessor, htmlProcessor, collector)

	// Connect components
	generator.OutputByName("document").PipeTo(pdfProcessor.InputByName("document"))
	generator.OutputByName("document").PipeTo(txtProcessor.InputByName("document"))
	generator.OutputByName("document").PipeTo(htmlProcessor.InputByName("document"))

	pdfProcessor.OutputByName("processed").PipeTo(collector.InputByName("result"))
	txtProcessor.OutputByName("processed").PipeTo(collector.InputByName("result"))
	htmlProcessor.OutputByName("processed").PipeTo(collector.InputByName("result"))

	// Start the mesh by sending a signal to the generator
	generator.InputByName("start").PutSignals(signal.New("start"))

	info, err := mesh.Run()
	if err != nil {
		fmt.Printf("Error running mesh: %v\n", err)
		return
	}

	fmt.Printf("Mesh completed in %v\n", info.Duration)
}
