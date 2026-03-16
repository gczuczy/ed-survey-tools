package types

type Set[T comparable] struct {
	items map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{items: make(map[T]struct{})}
}

func (s *Set[T]) Set(item T) {
	s.items[item] = struct{}{}
}

func (s *Set[T]) Unset(item T) {
	delete(s.items, item)
}

func (s Set[T]) IsSet(item T) bool {
	_, ok := s.items[item]
	return ok
}

func (s Set[T]) Len() int {
	return len(s.items)
}
