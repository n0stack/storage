package database

import (
	"context"
	"strings"

	"github.com/golang/protobuf/proto"
)

type MemoryDatabase struct {
	// 本当は `proto.Message` を入れたいが、何故か中身がなかったのでとりあえずシリアライズする
	Data    map[string][]byte
	Deleted map[string][]byte

	prefix string
}

func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{
		Data:    make(map[string][]byte),
		Deleted: make(map[string][]byte),
	}
}

// func (m *MemoryDatabase) AddPrefix(prefix string) datastore.Datastore {
// 	return &MemoryDatabase{
// 		Data:   m.Data,
// 		prefix: m.prefix + prefix + "/",
// 	}
// }

func (m MemoryDatabase) List(ctx context.Context, f func(length int) []Entity) error {
	l := 0
	for k, _ := range m.Data {
		if strings.HasPrefix(k, m.prefix) {
			l++
		}
	}

	if l == 0 {
		return ErrorNotFound
	}

	pb := f(l)
	i := 0
	for k, v := range m.Data {
		if !strings.HasPrefix(k, m.prefix) {
			continue
		}

		if err := proto.Unmarshal(v, pb[i]); err != nil {
			return err
		}

		i++
	}

	return nil
}

func (m MemoryDatabase) Get(ctx context.Context, name string, entity Entity) error {
	v, ok := m.Data[m.getKey(name)]
	if !ok {
		entity = nil
		return ErrorNotFound
	}

	if err := proto.Unmarshal(v, entity); err != nil {
		return err
	}

	return nil
}

func (m *MemoryDatabase) Apply(ctx context.Context, entity Entity) error {
	// if v, ok := m.Data[m.getKey(entity.GetMetadata().Name)]; ok {
	// 	if v.GetMetadata().Revision+1 != entity.GetMetadata().Revision {
	// 		return ErrorNotFound
	// 	}
	// }

	var err error
	m.Data[m.getKey(entity.GetMetadata().Name)], err = proto.Marshal(entity)
	if err != nil {
		return err
	}

	return nil
}

func (m *MemoryDatabase) SoftDelete(ctx context.Context, entity Entity) error {
	md := entity.GetMetadata()

	_, ok := m.Data[m.getKey(md.Name)]
	if ok {
		// if prev.GetMetadata().Revision+1 != md.Revision {
		// 	return ErrorNotFound
		// }

		delete(m.Data, m.getKey(md.Name))

		var err error
		m.Deleted[md.Uid], err = proto.Marshal(entity)
		if err != nil {
			return err
		}

		return nil
	}

	return ErrorNotFound
}

func (m *MemoryDatabase) HardDelete(ctx context.Context, entity Entity) error {
	md := entity.GetMetadata()

	_, ok := m.Deleted[m.getKey(md.Name)]
	if ok {
		// if prev.GetMetadata().Revision+1 != md.Revision {
		// 	return ErrorConflict
		// }

		delete(m.Deleted, m.getKey(md.Name))

		return nil
	}

	return ErrorNotFound
}

func (m MemoryDatabase) getKey(key string) string {
	return m.prefix + key
}
