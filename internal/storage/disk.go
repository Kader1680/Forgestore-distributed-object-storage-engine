package storage

import "io"

type DiskStore struct {
	BaseDir string
}

package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type DiskStore struct {
	BaseDir string
}

func NewDiskStore(baseDir string) (*DiskStore, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("storage directory cannot be empty")
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	return &DiskStore{
		BaseDir: baseDir,
	}, nil
}

func (d *DiskStore) filePath(key string) (string, error) {
	if key == "" || key == "." || key == ".." {
		return "", fmt.Errorf("invalid storage key")
	}

	for _, char := range key {
		valid := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.'

		if !valid {
			return "", fmt.Errorf("invalid storage key")
		}
	}

	return filepath.Join(d.BaseDir, key), nil
}

func (d *DiskStore) Put(key string, src io.Reader) error {
	destination, err := d.filePath(key)
	if err != nil {
		return err
	}
 
	tmp, err := os.CreateTemp(d.BaseDir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}

	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		return fmt.Errorf("write chunk: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync chunk: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}

	if err := os.Rename(tmpPath, destination); err != nil {
		return fmt.Errorf("save chunk: %w", err)
	}

	return nil
}

func (d *DiskStore) Get(key string) (io.ReadCloser, error) {
	path, err := d.filePath(key)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read chunk: %w", err)
	}

	return file, nil
}

func (d *DiskStore) Delete(key string) error {
	path, err := d.filePath(key)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete chunk: %w", err)
	}

	return nil
}

 
var _ Store = (*DiskStore)(nil)
