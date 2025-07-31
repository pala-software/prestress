package prestress

type LifecycleHook func() error

type Lifecycle struct {
	Start    Registry[LifecycleHook]
	Shutdown Registry[LifecycleHook]
}

func NewLifecycle() *Lifecycle {
	return new(Lifecycle)
}
