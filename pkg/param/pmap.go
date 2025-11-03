package param

type ParamMap map[string]string

func (pmap ParamMap) Details() map[string]string {
	return pmap
}
