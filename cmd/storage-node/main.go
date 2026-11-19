
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Kader1680/forgestore/internal/node"
	"github.com/Kader1680/forgestore/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data/node-1"
	}

	store, err := storage.NewDiskStore(dataDir)
	if err != nil {
		log.Fatal(err)
	}

	storageServer := node.NewServer(store)

	mux := storageServer.Handler()

	log.Printf("Storage Node running on :%s", port)
	log.Printf("Storage directory: %s", dataDir)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
