package prestress

import "context"

type ServerStartEvent struct {
	Context context.Context
}

func (ServerStartEvent) Event() string {
	return "ServerStartEvent"
}

type ServerShutdownEvent struct {
	Context context.Context
}

func (ServerShutdownEvent) Event() string {
	return "ServerShutdownEvent"
}
