package api

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
)

func TestHearbeatUp(t *testing.T) {
	last, _ := ptypes.TimestampProto(time.Now().Add(1 * time.Second))
	hb := &Heartbeat{
		NextHeartbeat: last,
	}
	if IsUp(hb) != true {
		t.Errorf("IsUp(%v): want=%v, got=%v", hb, true, IsUp(hb))
	}
}

func TestHearbeatDown(t *testing.T) {
	last, _ := ptypes.TimestampProto(time.Now().Add(-1 * time.Second))
	hb := &Heartbeat{
		NextHeartbeat: last,
	}
	if IsUp(hb) != false {
		t.Errorf("IsUp(%v): want=%v, got=%v", hb, false, IsUp(hb))
	}
}

func TestHearbeat(t *testing.T) {
	duration := 5 * time.Minute
	hb := NextHeartbeat(duration)

	next, err := ptypes.Timestamp(hb.NextHeartbeat)
	if err != nil {
		t.Errorf("NextHeartbeat is invalid: err=%v", err)
	}
	last, err := ptypes.Timestamp(hb.LastHeartbeat)
	if err != nil {
		t.Errorf("LastHeartbeat is invalid: err=%v", err)
	}

	// 丸め誤差で通らない
	if next.Sub(last)-duration > time.Millisecond {
		t.Errorf("next - last != duration: got=%v, want=%v", next.Sub(last), duration)
	}
}
