package storage

type Repositories interface {
	Set(kind, name, value string) error
	Get(kind, name string) (string, error)
	GetAll() ([]string, error)
}
