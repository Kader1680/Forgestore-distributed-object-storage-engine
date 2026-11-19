package coordinator

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/Kader1680/forgestore/internal/metadata"
)

const ChunkSize = 4 * 1024 * 1024 // 4 MB

type MetadataStore interface {
	CreateFile(ctx context.Context, file metadata.File) error
	CreateChunk(ctx context.Context, chunk metadata.Chunk) error
	GetFile(ctx context.Context, id string) (metadata.File, error)
	GetChunks(ctx context.Context, fileID string) ([]metadata.Chunk, error)
	DeleteFile(ctx context.Context, id string) error
}

type Service struct {
	db    MetadataStore
	nodes map[string]*NodeClient
}

func NewService(
	db MetadataStore,
	nodes map[string]*NodeClient,
) *Service {
	return &Service{
		db:    db,
		nodes: nodes,
	}
}

type uploadedChunk struct {
	id     string
	nodeID string
}

func generateID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (s *Service) Upload(
	ctx context.Context,
	name string,
	contentType string,
	size int64,
	src io.Reader,
) (metadata.File, error) {
	if size < 0 {
		return metadata.File{}, fmt.Errorf("invalid file size")
	}

	if len(s.nodes) == 0 {
		return metadata.File{}, fmt.Errorf("no storage nodes available")
	}

	fileID, err := generateID()
	if err != nil {
		return metadata.File{}, err
	}

	file := metadata.File{
		ID:          fileID,
		Name:        name,
		Size:        size,
		ContentType: contentType,
	}

	if err := s.db.CreateFile(ctx, file); err != nil {
		return metadata.File{}, fmt.Errorf("create file metadata: %w", err)
	}

	rollback := func(uploaded []uploadedChunk) {
		for _, chunk := range uploaded {
			if node := s.nodes[chunk.nodeID]; node != nil {
				_ = node.DeleteChunk(ctx, chunk.id)
			}
		}
		_ = s.db.DeleteFile(ctx, fileID)
	}

	nodeIDs := make([]string, 0, len(s.nodes))
	for id := range s.nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)

	var uploaded []uploadedChunk
	var totalRead int64
	var chunkIndex int

	for totalRead < size {
		currentSize := int64(ChunkSize)
		if remaining := size - totalRead; remaining < currentSize {
			currentSize = remaining
		}

		buffer := make([]byte, int(currentSize))

		n, readErr := io.ReadFull(src, buffer)
		if readErr != nil {
			rollback(uploaded)
			return metadata.File{}, fmt.Errorf("read file data: %w", readErr)
		}

		chunkID := fileID + "-" + strconv.Itoa(chunkIndex)
		nodeID := nodeIDs[chunkIndex%len(nodeIDs)]

		node := s.nodes[nodeID]

		if err := node.PutChunk(
			ctx,
			chunkID,
			bytes.NewReader(buffer[:n]),
		); err != nil {
			rollback(uploaded)
			return metadata.File{}, fmt.Errorf("upload chunk to %s: %w", nodeID, err)
		}

		uploaded = append(uploaded, uploadedChunk{
			id:     chunkID,
			nodeID: nodeID,
		})

		chunk := metadata.Chunk{
			ID:       chunkID,
			FileID:   fileID,
			Index:    chunkIndex,
			NodeID:   nodeID,
			Size:     int64(n),
			Checksum: "",
		}

		if err := s.db.CreateChunk(ctx, chunk); err != nil {
			rollback(uploaded)
			return metadata.File{}, fmt.Errorf("save chunk metadata: %w", err)
		}

		totalRead += int64(n)
		chunkIndex++
	}

	return file, nil
}

func (s *Service) GetFile(
	ctx context.Context,
	id string,
) (metadata.File, []metadata.Chunk, error) {
	file, err := s.db.GetFile(ctx, id)
	if err != nil {
		return metadata.File{}, nil, fmt.Errorf("get file: %w", err)
	}

	chunks, err := s.db.GetChunks(ctx, id)
	if err != nil {
		return metadata.File{}, nil, fmt.Errorf("get chunks: %w", err)
	}

	return file, chunks, nil
}



func (s *Service) Download(
	ctx context.Context,
	fileID string,
	dst io.Writer,
) (metadata.File, error) {
	file, chunks, err := s.GetFile(ctx, fileID)
	if err != nil {
		return metadata.File{}, err
	}

	for _, chunk := range chunks {
		node, ok := s.nodes[chunk.NodeID]
		if !ok {
			return metadata.File{}, fmt.Errorf(
				"storage node %s not found",
				chunk.NodeID,
			)
		}

		reader, err := node.GetChunk(ctx, chunk.ID)
		if err != nil {
			return metadata.File{}, fmt.Errorf(
				"get chunk %s: %w",
				chunk.ID,
				err,
			)
		}

		_, copyErr := io.Copy(dst, reader)
		closeErr := reader.Close()

		if copyErr != nil {
			return metadata.File{}, fmt.Errorf(
				"write chunk %s: %w",
				chunk.ID,
				copyErr,
			)
		}
		if closeErr != nil {
			return metadata.File{}, fmt.Errorf(
				"close chunk %s: %w",
				chunk.ID,
				closeErr,
			)
		}
	}

	return file, nil
}

func (s *Service) Delete(
	ctx context.Context,
	fileID string,
) error {
	_, chunks, err := s.GetFile(ctx, fileID)
	if err != nil {
		return err
	}

	// Delete physical chunks first.
	for _, chunk := range chunks {
		node, ok := s.nodes[chunk.NodeID]
		if !ok {
			return fmt.Errorf(
				"storage node %s not found",
				chunk.NodeID,
			)
		}

		if err := node.DeleteChunk(ctx, chunk.ID); err != nil {
			return fmt.Errorf(
				"delete chunk %s: %w",
				chunk.ID,
				err,
			)
		}
	}

	// Delete metadata after physical chunks.
	if err := s.db.DeleteFile(ctx, fileID); err != nil {
		return fmt.Errorf("delete file metadata: %w", err)
	}

	return nil
}
