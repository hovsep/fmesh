package main

import (
	"fmt"
)

func RunAsPipes(input int) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	done := make(chan struct{})
	var res int

	go func() {
		ch1 <- Mul(input, 2)
	}()

	go func() {
		ch2 <- Add(<-ch1, 3)
	}()

	go func() {
		res = Add(<-ch2, 5)
		done <- struct{}{}

	}()

	<-done
	fmt.Printf("Result is %v", res)
}
