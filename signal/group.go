package signal

type Group []*Signal

// NewGroup creates empty group
func NewGroup(payloads ...any) Group {
	group := make(Group, len(payloads))
	for i, payload := range payloads {
		group[i] = New(payload)
	}
	return group
}

// FirstPayload returns the first signal payload
func (group Group) FirstPayload() any {
	//Intentionally not checking the group len
	//as the method does not have returning error (api is simpler)
	//and we can not just return nil, as nil may be a valid payload.
	//So just let runtime panic
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
