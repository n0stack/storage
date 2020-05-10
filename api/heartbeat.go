package api

import (
	"time"

	"github.com/golang/protobuf/ptypes"
)

func IsUp(hb *Heartbeat) bool {
	if hb == nil {
		return false
	}

	t, err := ptypes.Timestamp(hb.NextHeartbeat)
	if err != nil {
		// 一度も heartbeat をしていないと考えられる
		return false
	}

	return t.After(time.Now())
}

// NextHeartbeat returns nil if duration < 0
func NextHeartbeat(duration time.Duration) *Heartbeat {
	if duration < 0 {
		return nil
	}

	last, _ := ptypes.TimestampProto(time.Now())
	next, _ := ptypes.TimestampProto(time.Now().Add(duration))

	return &Heartbeat{
		LastHeartbeat: last,
		NextHeartbeat: next,
	}
}
