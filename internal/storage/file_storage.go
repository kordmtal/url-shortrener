package storage

import (
	"encoding/json"
	"os"
)

type shortURL struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type fileRepository struct {
	uuid int
	file *os.File
	urls []shortURL
}

func NewFileRepository(path string) (*fileRepository, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() != 0 {
		dataJSON, err := os.ReadFile(file.Name())
		if err != nil {
			return nil, err
		}
		var urls []shortURL
		if err := json.Unmarshal(dataJSON, &urls); err != nil {
			return nil, err
		}
		return &fileRepository{
			file: file,
			urls: urls,
			uuid: len(urls),
		}, nil
	}
	return &fileRepository{
		file: file,
		urls: []shortURL{},
		uuid: 0,
	}, nil
}

func (rep *fileRepository) Get(id string) (string, error) {
	for _, url := range rep.urls {
		if url.ShortURL == id {
			return url.OriginalURL, nil
		}
	}
	return "", nil
}

func (rep *fileRepository) Set(url string, id string) error {
	rep.uuid++
	rep.urls = append(rep.urls, shortURL{
		UUID:        rep.uuid,
		ShortURL:    id,
		OriginalURL: url,
	})
	data, err := json.MarshalIndent(rep.urls, "", "   ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(rep.file.Name(), data, 0666); err != nil {
		return err
	}
	return nil
}

func (rep *fileRepository) FindByURL(url string) (string, bool) {
	for _, u := range rep.urls {
		if u.OriginalURL == url {
			return u.ShortURL, true
		}
	}
	return "", false
}

func (rep *fileRepository) Close() error {
	if rep.file != nil {
		return rep.file.Close()
	}
	return nil
}
