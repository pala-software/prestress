package prestress

type Event interface {
	// Description of the event
	Event() string

	// Details for logging
	Details() map[string]string
}

type EventListener func(event Event) error

func (server *Server) OnEvent(listener EventListener) {
	server.listeners = append(server.listeners, listener)
}

func (server Server) Emit(event Event) error {
	var err error
	for _, listener := range server.listeners {
		err = listener(event)
		if err != nil {
			return err
		}
	}
	return nil
}
