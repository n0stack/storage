package api

import (
	"fmt"
	"regexp"

	"github.com/golang/protobuf/ptypes"

	"github.com/google/uuid"
)

// NameValidationPattern is based on RFC 952
const NameValidationPattern = `^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]$`

var NameValidator = regexp.MustCompile(NameValidationPattern)

func ValidateName(name string) error {
	if !NameValidator.MatchString(name) {
		return WrapInvalidArgument(fmt.Errorf("the format of name field is %s", NameValidationPattern), "name")
	}

	return nil
}

func OnCreate(next *Metadata) error {
	if next.Uid == "" {
		id, err := uuid.NewRandom()
		if err != nil {
			return Wrapf(err, "uuid.NewRandom()")
		}

		next.Uid = id.String()
	} else {
		if _, err := uuid.Parse(next.Uid); err != nil {
			return WrapInvalidArgument(err, "metadata.uid")
		}
	}

	if next.Revision != 0 {
		return WrapInvalidArgument(fmt.Errorf("set 0 when creation"), "metadata.resource_version")
	}

	if next.CreatedAt != nil {
		return WrapInvalidArgument(fmt.Errorf("set blank"), "metadata.created_at")
	}
	next.CreatedAt = ptypes.TimestampNow()

	if next.DeletedAt != nil {
		return WrapInvalidArgument(fmt.Errorf("set blank"), "metadata.deleted_at")
	}

	return nil
}

func OnUpdate(prev, next *Metadata) error {
	if next.Uid == "" {
		next.Uid = prev.Uid
	}
	if next.Uid != prev.Uid {
		return WrapInvalidArgument(fmt.Errorf("mismatch stored and requested UIDs"), "metadata.uid")
	}

	if next.Revision == 0 {
		next.Revision = prev.Revision + 1
	}

	if next.CreatedAt != nil {
		return WrapInvalidArgument(fmt.Errorf("set blank"), "metadata.created_at")
	}
	next.CreatedAt = prev.CreatedAt

	if next.DeletedAt != nil {
		return WrapInvalidArgument(fmt.Errorf("set blank"), "metadata.deleted_at")
	}

	return nil
}
