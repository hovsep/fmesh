package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Collection is a port collection with useful methods
type Collection map[string]*Port

// NewPortsCollection creates empty collection
func NewPortsCollection() Collection {
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

// AnyHasSignal returns true if at least one port in collection has signal
func (collection Collection) AnyHasSignal() bool {
	for _, p := range collection {
		if p.HasSignal() {
			return true
		}
	}

	return false
}

// AllHaveSignal returns true when all ports in collection have signal
func (collection Collection) AllHaveSignal() bool {
	for _, p := range collection {
		if !p.HasSignal() {
			return false
		}
	}

	return true
}

// PutSignal puts a signal to all the port in collection
func (collection Collection) PutSignal(sig *signal.Signal) {
	for _, p := range collection {
		p.PutSignal(sig)
	}
}

// ClearSignal removes signals from all ports in collection
func (collection Collection) ClearSignal() {
	for _, p := range collection {
		p.ClearSignal()
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
