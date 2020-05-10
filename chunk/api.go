package chunk

import (
	context "context"

	pstorage "github.com/n0stack/storage/v1alpha"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func BeforeUpdate(ctx context.Context, prev, next *pstorage.Chunk) error {
	if prev.Fact.Offset > next.Fact.Offset {
		return grpc.Errorf(codes.FailedPrecondition, "do not decrease fact.offset")
	}

	if prev.Fact.Offset == next.Fact.Offset && prev.Fact.Crc64 != next.Fact.Crc64 {
		return grpc.Errorf(codes.FailedPrecondition, "do not change fact.crc64 when not changed fact.offset")
	}

	return nil
}

func BeforeApply(ctx context.Context, next *pstorage.Chunk) error {
	if len(next.Spec.Disks) == 0 {
		return grpc.Errorf(codes.Unavailable, "set spec.disks more than 0")
	}

	if len(next.Fact.HealthDisks) == 0 {
		return grpc.Errorf(codes.Unavailable, "set fact.health_disks more than 0")
	}

	return nil
}
