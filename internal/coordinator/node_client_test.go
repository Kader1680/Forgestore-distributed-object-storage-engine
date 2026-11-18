
package coordinator

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kader1680/forgestore/internal/node"
	"github.com/Kader1680/forgestore/internal/storage"
)

func TestNodeClientChunkLifecycle(t *testing.T) {
	store, err := storage.NewDiskStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	storageNode := node.NewServer(store)

	httpServer := httptest.NewServer(storageNode.Handler())
	defer httpServer.Close()

	client, err := NewNodeClient(httpServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	err = client.PutChunk(
		ctx,
		"chunk-001",
		strings.NewReader("Hello ForgeStore"),
	)
	if err != nil {
		t.Fatalf("PutChunk failed: %v", err)
	}

	chunk, err := client.GetChunk(ctx, "chunk-001")
	if err != nil {
		t.Fatalf("GetChunk failed: %v", err)
	}
	defer chunk.Close()

	data, err := io.ReadAll(chunk)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Hello ForgeStore" {
		t.Fatalf("expected %q, got %q",
			"Hello ForgeStore", string(data))
	}

	err = client.DeleteChunk(ctx, "chunk-001")
	if err != nil {
		t.Fatalf("DeleteChunk failed: %v", err)
	}

	_, err = client.GetChunk(ctx, "chunk-001")
	if err == nil {
		t.Fatal("expected error after deleting chunk")
	}
}
