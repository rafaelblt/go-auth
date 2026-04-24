package shared

type Set[T comparable] struct {
	data map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{data: make(map[T]struct{})}
}

func NewSetFrom[T comparable](items ...T) Set[T] {
	set := NewSet[T]()
	for _, v := range items {
		set.Add(v)
	}
	return set
}

func (s Set[T]) Add(value T) {
	s.data[value] = struct{}{}
}

func (s Set[T]) Contains(value T) bool {
	_, exists := s.data[value]
	return exists
}

func (s Set[T]) Remove(value T) bool {
	_, ok := s.data[value]
	delete(s.data, value)
	return ok
}

func (s Set[T]) Values() []T {
	keys := make([]T, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s Set[T]) Len() int {
	return len(s.data)
}
