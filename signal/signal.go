package signal

// @TODO:rename\refactor interface
type SignalInterface interface {
	//Remove IsAggregate\IsSingle, in favour of Len
	IsAggregate() bool
	IsSingle() bool
	GetPayload() any
	AllPayloads() []any //@TODO: refactor with true iterator
}

type Signal struct {
	Payload any
}

type Signals struct {
	Payload []*Signal
}

func (s Signal) IsAggregate() bool {
	return false
}

func (s Signal) IsSingle() bool {
	return !s.IsAggregate()
}

func (s Signals) IsAggregate() bool {
	return true
}

func (s Signals) IsSingle() bool {
	return !s.IsAggregate()
}

func (s Signals) GetPayload() any {
	return s.Payload
}

func (s Signal) GetPayload() any {
	return s.Payload
}

func (s Signal) AllPayloads() []any {
	return []any{s.Payload}
}

func (s Signals) AllPayloads() []any {
	all := make([]any, 0)
	for _, sig := range s.Payload {
		all = append(all, sig.GetPayload())
	}
	return all
}

func New(payload any) *Signal {
	return &Signal{Payload: payload}
}
