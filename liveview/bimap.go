package liveview

import "sync"

type BiMap[T1 comparable, T2 comparable] struct {
	m  sync.Mutex
	m1 map[T1]T2
	m2 map[T2]T1
}

func NewBiMap[T1 comparable, T2 comparable]() *BiMap[T1, T2] {
	return &BiMap[T1, T2]{
		m1: make(map[T1]T2),
		m2: make(map[T2]T1),
	}
}

func (b *BiMap[T1, T2]) Get(key T1) (T2, bool) {
	b.m.Lock()
	defer b.m.Unlock()
	v, ok := b.m1[key]
	return v, ok
}

func (b *BiMap[T1, T2]) Set(key T1, value T2) {
	b.m.Lock()
	defer b.m.Unlock()
	if _, ok := b.m1[key]; ok {
		return
	}
	if _, ok := b.m2[value]; ok {
		return
	}

	b.m1[key] = value
	b.m2[value] = key

}

func (b *BiMap[T1, T2]) GetByValue(value T2) (T1, bool) {
	b.m.Lock()
	defer b.m.Unlock()
	v, ok := b.m2[value]
	return v, ok
}

func (b *BiMap[T1, T2]) Delete(key T1) {
	b.m.Lock()
	defer b.m.Unlock()
	delete(b.m2, b.m1[key])
	delete(b.m1, key)
}

func (b *BiMap[T1, T2]) DeleteByValue(value T2) {
	b.m.Lock()
	defer b.m.Unlock()
	delete(b.m1, b.m2[value])
	delete(b.m2, value)
}

func (b *BiMap[T1, T2]) GetAll() map[T1]T2 {
	b.m.Lock()
	defer b.m.Unlock()
	return b.m1
}

func (b *BiMap[T1, T2]) GetAllValues() map[T2]T1 {
	b.m.Lock()
	defer b.m.Unlock()
	return b.m2
}
