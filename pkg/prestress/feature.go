package prestress

type Feature interface {
	Apply(*Server) error
}
