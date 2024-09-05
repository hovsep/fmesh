package main

type Port struct {
	signal Signal
	pipes  Pipes //Refs to pipes connected to that port (no in\out semantics)
}

func (p *Port) getSignal() Signal {
	if p == nil {
		panic("invalid port")
	}

	if !p.hasSignal() {
		return nil
	}

	if p.signal.IsAggregate() {
		return p.signal.(*AggregateSignal)
	}
	return p.signal.(*SingleSignal)
}

func (p *Port) putSignal(sig Signal) {
	if p.hasSignal() {
		//Aggregate signal
		var resValues []*SingleSignal

		//Extract existing signal(s)
		if p.signal.IsSingle() {
			resValues = append(resValues, p.signal.(*SingleSignal))
		} else if p.signal.IsAggregate() {
			resValues = p.signal.(*AggregateSignal).val
		}

		//Add new signal(s)
		if sig.IsSingle() {
			resValues = append(resValues, sig.(*SingleSignal))
		} else if sig.IsAggregate() {
			resValues = append(resValues, sig.(*AggregateSignal).val...)
		}

		p.signal = &AggregateSignal{
			val: resValues,
		}
		return
	}

	//Single signal
	p.signal = sig
}

func (p *Port) clearSignal() {
	p.signal = nil
}

func (p *Port) hasSignal() bool {
	return p.signal != nil
}

// Adds pipe reference to port, so all pipes of the port are easily iterable (no in\out semantics)
func (p *Port) addPipeRef(pipe *Pipe) {
	p.pipes = append(p.pipes, pipe)
}

// CreatePipeTo must be used to explicitly set pipe direction
func (p *Port) CreatePipesTo(toPorts ...*Port) {
	for _, toPort := range toPorts {
		newPipe := &Pipe{
			From: p,
			To:   toPort,
		}
		p.addPipeRef(newPipe)
		toPort.addPipeRef(newPipe)
	}

}
