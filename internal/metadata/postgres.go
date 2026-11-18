
package metadata

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
    pool *pgxpool.Pool
}

func NewPostgres(
    ctx context.Context,
    connectionString string,
) (*Postgres, error) {

    pool, err := pgxpool.New(ctx, connectionString)
    if err != nil {
        return nil, fmt.Errorf("create postgres pool: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("connect to postgres: %w", err)
    }

    return &Postgres{
        pool: pool,
    }, nil
}

func (p *Postgres) Close() {
    p.pool.Close()
}

func (p *Postgres) CreateFile(
    ctx context.Context,
    file File,
) error {

    query := `
        INSERT INTO files (
            id,
            name,
            size,
            content_type
        )
        VALUES ($1, $2, $3, $4)
    `

    _, err := p.pool.Exec(
        ctx,
        query,
        file.ID,
        file.Name,
        file.Size,
        file.ContentType,
    )

    if err != nil {
        return fmt.Errorf("create file metadata: %w", err)
    }

    return nil
}

func (p *Postgres) CreateChunk(
    ctx context.Context,
    chunk Chunk,
) error {

    query := `
        INSERT INTO chunks (
            id,
            file_id,
            chunk_index,
            node_id,
            size,
            checksum
        )
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    _, err := p.pool.Exec(
        ctx,
        query,
        chunk.ID,
        chunk.FileID,
        chunk.Index,
        chunk.NodeID,
        chunk.Size,
        chunk.Checksum,
    )

    if err != nil {
        return fmt.Errorf("create chunk metadata: %w", err)
    }

    return nil
}

func (p *Postgres) GetFile(
    ctx context.Context,
    id string,
) (File, error) {

    query := `
        SELECT id, name, size, content_type, created_at
        FROM files
        WHERE id = $1
    `

    var file File

    err := p.pool.QueryRow(
        ctx,
        query,
        id,
    ).Scan(
        &file.ID,
        &file.Name,
        &file.Size,
        &file.ContentType,
        &file.CreatedAt,
    )

    if err != nil {
        if err == pgx.ErrNoRows {
            return File{}, fmt.Errorf("file not found: %s", id)
        }

        return File{}, fmt.Errorf("get file metadata: %w", err)
    }

    return file, nil
}

func (p *Postgres) GetChunks(
    ctx context.Context,
    fileID string,
) ([]Chunk, error) {

    query := `
        SELECT id, file_id, chunk_index, node_id, size, checksum
        FROM chunks
        WHERE file_id = $1
        ORDER BY chunk_index ASC
    `

    rows, err := p.pool.Query(ctx, query, fileID)
    if err != nil {
        return nil, fmt.Errorf("get chunks: %w", err)
    }
    defer rows.Close()

    chunks := make([]Chunk, 0)

    for rows.Next() {
        var chunk Chunk

        err := rows.Scan(
            &chunk.ID,
            &chunk.FileID,
            &chunk.Index,
            &chunk.NodeID,
            &chunk.Size,
            &chunk.Checksum,
        )
        if err != nil {
            return nil, fmt.Errorf("scan chunk: %w", err)
        }

        chunks = append(chunks, chunk)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate chunks: %w", err)
    }

    return chunks, nil
}

func (p *Postgres) DeleteFile(
    ctx context.Context,
    id string,
) error {

    query := `DELETE FROM files WHERE id = $1`

    result, err := p.pool.Exec(ctx, query, id)
    if err != nil {
        return fmt.Errorf("delete file metadata: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("file not found: %s", id)
    }

    return nil
}
