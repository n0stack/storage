package api

import (
	"log"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func Wrapf(err error, format string, a ...interface{}) error {
	e := errors.Wrapf(err, format, a...)
	log.Printf("[WARNING] Wrapf(): %v", e)

	// TODO マスクする
	return grpc.Errorf(codes.Internal, e.Error())
}

func WrapInvalidArgument(err error, key string) error {
	return grpc.Errorf(codes.InvalidArgument, "%s: %s", key, err.Error())
}

func NewNotFound(key string) error {
	return grpc.Errorf(codes.NotFound, key)
}
