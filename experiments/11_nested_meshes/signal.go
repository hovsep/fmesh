package main

type Signal interface {
	IsAggregate() bool
	IsSingle() bool
	GetValue() any
}

// @TODO: enhance naming
type SingleSignal struct {
	val any
}

type AggregateSignal struct {
	val []*SingleSignal
}

func (s SingleSignal) IsAggregate() bool {
	return false
}

func (s SingleSignal) IsSingle() bool {
	return !s.IsAggregate()
}

func (s AggregateSignal) IsAggregate() bool {
	return true
}

func (s AggregateSignal) IsSingle() bool {
	return !s.IsAggregate()
}

func (s AggregateSignal) GetValue() any {
	return s.val
}

func (s SingleSignal) GetValue() any {
	return s.val
}

func newSignal(val any) *SingleSignal {
	return &SingleSignal{val: val}
}

func forwardSignal(source *Port, dest *Port) {
	dest.putSignal(source.getSignal())
}
