package main

type PortValue *int

type Port struct {
	val PortValue
}

func (p *Port) getValue() PortValue {
	return p.val
}

func (p *Port) setValue(val PortValue) {
	p.val = val
}

func (p *Port) hasValue() bool {
	return p.val != nil
}
