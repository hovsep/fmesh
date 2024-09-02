package main

type Pipe struct {
	From *Port
	To   *Port
}

type Pipes []*Pipe
