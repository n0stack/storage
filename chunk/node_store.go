package chunk

import (
	"context"
	"io"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	uuid "github.com/satori/go.uuid"
)

// TODO: connection pool的なやつをやる
// ファイルのコネクションプール的な役割とロック機構を作る

type ChunkStoreNode struct {
	open   func(uuid.UUID) (ChunkInterface, error)
	remote ChunkNode
}

func OpenChunkStoreNode(open func(uuid.UUID) (ChunkInterface, error), remote ChunkNode) *ChunkStoreNode {
	return &ChunkStoreNode{
		open:   open,
		remote: remote,
	}
}

func (n *ChunkStoreNode) WriteChunk(ctx context.Context, req *WriteChunkRequest) (*WriteChunkResponse, error) {
	eg, _ := errgroup.WithContext(ctx)
	fpr, fpw := io.Pipe()
	hpr, hpw := io.Pipe()

	eg.Go(func() error {
		defer fpw.Close()
		defer hpw.Close()

		// TODO: MultiWriterによって遅い方に引っ張られないか心配
		pw := io.MultiWriter(fpw, hpw)
		if _, err := io.Copy(pw, req.Data); err != nil {
			fpw.CloseWithError(errors.Wrap(err, "Copy()"))
			hpw.CloseWithError(errors.Wrap(err, "Copy()"))

			return errors.Wrap(err, "Copy()")
		}

		return nil
	})

	eg.Go(func() error {
		defer fpr.Close()
		// 片方が Read() しきらずに Close() されると、もう片方がエラーになるため、Pipeにすべてのデータが入れられることを待つために ReadAll() する
		// 正常系では io.EOF だけ帰ってくる
		defer ioutil.ReadAll(fpr)

		chunk, err := n.open(req.ID)
		if err != nil {
			return errors.Wrap(err, "Open()")
		}
		defer chunk.Close()

		// 同期が遅れているから非同期的に追いつくことを期待する
		if chunk.Size() < req.Offset {
			return errors.New("Delaying sync")
		}

		chunk.Lock()
		if _, err := chunk.Seek(req.Offset, io.SeekStart); err != nil {
			return errors.Wrap(err, "Seek()")
		}
		if chunk.Checksum() != req.PreChecksum {
			return errors.Errorf("mismatch pre-checksum: want=0x%016x, have=0x%016x", chunk.Checksum(), req.PreChecksum)
		}
		if _, err := io.Copy(chunk, fpr); err != nil {
			return errors.Wrap(err, "Copy()")
		}
		if chunk.Checksum() != req.PostChecksum {
			return errors.Errorf("mismatch post-checksum: want=0x%016x, have=0x%016x", chunk.Checksum(), req.PostChecksum)
		}
		chunk.Unlock()

		// TODO: Sync() までロックすべきなのかチェックする
		if err := chunk.Sync(); err != nil {
			return errors.Wrap(err, "Sync()")
		}

		return nil
	})

	// errorを返さない代わりにresponseにアクセスする
	res := &WriteChunkResponse{}
	eg.Go(func() (err error) {
		defer hpr.Close()
		defer ioutil.ReadAll(hpr)

		remoteReq := &WriteChunkRequest{
			ID:           req.ID,
			PreChecksum:  req.PreChecksum,
			PostChecksum: req.PostChecksum,
			Offset:       req.Offset,
			Data:         hpr,
			Nodes:        req.Nodes[1:],
		}
		res, err = n.remote.WriteChunk(ctx, remoteReq)

		return
	})

	if err := eg.Wait(); err != nil {
		log.Printf("[CRITICAL] desync err=%v", err)
		res.DesyncNodes = append(res.DesyncNodes, req.Nodes[0])

		return res, nil
	}
	res.HealthNodes = append(res.HealthNodes, req.Nodes[0])

	return res, nil
}
