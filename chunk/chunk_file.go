package chunk

import (
	"bytes"
	"hash/crc64"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pkg/errors"
)

type FileChunk struct {
	sync.RWMutex

	// writer *bytes.Buffer
	// reader *bytes.Reader

	// memory *MemoryChunk
	file *os.File

	chunkSize int64
	checksum  uint64
}

func OpenFileChunk(file string, chunkSize int64) (*FileChunk, error) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "os.OpenFile()")
	}

	c := &FileChunk{}
	c.file = f
	c.chunkSize = chunkSize

	buf, err := ioutil.ReadAll(c.file)
	if err != nil {
		return nil, errors.Wrap(err, "ioutilReadAll()")
	}
	c.checksum = crc64.Checksum(buf, CHUNK_CHECKSUM_TABLE)

	return c, nil
}

func (c *FileChunk) Close() error {
	c.Lock()
	defer c.Unlock()

	return c.file.Close()
}

func (c FileChunk) Read(b []byte) (n int, err error) {
	return c.file.Read(b)
}
func (c *FileChunk) Write(b []byte) (n int, err error) {
	offset, _ := c.file.Seek(0, io.SeekCurrent)

	if c.chunkSize < int64(len(b))+offset {
		return 0, io.EOF
	}

	if c.Size() > offset {
		buf := &bytes.Buffer{}
		if _, err := io.CopyN(buf, c.file, offset); err != nil {
			return 0, errors.Wrap(err, "io.CopyN()")
		}

		// TODO: パフォーマンスにとって最善ではない
		c.checksum = crc64.Checksum(buf.Bytes(), CHUNK_CHECKSUM_TABLE)
	}

	n, err = c.file.Write(b)
	if err != nil {
		return
	}

	c.checksum = crc64.Update(c.checksum, CHUNK_CHECKSUM_TABLE, b)

	return
}

// TODO: nの値の扱いについて
func (c *FileChunk) Seek(offset int64, whence int) (n int64, err error) {
	n, err = c.file.Seek(offset, whence)
	if err != nil {
		return 0, errors.Wrap(err, "file.Seek()")
	}

	return
}
func (c FileChunk) Sync() error {
	return c.file.Sync()
}

func (c FileChunk) Size() int64 {
	stat, err := c.file.Stat()
	if err != nil {
		return 0
	}

	return stat.Size()
}
func (c FileChunk) Checksum() uint64 {
	return c.checksum
}
