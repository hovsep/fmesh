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

// PutSignals adds signals to every port in collection
func (collection Collection) PutSignals(signals ...*signal.Signal) {
	for _, p := range collection {
		p.PutSignals(signals...)
	}
}

// withSignals adds signals to every port in collection and returns the collection
func (collection Collection) withSignals(signals ...*signal.Signal) Collection {
	collection.PutSignals(signals...)
	return collection
}

// Clear clears all ports in collection
func (collection Collection) Clear() {
	for _, p := range collection {
		p.Clear()
	}
}

// Flush flushes all ports in collection
func (collection Collection) Flush() {
	for _, p := range collection {
		p.Flush()
	}
}

// PipeTo creates pipes from each port in collection to given destination ports
func (collection Collection) PipeTo(destPorts ...*Port) {
	for _, p := range collection {
		p.PipeTo(destPorts...)
	}
}

// With adds ports to collection and returns it
func (collection Collection) With(ports ...*Port) Collection {
	for _, port := range ports {
		collection[port.Name()] = port
	}

	return collection
}

// WithIndexed creates ports with names like "o1","o2","o3" and so on
func (collection Collection) WithIndexed(prefix string, startIndex int, endIndex int) Collection {
	return collection.With(NewIndexedGroup(prefix, startIndex, endIndex)...)
}

// Signals returns all signals of all ports in the group
func (collection Collection) Signals() signal.Group {
	group := signal.NewGroup()
	for _, p := range collection {
		group = append(group, p.Signals()...)
	}
	return group
}
