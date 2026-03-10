package shared

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return Set[T]{}
} 

func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

func (s Set[T]) Contains(value T) bool {
	_, exists := s[value]
	return exists
}

func (s Set[T]) Values() []T {
	keys := make([]T, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}
