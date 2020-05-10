package pstorage

import (
	context "context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/n0stack/storage/api"
	"github.com/n0stack/storage/api/database"
)

type ChunkAPI struct {
	db database.Database

	BeforeCreate func(ctx context.Context, next *Chunk) error
	BeforeUpdate func(ctx context.Context, prev, next *Chunk) error
	BeforeApply  func(ctx context.Context, next *Chunk) error
}

func DefaultChunkBeforeCreate(ctx context.Context, next *Chunk) error {
	return nil
}
func DefaultChunkBeforeUpdate(ctx context.Context, prev, next *Chunk) error {
	return nil
}
func DefaultChunkBeforeApply(ctx context.Context, next *Chunk) error {
	return nil
}

func NewChunkAPI(db database.Database) *ChunkAPI {
	a := &ChunkAPI{
		db:           db,
		BeforeCreate: DefaultChunkBeforeCreate,
		BeforeUpdate: DefaultChunkBeforeUpdate,
		BeforeApply:  DefaultChunkBeforeApply,
	}

	return a
}

func (a *ChunkAPI) ApplyChunk(ctx context.Context, req *Chunk) (*Chunk, error) {
	if err := api.ValidateName(req.Metadata.Name); err != nil {
		return nil, err
	}

	prev := &Chunk{}
	if err := a.db.Get(ctx, req.Metadata.Name, prev); err != nil {
		if err == database.ErrorNotFound {
			if err := api.OnCreate(req.Metadata); err != nil {
				return nil, err
			}

			if err := a.BeforeCreate(ctx, req); err != nil {
				return nil, err
			}
		} else {
			return nil, api.Wrapf(err, "db.Get()")
		}
	} else {
		if err := api.OnUpdate(prev.Metadata, req.Metadata); err != nil {
			return nil, err
		}

		// ユーザーからのリクエストには Fact は含まれない可能性がある
		if req.Fact == nil {
			req.Fact = prev.Fact
		}

		if err := a.BeforeUpdate(ctx, prev, req); err != nil {
			return nil, err
		}
	}

	if err := a.BeforeApply(ctx, req); err != nil {
		return nil, err
	}

	if err := a.db.Apply(ctx, req); err != nil {
		return nil, api.Wrapf(err, "db.Apply()")
	}

	return req, nil
}

func (a *ChunkAPI) ListChunks(ctx context.Context, req *ListChunksRequest) (*ListChunksResponse, error) {
	res := &ListChunksResponse{}
	f := func(s int) []database.Entity {
		res.Chunks = make([]*Chunk, s)
		for i := range res.Chunks {
			res.Chunks[i] = &Chunk{}
		}

		m := make([]database.Entity, s)
		for i, v := range res.Chunks {
			m[i] = v
		}

		return m
	}

	if err := a.db.List(ctx, f); err != nil {
		if err == database.ErrorNotFound {
			return nil, api.NewNotFound("")
		}

		return nil, api.Wrapf(err, "db.List()")
	}
	if len(res.Chunks) == 0 {
		return nil, api.NewNotFound("")
	}

	return res, nil
}

func (a *ChunkAPI) GetChunk(ctx context.Context, req *GetChunkRequest) (*Chunk, error) {
	if err := api.ValidateName(req.Name); err != nil {
		return nil, err
	}

	instance := &Chunk{}
	if err := a.db.Get(ctx, req.Name, instance); err != nil {
		if err == database.ErrorNotFound {
			return nil, api.NewNotFound(req.Name)
		}

		return nil, api.Wrapf(err, "db.Get()")
	}

	return instance, nil
}

func (a *ChunkAPI) DeleteChunk(ctx context.Context, req *DeleteChunkRequest) (*empty.Empty, error) {
	if err := api.ValidateName(req.Name); err != nil {
		return nil, err
	}

	prev := &Chunk{}
	if err := a.db.Get(ctx, req.Name, prev); err != nil {
		if err == database.ErrorNotFound {
			return nil, api.NewNotFound(req.Name)
		}

		return nil, api.Wrapf(err, "db.Get()")
	}

	if err := a.db.SoftDelete(ctx, prev); err != nil {
		return nil, api.Wrapf(err, "db.Delete()")
	}

	return &empty.Empty{}, nil
}
