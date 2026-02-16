package storage

type mapRepository struct {
	urls map[string]string
}

func NewMapRepository() *mapRepository {
	return &mapRepository{
		urls: make(map[string]string),
	}
}

func (rep *mapRepository) Get(id string) (string, error) {
	url, ok := rep.urls[id]
	if !ok {
		return "", nil
	}
	return url, nil
}

func (rep *mapRepository) Set(url string, id string) error {
	rep.urls[id] = url
	return nil
}

func (rep *mapRepository) FindByURL(url string) (string, bool) {
	for id, u := range rep.urls {
		if u == url {
			return id, true
		}
	}
	return "", false
}
func (rep *mapRepository) Close() error {
	return nil
}
