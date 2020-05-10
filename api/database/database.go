package database

import (
	"context"

	"github.com/n0stack/storage/api"

	"github.com/golang/protobuf/proto"
)

type Entity interface {
	proto.Message

	GetMetadata() *api.Metadata
}

type Database interface {
	List(ctx context.Context, f func(length int) []Entity) error
	Get(ctx context.Context, name string, entity Entity) error
	Apply(ctx context.Context, entity Entity) error
	SoftDelete(ctx context.Context, entity Entity) error
	HardDelete(ctx context.Context, entity Entity) error
}
