package signal

// Group represents a list of signals
type Group []*Signal

// NewGroup creates empty group
func NewGroup(payloads ...any) Group {
	group := make(Group, len(payloads))
	for i, payload := range payloads {
		group[i] = New(payload)
	}
	return group
}

// First returns the first signal in the group
func (group Group) First() *Signal {
	return group[0]
}

// FirstPayload returns the first signal payload
func (group Group) FirstPayload() any {
	return group.First().Payload()
}

// AllPayloads returns a slice with all payloads of the all signals in the group
func (group Group) AllPayloads() []any {
	all := make([]any, len(group), len(group))
	for i, sig := range group {
		all[i] = sig.Payload()
	}
	return all
}

// With returns the group with added signals
func (group Group) With(signals ...*Signal) Group {
	newGroup := make(Group, len(group)+len(signals))
	copy(newGroup, group)
	for i, sig := range signals {
		newGroup[len(group)+i] = sig
	}

	return newGroup
}

// WithPayloads returns a group with added signals created from provided payloads
func (group Group) WithPayloads(payloads ...any) Group {
	newGroup := make(Group, len(group)+len(payloads))
	copy(newGroup, group)
	for i, p := range payloads {
		newGroup[len(group)+i] = New(p)
	}
	return newGroup
}
