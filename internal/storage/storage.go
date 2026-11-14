package storage

import "io"

type Store interface {
    Put(key string, src io.Reader) error
    // src is stream of byte that i want to write in disk
    Get(key string) (io.ReadCloser, error)
    Delete(key string) error
}

