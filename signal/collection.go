package signal

import (
	"fmt"
)

type Collection map[string]*Signal

func NewCollection() Collection {
	return make(Collection)
}

func (collection Collection) Add(signals ...*Signal) Collection {
	for _, sig := range signals {
		signalKey := collection.newKey(sig)
		collection[signalKey] = sig
	}
	return collection
}

func (collection Collection) AddPayload(payloads ...any) Collection {
	for _, p := range payloads {
		collection.Add(New(p))
	}
	return collection
}

func (collection Collection) newKey(signal *Signal) string {
	return fmt.Sprintf("%d", len(collection)+1)
}

func (collection Collection) AsGroup() Group {
	group := NewGroup()
	for _, sig := range collection {
		group = append(group, sig)
	}
	return group
}

func (collection Collection) FirstPayload() any {
	return collection.AsGroup().FirstPayload()
}

func (collection Collection) AllPayloads() []any {
	return collection.AsGroup().AllPayloads()
}

func (collection Collection) GetKeys() []string {
	keys := make([]string, 0)
	for k, _ := range collection {
		keys = append(keys, k)
	}
	return keys
}

func (collection Collection) DeleteKeys(keys []string) Collection {
	for _, key := range keys {
		delete(collection, key)
	}
	return collection
}
