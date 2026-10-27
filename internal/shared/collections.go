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
