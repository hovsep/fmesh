package main

// @TODO: this type must have good tooling for working with collection
// like adding new ports, filtering and so on
type Ports map[string]*Port

// @TODO: add error handling (e.g. when port does not exist)
func (ports Ports) byName(name string) *Port {
	return ports[name]
}

func (ports Ports) manyByName(names ...string) Ports {
	selectedPorts := make(Ports)

	for _, name := range names {
		if p, ok := ports[name]; ok {
			selectedPorts[name] = p
		}
	}

	return selectedPorts
}

func (ports Ports) anyHasValue() bool {
	for _, p := range ports {
		if p.hasSignal() {
			return true
		}
	}

	return false
}

func (ports Ports) allHaveValue() bool {
	for _, p := range ports {
		if !p.hasSignal() {
			return false
		}
	}

	return true
}

func (ports Ports) setAll(val Signal) {
	for _, p := range ports {
		p.putSignal(val)
	}
}

func (ports Ports) clearAll() {
	for _, p := range ports {
		p.clearSignal()
	}
}
