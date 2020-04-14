package chunk

import "hash/crc64"

var CHUNK_CHECKSUM_TABLE = crc64.MakeTable(crc64.ECMA)

type ChunkInterface interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)

	Close() error
	Seek(offset int64, whence int) (n int64, err error)
	Sync() error

	Size() int64
	Checksum() uint64

	RLock()
	RUnlock()
	Lock()
	Unlock()
}
