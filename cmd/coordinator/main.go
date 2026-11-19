
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
