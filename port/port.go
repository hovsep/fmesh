package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Port defines a connectivity point of a component
type Port struct {
	name   string
	signal *signal.Signal //Current signal set on the port
	pipes  Pipes          //Refs to all outbound pipes connected to this port
}

// Ports is just useful collection type
type Ports map[string]*Port

// NewPort creates a new port
func NewPort(name string) *Port {
	return &Port{name: name}
}

// NewPorts creates a new port with the given name
func NewPorts(names ...string) Ports {
	ports := make(Ports, len(names))
	for _, name := range names {
		ports[name] = NewPort(name)
	}
	return ports
}

// Name getter
func (p *Port) Name() string {
	return p.name
}

// Pipes getter
func (p *Port) Pipes() Pipes {
	return p.pipes
}

// Signal returns current signal set on the port
func (p *Port) Signal() *signal.Signal {
	return p.signal
}

// PutSignal adds a signal to current signal
func (p *Port) PutSignal(sig *signal.Signal) {
	p.signal = sig.Combine(p.Signal())
}

// ClearSignal removes current signal from the port
func (p *Port) ClearSignal() {
	p.signal = nil
}

// HasSignal says whether port signal is set or not
func (p *Port) HasSignal() bool {
	return p.signal != nil
}

// Adds pipe reference to the port, so all pipes of the port are easily accessible
func (p *Port) addPipeRef(pipe *Pipe) {
	if pipe.From == nil || pipe.To == nil {
		return
	}
	p.pipes = append(p.pipes, pipe)
}

// PipeTo creates one or multiple pipes to other port(s)
func (p *Port) PipeTo(toPorts ...*Port) {
	for _, toPort := range toPorts {
		newPipe := NewPipe(p, toPort)
		p.addPipeRef(newPipe)
		toPort.addPipeRef(newPipe)
	}
}

// ByName returns a port by its name
func (ports Ports) ByName(name string) *Port {
	return ports[name]
}

// ByNames returns multiple ports by their names
func (ports Ports) ByNames(names ...string) Ports {
	selectedPorts := make(Ports)

	for _, name := range names {
		if p, ok := ports[name]; ok {
			selectedPorts[name] = p
		}
	}

	return selectedPorts
}

// AnyHasSignal returns true if at least one port in collection has signal
func (ports Ports) AnyHasSignal() bool {
	for _, p := range ports {
		if p.HasSignal() {
			return true
		}
	}

	return false
}

// AllHaveSignal returns true when all ports in collection have signal
func (ports Ports) AllHaveSignal() bool {
	for _, p := range ports {
		if !p.HasSignal() {
			return false
		}
	}

	return true
}

// PutSignal puts a signal to all the port in collection
func (ports Ports) PutSignal(sig *signal.Signal) {
	for _, p := range ports {
		p.PutSignal(sig)
	}
}

// ClearSignal removes signals from all ports in collection
func (ports Ports) ClearSignal() {
	for _, p := range ports {
		p.ClearSignal()
	}
}

// ForwardSignal puts a signal from source port to dest port, without removing it on source port
func ForwardSignal(source *Port, dest *Port) {
	dest.PutSignal(source.Signal())
}
