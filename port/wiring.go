package port

import "fmt"

// Pipe is a single wiring edge, from an output port to an input port.
type Pipe struct {
	From *Port
	To   *Port
}

// MultiPipe wires several 1:1 connections and says which one failed. A nil port
// is reported rather than panicked on: the usual cause is a port name that does
// not exist, and the name is what the caller needs told.
func MultiPipe(pipes ...Pipe) error {
	for _, p := range pipes {
		if p.From == nil || p.To == nil {
			return fmt.Errorf("cannot pipe: %s", p)
		}
		if err := p.From.PipeTo(p.To); err != nil {
			return fmt.Errorf("piping %s: %w", p, err)
		}
	}
	return nil
}

// String describes the edge, naming whichever end is missing.
func (p Pipe) String() string {
	return fmt.Sprintf("%s -> %s", portName(p.From), portName(p.To))
}

// Pair is a source and destination port, for forwarding rather than piping.
type Pair struct {
	From *Port
	To   *Port
}

// MultiForward forwards signals across several 1:1 pairs, reporting nil ports
// and naming the pair that failed, as MultiPipe does.
func MultiForward(pairs ...Pair) error {
	for _, pair := range pairs {
		if pair.From == nil || pair.To == nil {
			return fmt.Errorf("cannot forward: %s", pair)
		}
		if err := ForwardSignals(pair.From, pair.To); err != nil {
			return fmt.Errorf("forwarding %s: %w", pair, err)
		}
	}
	return nil
}

// String describes the pair, naming whichever end is missing.
func (p Pair) String() string {
	return fmt.Sprintf("%s -> %s", portName(p.From), portName(p.To))
}

func portName(p *Port) string {
	if p == nil {
		return "<missing port>"
	}
	return p.Name()
}
