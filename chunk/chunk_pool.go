package chunk

type ChunkPool struct{}

func StartChunkPool(open func(key string) (ChunkInterface, error)) (*ChunkPool, error) {
	return nil, nil
}

func (c *ChunkPool) OpenChunk(key string) (ChunkInterface, error) {
	return nil, nil
}
