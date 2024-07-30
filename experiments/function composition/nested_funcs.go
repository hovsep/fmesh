package main

import "fmt"

func RunAsNestedFuncs(input int) {
	res := Add(Add(Mul(input, 2), 3), 5)

	fmt.Printf("Result is %v", res)
}
