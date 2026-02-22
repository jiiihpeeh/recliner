package store

type Atom[T any] struct {
	Store[T]
}

func NewAtom[T any](initial T) *Atom[T] {
	return &Atom[T]{Store: *NewStore(initial)}
}
