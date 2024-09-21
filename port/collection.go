package port

import (
	"github.com/hovsep/fmesh/signal"
)

// Collection is a port collection with useful methods
type Collection map[string]*Port

// Metadata contains metadata about the port
type Metadata struct {
	SignalBufferLen int
}

type MetadataMap map[string]*Metadata

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

// WithSignals adds signals to every port in collection and returns the collection
func (collection Collection) WithSignals(signals ...*signal.Signal) Collection {
	collection.PutSignals(signals...)
	return collection
}

// ClearSignals removes signals from all ports in collection
func (collection Collection) ClearSignals() {
	for _, p := range collection {
		p.ClearSignals()
	}
}

// Flush flushes all ports in collection
func (collection Collection) Flush(clearFlushed bool) {
	for _, p := range collection {
		p.Flush(clearFlushed)
	}
}

// PipeTo creates pipes from each port in collection
func (collection Collection) PipeTo(toPorts ...*Port) {
	for _, p := range collection {
		p.PipeTo(toPorts...)
	}
}

// Add adds ports to collection
func (collection Collection) Add(ports ...*Port) Collection {
	for _, port := range ports {
		if port == nil {
			continue
		}
		collection[port.Name()] = port
	}

	return collection
}

// AddIndexed creates ports with names like "o1","o2","o3" and so on
func (collection Collection) AddIndexed(prefix string, startIndex int, endIndex int) Collection {
	return collection.Add(NewIndexedGroup(prefix, startIndex, endIndex)...)
}

func (collection Collection) AllSignals() signal.Group {
	group := signal.NewGroup()
	for _, p := range collection {
		group = append(group, p.Signals()...)
	}
	return group
}

// GetPortsMetadata returns info about current length of each port in collection
func (collection Collection) GetPortsMetadata() MetadataMap {
	res := make(MetadataMap)
	for _, p := range collection {
		res[p.Name()] = &Metadata{
			SignalBufferLen: len(p.Signals()),
		}
	}
	return res
}

func (collection Collection) DisposeProcessedSignals(portsMetadata MetadataMap) {
	for pName, meta := range portsMetadata {
		collection.ByName(pName).DisposeFirstNSignals(meta.SignalBufferLen)
	}
}

func (collection Collection) FlushProcessedSignals(portsMetadata MetadataMap) {
	for pName, meta := range portsMetadata {
		collection.ByName(pName).FlushAndDisposeNSignals(meta.SignalBufferLen)
	}
}
