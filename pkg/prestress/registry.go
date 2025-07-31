package prestress

type ReadRegistry[T any] interface {
	Value() []T
}

type WriteRegistry[T any] interface {
	Register(T)
}

type RegistryInterface[T any] interface {
	ReadRegistry[T]
	WriteRegistry[T]
}

type Registry[T any] struct {
	value []T
}

func (reg *Registry[T]) Register(value T) {
	reg.value = append(reg.value, value)
}

func (reg Registry[T]) Value() []T {
	return reg.value
}
