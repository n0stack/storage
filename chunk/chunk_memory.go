package chunk

import (
	"bytes"
	"hash/crc64"
	"io"
	"sync"

	"github.com/pkg/errors"
)

type MemoryChunk struct {
	sync.RWMutex

	writer *bytes.Buffer
	reader *bytes.Reader

	chunkSize int64
	checksum  uint64
}

func OpenMemoryChunk(chunkSize int64) *MemoryChunk {
	c := &MemoryChunk{}
	c.writer = bytes.NewBuffer(make([]byte, 0, chunkSize))
	c.reader = bytes.NewReader(c.writer.Bytes())
	c.chunkSize = chunkSize
	c.checksum = 0 // = crc64.Checksum(c.writer.Bytes(), CHECKSUM_TABLE)

	return c
}

func (c *MemoryChunk) updateReader() error {
	c.reader = bytes.NewReader(c.writer.Bytes())
	if _, err := c.reader.Seek(c.Size(), io.SeekStart); err != nil {
		return errors.Wrap(err, "Seek()")
	}

	return nil
}

func (c *MemoryChunk) Close() error {
	return nil
}

func (c MemoryChunk) Read(b []byte) (n int, err error) {
	n, err = c.reader.Read(b)

	return
}
func (c *MemoryChunk) Write(b []byte) (n int, err error) {
	offset, _ := c.reader.Seek(0, io.SeekCurrent)

	if c.chunkSize < int64(len(b))+offset {
		return 0, io.EOF
	}

	if c.writer.Len() > int(offset) {
		c.writer.Truncate(int(offset))

		// TODO: パフォーマンスにとって最善ではない
		c.checksum = crc64.Checksum(c.writer.Bytes(), CHUNK_CHECKSUM_TABLE)
	}

	n, err = c.writer.Write(b)
	if err != nil {
		return
	}

	c.checksum = crc64.Update(c.checksum, CHUNK_CHECKSUM_TABLE, b)
	err = c.updateReader()

	return
}

func (c MemoryChunk) Sync() error {
	return nil
}
func (c *MemoryChunk) Seek(offset int64, whence int) (n int64, err error) {
	n, err = c.reader.Seek(offset, whence)

	return
}

func (c MemoryChunk) Size() int64 {
	return int64(c.writer.Len())
}
func (c MemoryChunk) ChunkSize() int64 {
	return c.chunkSize
}
func (c MemoryChunk) Checksum() uint64 {
	return c.checksum
}
