package prestress

type Loggable interface {
	// Details used when logging
	Details() map[string]string
}
