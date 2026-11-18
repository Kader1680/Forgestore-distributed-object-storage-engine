
package storage

import (
    "io"
    "strings"
    "testing"
)

func TestPutAndGet(t *testing.T) {
    store, err := NewDiskStore(t.TempDir())
    if err != nil {
        t.Fatal(err)
    }

    original := "Hello ForgeStore"

    err = store.Put("chunk-001", strings.NewReader(original))
    if err != nil {
        t.Fatal(err)
    }

    chunk, err := store.Get("chunk-001")
    if err != nil {
        t.Fatal(err)
    }
    defer chunk.Close()

    result, err := io.ReadAll(chunk)
    if err != nil {
        t.Fatal(err)
    }

    if string(result) != original {
        t.Fatalf("expected %q, got %q", original, result)
    }
}

func TestDelete(t *testing.T) {
    store, err := NewDiskStore(t.TempDir())
    if err != nil {
        t.Fatal(err)
    }

    err = store.Put("chunk-001", strings.NewReader("data"))
    if err != nil {
        t.Fatal(err)
    }

    err = store.Delete("chunk-001")
    if err != nil {
        t.Fatal(err)
    }

    _, err = store.Get("chunk-001")
    if err == nil {
        t.Fatal("expected error after deleting chunk")
    }
}

func TestInvalidKey(t *testing.T) {
    store, err := NewDiskStore(t.TempDir())
    if err != nil {
        t.Fatal(err)
    }

    err = store.Put("../secret", strings.NewReader("data"))
    if err == nil {
        t.Fatal("expected invalid key error")
    }
}


func TestNewDiskStoreEmptyPath(t *testing.T) {
    _, err := NewDiskStore("")

    if err == nil {
        t.Fatal("expected error for empty storage path")
    }
}

func TestGetInvalidKey(t *testing.T) {
    store, err := NewDiskStore(t.TempDir())
    if err != nil {
        t.Fatal(err)
    }

    _, err = store.Get("../secret")

    if err == nil {
        t.Fatal("expected error for invalid key")
    }
}

func TestDeleteInvalidKey(t *testing.T) {
    store, err := NewDiskStore(t.TempDir())
    if err != nil {
        t.Fatal(err)
    }

    err = store.Delete("../secret")

    if err == nil {
        t.Fatal("expected error for invalid key")
    }
}