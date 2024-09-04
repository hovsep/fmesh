package main

type Components []*Component

func (components Components) byName(name string) *Component {
	for _, c := range components {
		if c.name == name {
			return c
		}
	}
	return nil
}
