package signal

type Group []*Signal

// NewGroup creates empty group
func NewGroup(payloads ...any) Group {
	group := make(Group, 0)
	for _, payload := range payloads {
		group = append(group, New(payload))
	}
	return group
}

// FirstPayload returns the first signal payload in a group
func (group Group) FirstPayload() any {
	if len(group) == 0 {
		return nil
	}

	return group[0].Payload()
}

// AllPayloads returns a slice with all payloads of the all signals in the group
func (group Group) AllPayloads() []any {
	all := make([]any, 0)
	for _, s := range group {
		all = append(all, s.Payload())
	}
	return all
}
