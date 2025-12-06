package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage writes files to local disk under root directory.
type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) (*LocalStorage, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &LocalStorage{root: root}, nil
}

func (s *LocalStorage) Save(subPath string, r io.Reader) (string, error) {
	target := filepath.Join(s.root, subPath)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return subPath, nil
}

func (s *LocalStorage) PublicPath(subPath string) string {
	return fmt.Sprintf("/static/%s", subPath)
}
