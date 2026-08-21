package component_test

import (
	"context"
	"fmt"

	"github.com/hovsep/fmesh/component"
)

// ExampleRequireInputs composes an activation function that waits (keeping
// partial inputs) until both operands have arrived.
func ExampleRequireInputs() {
	adder, err := component.New("adder",
		component.WithInputs("a", "b"),
		component.WithOutputs("sum"),
		component.WithActivationFunc(component.Sequential(
			component.RequireInputs("a", "b"),
			func(_ context.Context, this *component.Component) error {
				a, err := this.InputByName("a").Signals().FirstAs[int]()
				if err != nil {
					return err
				}
				b, err := this.InputByName("b").Signals().FirstAs[int]()
				if err != nil {
					return err
				}
				return this.OutputByName("sum").PutPayloads(a + b)
			},
		)))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Only one operand so far: the component suspends and keeps it.
	_ = adder.InputByName("a").PutPayloads(2)
	fmt.Println(adder.MaybeActivate(ctx).Code())

	// The second operand arrives: the activation runs.
	_ = adder.InputByName("b").PutPayloads(3)
	fmt.Println(adder.MaybeActivate(ctx).Code())

	sum, _ := adder.OutputByName("sum").Signals().FirstPayload()
	fmt.Println(sum)
	// Output:
	// Waiting for input (keep)
	// Success
	// 5
}

// ExampleComponent_State keeps a counter that survives across activations.
func ExampleComponent_State() {
	counter, err := component.New("counter",
		component.WithInputs("in"),
		component.WithOutputs("count"),
		component.WithActivationFunc(func(_ context.Context, this *component.Component) error {
			seen := this.State().UpdateAndGet("seen", func(old any) any {
				if old == nil {
					return 1
				}
				return old.(int) + 1
			})
			return this.OutputByName("count").PutPayloads(seen)
		}))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for range 3 {
		_ = counter.InputByName("in").PutPayloads("tick")
		_ = counter.MaybeActivate(ctx)
		_ = counter.ClearInputs(ctx)
	}

	fmt.Println(counter.OutputByName("count").Signals().AllPayloads())
	// Output:
	// [1 2 3]
}
