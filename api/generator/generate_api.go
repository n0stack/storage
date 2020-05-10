package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"

	"github.com/urfave/cli"
)

const template = `package %[1]s

import (
	context "context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/n0stack/storage/api"
	"github.com/n0stack/storage/api/database"
)

type %[2]sAPI struct {
	db database.Database

	BeforeCreate func(ctx context.Context, next *%[2]s) error
	BeforeUpdate func(ctx context.Context, prev, next *%[2]s) error
	BeforeApply  func(ctx context.Context, next *%[2]s) error
}

func Default%[2]sBeforeCreate(ctx context.Context, next *%[2]s) error {
	return nil
}
func Default%[2]sBeforeUpdate(ctx context.Context, prev, next *%[2]s) error {
	return nil
}
func Default%[2]sBeforeApply(ctx context.Context, next *%[2]s) error {
	return nil
}

func New%[2]sAPI(db database.Database) *%[2]sAPI {
	a := &%[2]sAPI{
		db:           db,
		BeforeCreate: Default%[2]sBeforeCreate,
		BeforeUpdate: Default%[2]sBeforeUpdate,
		BeforeApply:  Default%[2]sBeforeApply,
	}

	return a
}

func (a *%[2]sAPI) Apply%[2]s(ctx context.Context, req *%[2]s) (*%[2]s, error) {
	if err := api.ValidateName(req.Metadata.Name); err != nil {
		return nil, err
	}

	prev := &%[2]s{}
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

func (a *%[2]sAPI) List%[2]ss(ctx context.Context, req *List%[2]ssRequest) (*List%[2]ssResponse, error) {
	res := &List%[2]ssResponse{}
	f := func(s int) []database.Entity {
		res.%[2]ss = make([]*%[2]s, s)
		for i := range res.%[2]ss {
			res.%[2]ss[i] = &%[2]s{}
		}

		m := make([]database.Entity, s)
		for i, v := range res.%[2]ss {
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
	if len(res.%[2]ss) == 0 {
		return nil, api.NewNotFound("")
	}

	return res, nil
}

func (a *%[2]sAPI) Get%[2]s(ctx context.Context, req *Get%[2]sRequest) (*%[2]s, error) {
	if err := api.ValidateName(req.Name); err != nil {
		return nil, err
	}

	instance := &%[2]s{}
	if err := a.db.Get(ctx, req.Name, instance); err != nil {
		if err == database.ErrorNotFound {
			return nil, api.NewNotFound(req.Name)
		}

		return nil, api.Wrapf(err, "db.Get()")
	}

	return instance, nil
}

func (a *%[2]sAPI) Delete%[2]s(ctx context.Context, req *Delete%[2]sRequest) (*empty.Empty, error) {
	if err := api.ValidateName(req.Name); err != nil {
		return nil, err
	}

	prev := &%[2]s{}
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
`

func Generate(cctx *cli.Context) error {
	pkg := cctx.GlobalString("package")
	entity := cctx.GlobalString("entity")
	output := cctx.GlobalString("output")

	f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrapf(err, "os.Open(%v)", output)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, template, pkg, entity); err != nil {
		return errors.Wrap(err, "fmt.Fprintf()")
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "api-generator"
	// app.Version = ""
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "output,o",
		},
		cli.StringFlag{
			Name: "package,p",
		},
		cli.StringFlag{
			Name: "entity,e",
		},
	}
	app.Action = Generate

	if err := app.Run(os.Args); err != nil {
		log.Fatalf(err.Error())
	}
}
