package storage

type URLsStorage interface {
	Get(id string) (string, error)
	Set(url string, id string) error
	FindByURL(url string) (string, bool)
	Close() error
}
