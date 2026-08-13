package signal_test

import (
	"fmt"

	"github.com/hovsep/fmesh/signal"
)

// ExampleAs reads a payload back out of a signal with its type checked,
// reporting a mismatch instead of panicking.
func ExampleAs() {
	sig := signal.New(42)

	n, err := signal.As[int](sig)
	fmt.Println(n, err)

	// A wrong type is an error, not a panic.
	_, err = signal.As[string](sig)
	fmt.Println(err)
	// Output:
	// 42 <nil>
	// signal payload is int, not string
}

// ExampleAsOrDefault carries on with a default when the payload is missing or
// of another type. T is inferred from the default — an untyped 0 means int, so
// pass 0.0 (or use AsFloat64OrDefault) for float64 payloads.
func ExampleAsOrDefault() {
	sig := signal.New("not a number")

	fmt.Println(signal.AsOrDefault(sig, 7))

	temperature := signal.New(36.6)
	fmt.Println(signal.AsOrDefault(temperature, 0))   // untyped 0 infers int: default wins
	fmt.Println(signal.AsOrDefault(temperature, 0.0)) // 0.0 infers float64: payload wins
	// Output:
	// 7
	// 0
	// 36.6
}
