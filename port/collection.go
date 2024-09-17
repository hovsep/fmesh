package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Collection is a port collection with useful methods
type Collection map[string]*Port

// NewCollection creates empty collection
func NewCollection() Collection {
	return make(Collection)
}

// ByName returns a port by its name
func (collection Collection) ByName(name string) *Port {
	return collection[name]
}

// ByNames returns multiple ports by their names
func (collection Collection) ByNames(names ...string) Collection {
	selectedPorts := make(Collection)

	for _, name := range names {
		if p, ok := collection[name]; ok {
			selectedPorts[name] = p
		}
	}

	return selectedPorts
}

// AnyHasSignals returns true if at least one port in collection has signals
func (collection Collection) AnyHasSignals() bool {
	for _, p := range collection {
		if p.HasSignals() {
			return true
		}
	}

	return false
}

// AllHaveSignals returns true when all ports in collection have signals
func (collection Collection) AllHaveSignals() bool {
	for _, p := range collection {
		if !p.HasSignals() {
			return false
		}
	}

	return true
}

// PutSignals puts a signals to all the port in collection
func (collection Collection) PutSignals(signals ...*signal.Signal) {
	for _, p := range collection {
		p.PutSignals(signals...)
	}
}

// ClearSignals removes signals from all ports in collection
func (collection Collection) ClearSignals() {
	for _, p := range collection {
		p.ClearSignals()
	}
}

func (collection Collection) Add(ports ...*Port) Collection {
	for _, port := range ports {
		if port == nil {
			continue
		}
		collection[port.Name()] = port
	}

	return collection
}
