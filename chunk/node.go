package chunk

import (
	"context"
	"io"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

const DEFAULT_MAX_CHUNK_SIZE int64 = 128 * 1024 * 1024 // = 128 MB

type WriteChunkRequest struct {
	ID uuid.UUID

	PreChecksum  uint64
	PostChecksum uint64
	Offset       int64
	Data         io.Reader

	Nodes []string
}

type WriteChunkResponse struct {
	HealthNodes []string `json:"health_nodes"`
	DesyncNodes []string `json:"desync_nodes"`
}

var (
	ErrorExceededChunkSize = errors.New("ExceededChunkSize") // = http.StatusBadRequest
)

type ChunkNode interface {
	WriteChunk(ctx context.Context, req *WriteChunkRequest) (*WriteChunkResponse, error)
	// ReadChunk(ctx context.Context, req *WriteChunkRequest) (*WriteChunkResponse, error)
}
