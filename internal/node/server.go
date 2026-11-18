
package node

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Kader1680/forgestore/internal/storage"
)

type Server struct {
	store storage.Store
}

func NewServer(store storage.Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("PUT /chunks/{id}", s.handlePutChunk)
	mux.HandleFunc("GET /chunks/{id}", s.handleGetChunk)
	mux.HandleFunc("DELETE /chunks/{id}", s.handleDeleteChunk)

	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func (s *Server) handlePutChunk(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, "missing chunk id", http.StatusBadRequest)
		return
	}

	err := s.store.Put(id, r.Body)
	if err != nil {
		if strings.Contains(err.Error(), "invalid storage key") {
			http.Error(w, "invalid chunk id", http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to store chunk", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "chunk stored")
}

func (s *Server) handleGetChunk(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	chunk, err := s.store.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid storage key") {
			http.Error(w, "invalid chunk id", http.StatusBadRequest)
			return
		}

		http.Error(w, "chunk not found", http.StatusNotFound)
		return
	}
	defer chunk.Close()

	w.Header().Set("Content-Type", "application/octet-stream")

	if _, err := io.Copy(w, chunk); err != nil {
		return
	}
}

func (s *Server) handleDeleteChunk(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := s.store.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "invalid storage key") {
			http.Error(w, "invalid chunk id", http.StatusBadRequest)
			return
		}

		http.Error(w, "chunk not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
