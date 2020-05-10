package database

import (
	"context"
	"testing"

	"github.com/n0stack/storage/api"
)

func TestMemoryDatastore(t *testing.T) {
	m := NewMemoryDatabase()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	k := "test"
	v := &api.TestEntity{
		Metadata: &api.Metadata{
			Name: k,
		},
	}

	if err := m.Apply(ctx, v); err != nil {
		t.Fatalf("Apply('%v') err='%s'", v, err.Error())
	}

	e := &api.TestEntity{}
	if err := m.Get(ctx, k, e); err != nil {
		t.Errorf("Get() err='%s'", err.Error())
	} else if e == nil {
		t.Errorf("Get() result is nil")
	}

	res := []*api.TestEntity{}
	f := func(s int) []Entity {
		res = make([]*api.TestEntity, s)
		for i := range res {
			res[i] = &api.TestEntity{}
		}

		m := make([]Entity, s)
		for i, v := range res {
			m[i] = v
		}

		return m
	}
	if err := m.List(ctx, f); err != nil {
		t.Errorf("List() key='%s', value='%v', err='%s'", k, v, err.Error())
	}
	if len(res) != 1 {
		t.Errorf("List() number of listed keys is mismatch: have='%d', want='%d'", len(res), 1)
	}
	if res[0].Metadata.Name != v.Metadata.Name {
		t.Errorf("List() got 'Name' is wrong: key='%s', have='%s', want='%s'", k, res[0].Metadata.Name, v.Metadata.Name)
	}

	if err := m.SoftDelete(ctx, v); err != nil {
		t.Errorf("SoftDelete() err='%s'", err.Error())
	}
}

func TestMemoryDatastoreNotFound(t *testing.T) {
	m := NewMemoryDatabase()
	k := "test"

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	e := &api.TestEntity{
		Metadata: &api.Metadata{
			Name: "test",
		},
	}
	if err := m.Get(ctx, k, e); err != ErrorNotFound {
		t.Errorf("Get() return wrong error: got=%v, want=%v", err, ErrorNotFound)
	}

	res := []*api.TestEntity{}
	f := func(s int) []Entity {
		res = make([]*api.TestEntity, s)
		for i := range res {
			res[i] = &api.TestEntity{}
		}

		m := make([]Entity, s)
		for i, v := range res {
			m[i] = v
		}

		return m
	}
	if err := m.List(ctx, f); err != ErrorNotFound {
		t.Errorf("List() return wrong error: got=%v, want=%v", err, ErrorNotFound)
	}

	if err := m.SoftDelete(ctx, e); err != ErrorNotFound {
		t.Errorf("Delete() return wrong error: got=%v, want=%v", err, ErrorNotFound)
	}
}

// func TestConfliction(t *testing.T) {
// 	m := NewMemoryDatabase()
// 	k := "test"
// 	v := &api.TestEntity{
// 		Metadata: &api.Metadata{
// 			Name: k,
// 		},
// 	}

// 	ctx := context.Background()
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	if err := m.Apply(ctx, v); err != nil {
// 		t.Fatalf("Apply('%v', '%v', '%v') err='%s'", k, v, 0, err.Error())
// 	}
// 	if err := m.Apply(ctx, v); err != nil {
// 		t.Fatalf("Apply('%v', '%v', '%v') err='%s'", k, v, 1, err.Error())
// 	}

// 	v.Metadata.Revision = 0
// 	if err := m.Apply(ctx, v); err == nil {
// 		t.Errorf("Apply('%v', '%v', '%v') no error on applying confliction", k, v, 1)
// 	} else if _, ok := err.(datastore.ConflictedError); !ok {
// 		t.Errorf("Apply('%v', '%v', '%v') wrong error on applying confliction: err=%+v", k, v, 1, err)
// 	}

// 	if err := m.SoftDelete(ctx, v); err == nil {
// 		t.Errorf("Delete('%v', '%v') no error on applying confliction", k, 1)
// 	}

// 	v.Metadata.Revision = 1
// 	if err := m.SoftDelete(ctx, v); err != nil {
// 		t.Errorf("Apply('%v', '%v') err='%s'", k, 2, err.Error())
// 	}
// }

// func TestPrefixCollision(t *testing.T) {
// 	m := NewMemoryDatabase()

// 	prefix := "prefix"
// 	withPrefix := m.AddPrefix(prefix)

// 	k := "test"
// 	v := &datastore.Test{Name: "value"}

// 	ctx := context.Background()
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	if _, err := withPrefix.Apply(ctx, k, v, 0); err != nil {
// 		t.Fatalf("Failed to apply: err='%s'", err.Error())
// 	}
// 	e := &datastore.Test{}
// 	if _, err := m.Get(ctx, filepath.Join(prefix, k), e); err != nil {
// 		t.Errorf("Failed to get: err=%s", err.Error())
// 	}
// 	if e == nil || e.Name != v.Name {
// 		t.Errorf("Response is invalid")
// 	}

// 	k2 := "test"
// 	v2 := &datastore.Test{Name: "value"}

// 	if _, err := m.Apply(ctx, k2, v2, 0); err != nil {
// 		t.Fatalf("Failed to apply secondary: err='%s'", err.Error())
// 	}

// 	res := []*datastore.Test{}
// 	f := func(s int) []proto.Message {
// 		res = make([]*datastore.Test, s)
// 		for i := range res {
// 			res[i] = &datastore.Test{}
// 		}

// 		m := make([]proto.Message, s)
// 		for i, v := range res {
// 			m[i] = v
// 		}

// 		return m
// 	}
// 	if err := withPrefix.List(ctx, f); err != nil {
// 		t.Errorf("Failed to list: err='%s'", err.Error())
// 	}
// 	if len(res) != 1 {
// 		t.Errorf("Number of listed keys is mismatch: have='%d', want='%d'", len(res), 1)
// 	}
// }
