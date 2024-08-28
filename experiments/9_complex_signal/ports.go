package main

type Ports map[string]*Port

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
		if p.hasValue() {
			return true
		}
	}

	return false
}

func (ports Ports) allHaveValue() bool {
	for _, p := range ports {
		if !p.hasValue() {
			return false
		}
	}

	return true
}

func (ports Ports) setAll(val Signal) {
	for _, p := range ports {
		p.setValue(val)
	}
}

func (ports Ports) clearAll() {
	for _, p := range ports {
		p.clearValue()
	}
}
