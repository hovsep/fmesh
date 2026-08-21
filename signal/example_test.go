package signal_test

import (
	"fmt"

	"github.com/hovsep/fmesh/signal"
)

// ExampleSignal_As reads a payload back out of a signal with its type checked,
// reporting a mismatch instead of panicking.
func ExampleSignal_As() {
	sig := signal.New(42)

	n, err := sig.As[int]()
	fmt.Println(n, err)

	// A wrong type is an error, not a panic.
	_, err = sig.As[string]()
	fmt.Println(err)
	// Output:
	// 42 <nil>
	// signal payload is int, not string
}

// ExampleSignal_PayloadOrDefault carries on with a default when the payload is
// missing or of another type. T is inferred from the default — an untyped 0
// means int, so pass 0.0 (or use Float64OrDefault) for float64 payloads.
func ExampleSignal_PayloadOrDefault() {
	sig := signal.New("not a number")

	fmt.Println(sig.PayloadOrDefault(7))

	temperature := signal.New(36.6)
	fmt.Println(temperature.PayloadOrDefault(0))   // untyped 0 infers int: default wins
	fmt.Println(temperature.PayloadOrDefault(0.0)) // 0.0 infers float64: payload wins
	// Output:
	// 7
	// 0
	// 36.6
}
