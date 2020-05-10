package pstorage

import (
	context "context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/n0stack/storage/api"
	"github.com/n0stack/storage/api/database"
)

type DiskAPI struct {
	db database.Database

	BeforeCreate func(ctx context.Context, next *Disk) error
	BeforeUpdate func(ctx context.Context, prev, next *Disk) error
	BeforeApply  func(ctx context.Context, next *Disk) error
}

func DefaultDiskBeforeCreate(ctx context.Context, next *Disk) error {
	return nil
}
func DefaultDiskBeforeUpdate(ctx context.Context, prev, next *Disk) error {
	return nil
}
func DefaultDiskBeforeApply(ctx context.Context, next *Disk) error {
	return nil
}

func NewDiskAPI(db database.Database) *DiskAPI {
	a := &DiskAPI{
		db:           db,
		BeforeCreate: DefaultDiskBeforeCreate,
		BeforeUpdate: DefaultDiskBeforeUpdate,
		BeforeApply:  DefaultDiskBeforeApply,
	}

	return a
}

func (a *DiskAPI) ApplyDisk(ctx context.Context, req *Disk) (*Disk, error) {
	if err := api.ValidateName(req.Metadata.Name); err != nil {
		return nil, err
	}

	prev := &Disk{}
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

func (a *DiskAPI) ListDisks(ctx context.Context, req *ListDisksRequest) (*ListDisksResponse, error) {
	res := &ListDisksResponse{}
	f := func(s int) []database.Entity {
		res.Disks = make([]*Disk, s)
		for i := range res.Disks {
			res.Disks[i] = &Disk{}
		}

		m := make([]database.Entity, s)
		for i, v := range res.Disks {
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
	if len(res.Disks) == 0 {
		return nil, api.NewNotFound("")
	}

	return res, nil
}

func (a *DiskAPI) GetDisk(ctx context.Context, req *GetDiskRequest) (*Disk, error) {
	if err := api.ValidateName(req.Name); err != nil {
		return nil, err
	}

	instance := &Disk{}
	if err := a.db.Get(ctx, req.Name, instance); err != nil {
		if err == database.ErrorNotFound {
			return nil, api.NewNotFound(req.Name)
		}

		return nil, api.Wrapf(err, "db.Get()")
	}

	return instance, nil
}

func (a *DiskAPI) DeleteDisk(ctx context.Context, req *DeleteDiskRequest) (*empty.Empty, error) {
	if err := api.ValidateName(req.Name); err != nil {
		return nil, err
	}

	prev := &Disk{}
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
