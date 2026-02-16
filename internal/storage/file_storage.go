package storage

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
	// Создаем директории, если они не существуют
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	var urls []shortURL
	decoder := json.NewDecoder(file)
	for {
		var url shortURL
		if err := decoder.Decode(&url); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return &fileRepository{
		file: file,
		urls: urls,
		uuid: len(urls),
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
	newURL := shortURL{
		UUID:        rep.uuid,
		ShortURL:    id,
		OriginalURL: url,
	}
	rep.urls = append(rep.urls, newURL)

	encoder := json.NewEncoder(rep.file)
	if err := encoder.Encode(newURL); err != nil {
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
