
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Kader1680/forgestore/internal/coordinator"
	"github.com/Kader1680/forgestore/internal/metadata"
)

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	db, err := metadata.NewPostgres(
		ctx,
		"postgres://postgres:1234@localhost:5432/forgestore",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL")

	node1, err := coordinator.NewNodeClient("http://localhost:8081")
	if err != nil {
		log.Fatal(err)
	}

	node2, err := coordinator.NewNodeClient("http://localhost:8082")
	if err != nil {
		log.Fatal(err)
	}

	service := coordinator.NewService(
		db,
		map[string]*coordinator.NodeClient{
			"node-1": node1,
			"node-2": node2,
		},
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service":"coordinator","status":"healthy"}`))
	})

	mux.HandleFunc("POST /files", service.UploadHandler)
	mux.HandleFunc("GET /files/{id}/download", service.DownloadHandler)
	mux.HandleFunc("DELETE /files/{id}", service.DeleteHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Coordinator running on :8080")

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
