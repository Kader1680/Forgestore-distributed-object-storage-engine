package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Kader1680/forgestore/internal/coordinator"
	"github.com/Kader1680/forgestore/internal/metadata"
)

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := metadata.NewPostgres(
		ctx,
		env(
			"DATABASE_URL",
			"postgres://postgres:1234@localhost:5432/forgestore?sslmode=disable",
		),
	)
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL")

	nodeURLs := map[string]string{
		"node-1": env("NODE_1_URL", "http://localhost:8081"),
		"node-2": env("NODE_2_URL", "http://localhost:8082"),
		"node-3": env("NODE_3_URL", "http://localhost:8083"),
	}

	nodes := make(map[string]*coordinator.NodeClient)

	for id, url := range nodeURLs {
		client, err := coordinator.NewNodeClient(url)
		if err != nil {
			log.Fatalf("Node %s configuration failed: %v", id, err)
		}
		nodes[id] = client
	}

	service := coordinator.NewService(db, nodes)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("coordinator healthy"))
	})

	mux.HandleFunc("POST /files", service.UploadHandler)
	mux.HandleFunc("GET /files/{id}/download", service.DownloadHandler)
	mux.HandleFunc("DELETE /files/{id}", service.DeleteHandler)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Println("Coordinator running on :8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}