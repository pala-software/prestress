package prestress

import "context"

type ServerStartEvent struct {
	Context context.Context
}

func (ServerStartEvent) Event() string {
	return "ServerStartEvent"
}

func (ServerStartEvent) Details() map[string]string {
	return map[string]string{}
}

type ServerShutdownEvent struct {
	Context context.Context
}

func (ServerShutdownEvent) Event() string {
	return "ServerShutdownEvent"
}

func (ServerShutdownEvent) Details() map[string]string {
	return map[string]string{}
}
