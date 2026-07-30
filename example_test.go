package fmesh_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

// Example mirrors the README quick-start: two components connected by a pipe,
// run to completion in discrete cycles.
func Example() {
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	concat, err := component.New("concat",
		component.WithInputs("i1", "i2"),
		component.WithOutputs("res"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			word1 := signal.AsOrDefault(this.InputByName("i1").Signals().First(), "")
			word2 := signal.AsOrDefault(this.InputByName("i2").Signals().First(), "")
			return this.OutputByName("res").PutSignals(signal.New(word1 + word2))
		}))
	must(err)

	uppercase, err := component.New("uppercase",
		component.WithInputs("i1"),
		component.WithOutputs("res"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			str := signal.AsOrDefault(this.InputByName("i1").Signals().First(), "")
			return this.OutputByName("res").PutSignals(signal.New(strings.ToUpper(str)))
		}))
	must(err)

	fm, err := fmesh.New("hello world")
	must(err)
	must(fm.AddComponents(concat, uppercase))

	must(concat.OutputByName("res").PipeTo(uppercase.InputByName("i1")))

	must(concat.InputByName("i1").PutSignals(signal.New("hello ")))
	must(concat.InputByName("i2").PutSignals(signal.New("world!")))

	_, err = fm.Run(context.Background())
	must(err)

	result, err := uppercase.OutputByName("res").Signals().FirstPayload()
	must(err)
	fmt.Printf("Result: %v\n", result)
	// Output: Result: HELLO WORLD!
}
