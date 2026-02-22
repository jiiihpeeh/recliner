package store

import "maps"

type Map[K comparable, V any] struct {
	Store[map[K]V]
}

func NewMap[K comparable, V any](initial map[K]V) *Map[K, V] {
	copyInit := make(map[K]V)
	maps.Copy(copyInit, initial)
	return &Map[K, V]{Store: *NewStore(copyInit)}
}

func (m *Map[K, V]) SetKey(key K, value V) {
	m.Update(func(current map[K]V) map[K]V {
		if current == nil {
			current = make(map[K]V)
		}
		newMap := make(map[K]V, len(current)+1)
		maps.Copy(newMap, current)
		newMap[key] = value
		return newMap
	})
}

func (m *Map[K, V]) DeleteKey(key K) {
	m.Update(func(current map[K]V) map[K]V {
		if current == nil {
			return nil
		}
		newMap := make(map[K]V, len(current))
		for k, v := range current {
			if k == key {
				continue
			}
			newMap[k] = v
		}
		return newMap
	})
}
