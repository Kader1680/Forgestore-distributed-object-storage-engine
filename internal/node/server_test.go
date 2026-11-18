
package node

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kader1680/forgestore/internal/storage"
)

func TestHealth(t *testing.T) {
	store, err := storage.NewDiskStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer(store)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d",
			http.StatusOK, rec.Code)
	}

	if strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Fatalf("unexpected response: %q", rec.Body.String())
	}
}

func TestChunkLifecycle(t *testing.T) {
	store, err := storage.NewDiskStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer(store)
	handler := server.Handler()

	// 1. PUT: upload a chunk
	putReq := httptest.NewRequest(
		http.MethodPut,
		"/chunks/chunk-001",
		strings.NewReader("Hello ForgeStore"),
	)
	putRec := httptest.NewRecorder()

	handler.ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusCreated {
		t.Fatalf("PUT: expected status %d, got %d",
			http.StatusCreated, putRec.Code)
	}

	// 2. GET: retrieve the chunk
	getReq := httptest.NewRequest(
		http.MethodGet,
		"/chunks/chunk-001",
		nil,
	)
	getRec := httptest.NewRecorder()

	handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET: expected status %d, got %d",
			http.StatusOK, getRec.Code)
	}

	body, err := io.ReadAll(getRec.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Hello ForgeStore" {
		t.Fatalf("expected %q, got %q",
			"Hello ForgeStore", string(body))
	}

	// 3. DELETE: remove the chunk
	deleteReq := httptest.NewRequest(
		http.MethodDelete,
		"/chunks/chunk-001",
		nil,
	)
	deleteRec := httptest.NewRecorder()

	handler.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("DELETE: expected status %d, got %d",
			http.StatusNoContent, deleteRec.Code)
	}

	// 4. GET after DELETE: chunk should not exist
	getAgainReq := httptest.NewRequest(
		http.MethodGet,
		"/chunks/chunk-001",
		nil,
	)
	getAgainRec := httptest.NewRecorder()

	handler.ServeHTTP(getAgainRec, getAgainReq)

	if getAgainRec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d after deletion, got %d",
			http.StatusNotFound, getAgainRec.Code)
	}
}
