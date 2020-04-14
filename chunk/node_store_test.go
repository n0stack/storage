package chunk

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	uuid "github.com/satori/go.uuid"
)

type ChunkTestNode struct {
	funcWriteChunk func(ctx context.Context, req *WriteChunkRequest) (*WriteChunkResponse, error)
}

func (n *ChunkTestNode) WriteChunk(ctx context.Context, req *WriteChunkRequest) (*WriteChunkResponse, error) {
	// time.Sleep(200 * time.Microsecond)
	return n.funcWriteChunk(ctx, req)
}

type TestReadCloser struct {
	b *bytes.Buffer
}

func (rc TestReadCloser) Read(b []byte) (int, error) {
	return rc.b.Read(b)
}
func (rc TestReadCloser) Close() error {
	return nil
}

func TestChunkStore(t *testing.T) {
	id, _ := uuid.NewV4()
	path := filepath.Join("./", id.String())
	open := func(id uuid.UUID) (ChunkInterface, error) {
		return OpenFileChunk(path, DEFAULT_MAX_CHUNK_SIZE)
	}
	defer os.Remove(path)

	n := OpenChunkStoreNode(open, &ChunkTestNode{funcWriteChunk: func(ctx context.Context, req *WriteChunkRequest) (*WriteChunkResponse, error) {
		if len(req.Nodes) != 0 {
			t.Errorf("called remote WriteChunk() with wrong request: want=%v, got=%v", []string{}, req.Nodes)
		}

		return &WriteChunkResponse{
			HealthNodes: req.Nodes,
			DesyncNodes: make([]string, 0),
		}, nil
	}})
	data := "Hello World!!"
	req := &WriteChunkRequest{
		ID:           id,
		Offset:       0,
		PreChecksum:  0,
		PostChecksum: 0xf5a8a397b60da2e1,
		Data:         TestReadCloser{b: bytes.NewBufferString(data)},
		Nodes:        []string{"self"},
	}
	expected := &WriteChunkResponse{
		HealthNodes: []string{"self"},
		DesyncNodes: []string{},
	}
	if res, err := n.WriteChunk(context.Background(), req); err != nil {
		t.Errorf("WriteChunk() returns err=%v", err)
	} else if diff := cmp.Diff(res, expected); diff != "" {
		t.Errorf("WriteChunk() wrong response (-got +want)\n%s", diff)
	}

	if buf, err := ioutil.ReadFile(path); err != nil {
		t.Errorf("ioutil.ReadAll(%s) returns err=%v", path, err)
	} else if string(buf) != data {
		t.Errorf("mismatch stored data: want=%s, got=%s", string(buf), data)
	}
}
