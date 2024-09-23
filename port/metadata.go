package port

// Metadata contains metadata about the port
type Metadata struct {
	SignalBufferLen int
}

// MetadataMap contains port metadata indexed by port name
type MetadataMap map[string]*Metadata

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
