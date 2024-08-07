package main

type Ports map[string]*Port

func (ports Ports) byName(name string) *Port {
	return ports[name]
}

func (ports Ports) anyHasValue() bool {
	for _, p := range ports {
		if p.hasValue() {
			return true
		}
	}

	return false
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
