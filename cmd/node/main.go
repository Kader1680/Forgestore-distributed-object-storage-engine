
package main

import (
	"log"
	"net/http"

	"github.com/Kader1680/forgestore/internal/node"
	"github.com/Kader1680/forgestore/internal/storage"
)

func main() {
	store, err := storage.NewDiskStore("./data")
	if err != nil {
		log.Fatal(err)
	}

	server := node.NewServer(store)

	log.Println("Storage Node listening on :8081")

	if err := http.ListenAndServe(":8081", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
