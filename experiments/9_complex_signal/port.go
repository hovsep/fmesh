package main

type Port struct {
	val   Signal
	pipes Pipes //Refs to pipes connected to that port (no in\out semantics)
}

func (p *Port) getValue() Signal {
	return p.val
}

func (p *Port) setValue(val Signal) {
	if p.hasValue() {
		//Aggregate signal
		var resValues []*SingleSignal

		//Extract existing signal(s)
		if p.val.IsSingle() {
			resValues = append(resValues, p.val.(*SingleSignal))
		} else if p.val.IsAggregate() {
			resValues = p.val.(*AggregateSignal).val
		}

		//Add new signal(s)
		if val.IsSingle() {
			resValues = append(resValues, val.(*SingleSignal))
		} else if val.IsAggregate() {
			resValues = append(resValues, val.(*AggregateSignal).val...)
		}

		p.val = &AggregateSignal{
			val: resValues,
		}
		return
	}

	//Single signal
	p.val = val
}

func (p *Port) clearValue() {
	p.val = nil
}

func (p *Port) hasValue() bool {
	return p.val != nil
}

// Adds pipe reference to port, so all pipes of the port are easily iterable (no in\out semantics)
func (p *Port) addPipeRef(pipe *Pipe) {
	p.pipes = append(p.pipes, pipe)
}

// CreatePipeTo must be used to explicitly set pipe direction
func (p *Port) CreatePipeTo(toPort *Port) {
	newPipe := &Pipe{
		From: p,
		To:   toPort,
	}
	p.addPipeRef(newPipe)
	toPort.addPipeRef(newPipe)
}
