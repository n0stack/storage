package chunk

import "hash/crc64"

var CHUNK_CHECKSUM_TABLE = crc64.MakeTable(crc64.ECMA)
