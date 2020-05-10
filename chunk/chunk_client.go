package chunk

import (
	"bytes"
	context "context"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	pstorage "github.com/n0stack/storage/v1alpha"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// type ChunkClient struct {
// 	api      pstorage.ChunkServiceClient
// 	metadata *pstorage.Chunk
// 	memory   *MemoryChunk

// 	// nextSyncOffset   int64
// 	// lastSyncChecksum uint64

// 	id    uuid.UUID
// 	nodes []string
// }

// func OpenChunkClientWithMaster(ctx context.Context, id uuid.UUID) (*ChunkClient, error) {
// 	mr, err := GetChunkMaster(ctx, id)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "GetChunkMaster()")
// 	}

// 	c := &ChunkClient{
// 		memory: OpenMemoryChunk(MAX_CHUNK_SIZE),

// 		id:    id,
// 		nodes: mr.nodes,
// 	}

// 	for _, node := range c.nodes {
// 		resp, err := http.Get(filepath.Join(node, c.id.String()))
// 		if err != nil {
// 			log.Printf("[WARNING] failed to get from %s, err=%v", node, err)
// 		} else {
// 			defer resp.Body.Close()

// 			if _, err := io.Copy(c.memory, resp.Body); err != nil {
// 				return nil, errors.Wrap(err, "io.Copy")
// 			}

// 			if c.memory.Checksum() != mr.Checksum {
// 				log.Printf("[WARNING] checksum mismatch about %s on %s, err=%v", id.String(), node, err)
// 				c.memory = OpenMemoryChunk(MAX_CHUNK_SIZE)

// 				continue
// 			}

// 			offset, err := c.memory.Seek(0, io.SeekCurrent)
// 			if err != nil {
// 				return nil, errors.Wrap(err, "Seek()")
// 			}
// 			if offset != mr.Offset {
// 				log.Printf("[WARNING] checksum mismatch about %s on %s, err=%v", id.String(), node, err)
// 				c.memory = OpenMemoryChunk(MAX_CHUNK_SIZE)

// 				continue
// 			}

// 			break
// 		}
// 	}

// 	return c, nil
// }

// func OpenChunkClient(ctx context.Context, id uuid.UUID) (*ChunkClient, error) {
// 	mr, err := GetChunkMaster(ctx, id)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "GetChunkMaster()")
// 	}

// 	c := &ChunkClient{
// 		memory: OpenMemoryChunk(MAX_CHUNK_SIZE),

// 		id:    id,
// 		nodes: mr.nodes,
// 	}

// 	for _, node := range c.nodes {
// 		resp, err := http.Get(filepath.Join(node, c.id.String()))
// 		if err != nil {
// 			log.Printf("[WARNING] failed to get from %s, err=%v", node, err)
// 		} else {
// 			defer resp.Body.Close()

// 			if _, err := io.Copy(c.memory, resp.Body); err != nil {
// 				return nil, errors.Wrap(err, "io.Copy")
// 			}

// 			if c.memory.Checksum() != mr.Checksum {
// 				log.Printf("[WARNING] checksum mismatch about %s on %s, err=%v", id.String(), node, err)
// 				c.memory = OpenMemoryChunk(MAX_CHUNK_SIZE)

// 				continue
// 			}

// 			offset, err := c.memory.Seek(0, io.SeekCurrent)
// 			if err != nil {
// 				return nil, errors.Wrap(err, "Seek()")
// 			}
// 			if offset != mr.Offset {
// 				log.Printf("[WARNING] checksum mismatch about %s on %s, err=%v", id.String(), node, err)
// 				c.memory = OpenMemoryChunk(MAX_CHUNK_SIZE)

// 				continue
// 			}

// 			break
// 		}
// 	}

// 	return c, nil
// }

type ChunkClient struct {
	api      pstorage.ChunkServiceClient
	metadata *pstorage.Chunk

	memory *MemoryChunk

	cancelAsyncSync context.CancelFunc
}

func OpenChunkClient(ctx context.Context, id uuid.UUID, api pstorage.ChunkServiceClient) (*ChunkClient, error) {
	c := &ChunkClient{
		memory: OpenMemoryChunk(DEFAULT_MAX_CHUNK_SIZE),
		api:    api,
	}

	// TODO: exponential backoff
	var err error
	c.metadata, err = c.api.GetChunk(ctx, &pstorage.GetChunkRequest{
		Name: id.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "api.GetChunk()")
	}

	for _, disk := range c.metadata.Spec.Disks {
		resp, err := http.Get(filepath.Join(disk.Fact.Url, id.String()))
		if err != nil {
			log.Printf("[WARNING] failed to get from %v, err=%v", disk, err)
		} else {
			defer resp.Body.Close()

			if _, err := io.Copy(c.memory, resp.Body); err != nil {
				return nil, errors.Wrap(err, "io.Copy")
			}

			if c.memory.Checksum() != c.metadata.Fact.Crc64 {
				log.Printf("[WARNING] checksum mismatch about %s on %v, err=%v", id.String(), disk, err)
				c.memory = OpenMemoryChunk(c.memory.ChunkSize())

				continue
			}

			offset, err := c.memory.Seek(0, io.SeekCurrent)
			if err != nil {
				return nil, errors.Wrap(err, "Seek()")
			}
			if offset != c.metadata.Fact.Offset {
				log.Printf("[WARNING] checksum mismatch about %s on %v, err=%v", id.String(), disk, err)
				c.memory = OpenMemoryChunk(c.memory.ChunkSize())

				continue
			}

			break
		}
	}

	ctx, c.cancelAsyncSync = context.WithCancel(context.Background())
	c.asyncSync(ctx)

	return c, nil
}

func (c *ChunkClient) asyncSync(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-t.C:
			c.Sync()
		}
	}
}

