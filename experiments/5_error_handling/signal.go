package main

type Signal interface {
	IsAggregate() bool
	IsSingle() bool
}

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

func (s SingleSignal) GetInt() int {
	return s.val.(int)
}
