
package main

import (
    "context"
    "log"
    "net/http"
    "time"
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

    log.Println("Connected to PostgreSQL!")

    file := metadata.File{
        ID:          "file-001",
        Name:        "video.mp4",
        Size:        1024,
        ContentType: "video/mp4",
    }

    err = db.CreateFile(ctx, file)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("File metadata saved!")

    chunk := metadata.Chunk{
        ID:       "chunk-001",
        FileID:   "file-001",
        Index:    0,
        NodeID:   "node-1",
        Size:     1024,
        Checksum: "abc123",
    }
    
    err = db.CreateChunk(ctx, chunk)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Chunk metadata saved!")
    
    mux := http.NewServeMux()

    mux.HandleFunc("GET /health", func(
        w http.ResponseWriter,
        r *http.Request,
    ) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"service":"coordinator","status":"healthy"}`))
    })

    log.Println("Coordinator running on :8080")

    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal(err)
    }
}