func (c *ChunkClient) Close() error {
	c.cancelAsyncSync()
	return c.Sync()
}

func (c ChunkClient) Read(b []byte) (n int, err error) {
	c.memory.RLock()
	defer c.memory.RUnlock()

	return c.memory.Read(b)
}
func (c ChunkClient) Write(b []byte) (n int, err error) {
	c.memory.Lock()
	defer c.memory.Unlock()

	return c.memory.Write(b)
}
func (c *ChunkClient) Seek(offset int64, whence int) (n int64, err error) {
	return c.memory.Seek(offset, whence)
}

func (c ChunkClient) Sync() error {
	bodyBuf := &bytes.Buffer{}
	body := multipart.NewWriter(bodyBuf)

	// TODO: set file name
	fw, err := body.CreateFormFile("data", "data")
	if err != nil {
		return errors.Wrap(err, "CreateFormFile()")
	}

	c.memory.Lock()
	s, err := c.memory.Seek(0, io.SeekCurrent)
	if err != nil {
		return errors.Wrap(err, "memory.Seek()")
	}

	if _, err := c.memory.Seek(c.metadata.Fact.Offset, io.SeekStart); err != nil {
		return errors.Wrap(err, "memory.Seek()")
	}

	if _, err := io.Copy(fw, c.memory); err != nil {
		return errors.Wrap(err, "io.Copy()")
	}

	if _, err := c.memory.Seek(s, io.SeekStart); err != nil {
		return errors.Wrap(err, "memory.Seek()")
	}
	c.memory.Unlock()

	contentType := body.FormDataContentType()
	body.Close()

	// for i, disk := range c.metadata.Spec.Disks {
	// 	values := url.Values{}
	// 	values.Add(WRITE_QUERY_PRE_CHECKSUM, fmt.Sprintf("%016x", c.metadata.Fact.Crc64))
	// 	values.Add(WRITE_QUERY_POST_CHECKSUM, fmt.Sprintf("%016x", c.memory.Checksum()))
	// 	values.Add(WRITE_QUERY_OFFSET, fmt.Sprintf("%d", c.metadata.Fact.Offset))
	// 	values.Add(WRITE_QUERY_REPLICA_NODES, strings.Join(c.nodes[i:], ","))

	// 	resp, err := http.Post(filepath.Join(node, c.id.String())+"?"+values.Encode(), contentType, bodyBuf)
	// 	if err != nil {
	// 		log.Printf("[WARNING] failed to replicate to %s, err=%v", node, err)
	// 	} else {
	// 		defer resp.Body.Close()

	// 		r, err := ioutil.ReadAll(resp.Body)
	// 		if err != nil {
	// 			return errors.Wrap(err, "ReadAll()")
	// 		}

	// 		wcr := &WriteChunkResponse{}
	// 		if err := json.Unmarshal(r, wcr); err != nil {
	// 			return errors.Wrap(err, "json.Unmarshal()")
	// 		}

	// 		c.metadata.Fact = &pstorage.ChunkFact{
	// 			Offset: c.memory.Size(),
	// 			Crc64:  c.memory.Checksum(),
	// 			// DesyncDisks:
	// 		}

	// 		break
	// 	}
	// }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := c.api.ApplyChunk(ctx, c.metadata)
	if err != nil {
		return errors.Wrap(err, "api.ApplyChunk()")
	}
	c.metadata = res

	return nil
}

// 	// c.api.ApplyChunk(ctx, &)

// 	// c.nextSyncOffset =
// 	// c.lastSyncChecksum =

// 	return nil
// }

func (c ChunkClient) Size() int64 {
	return c.memory.Size()
}
func (c ChunkClient) ChunkSize() int64 {
	return c.memory.ChunkSize()
}
func (c ChunkClient) Checksum() uint64 {
	return c.memory.Checksum()
}
