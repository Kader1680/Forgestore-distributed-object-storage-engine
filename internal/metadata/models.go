
package metadata

import "time"

type File struct {
    ID          string
    Name        string
    Size        int64
    ContentType string
    CreatedAt   time.Time
}

type Chunk struct {
    ID         string
    FileID     string
    Index      int
    NodeID     string
    Size       int64
    Checksum   string
}
