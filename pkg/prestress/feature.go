package prestress

type Feature interface {
	// Provider return value should be a function that accepts any number of
	// parameters (dependencies) and returns any number of results with last
	// return value optionally being an error. Provider should not have side
	// effects.
	Provider() any

	// Invoker return value should be a function that accepts any number of
	// parameters (dependencies) and returns error. Invoker should setup the
	// feature.
	Invoker() any
}
