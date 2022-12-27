package testtime

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TimeFunc func() time.Time

type MetaV1TimeFunc func() *metav1.Time

type ProtobufTimeFunc func() *timestamppb.Timestamp

// NewFrozenTime returns a FrozenTime for a given unix second.
func NewFrozenTimeUnix(unix int64) *FrozenTime {
	return &FrozenTime{t: time.Unix(unix, 0).UTC()}
}

// NewFrozenTime returns a FrozenTime for a given time.Time.
func NewFrozenTime(t time.Time) *FrozenTime {
	return &FrozenTime{t.UTC()}
}

// FrozenTime is a type for testing out fake times.
type FrozenTime struct {
	t time.Time
}

// Now never changes.
func (f *FrozenTime) Now() time.Time { return f.t }

// Before returns a time before FrozenTime.Now() by a given duration.
func (f *FrozenTime) Before(d time.Duration) time.Time { return f.Now().Add(-d) }

// After returns a time after FrozenTime.Now() by a given duration.
func (f *FrozenTime) After(d time.Duration) time.Time { return f.Now().Add(d) }

// Before Now() by int64 seconds.
func (f *FrozenTime) BeforeSec(s int64) time.Time {
	return f.Now().Add(time.Duration(-s) * time.Second)
}

// After Now() by int64 seconds.
func (f *FrozenTime) AfterSec(s int64) time.Time { return f.Now().Add(time.Duration(s) * time.Second) }

// BeforeFunc returns a TimeFunc where the return value is a time before FrozenTime.Now() by a given duration.
func (f *FrozenTime) BeforeFunc(d time.Duration) TimeFunc {
	return func() time.Time { return f.Before(d) }
}

// AfterFunc returns a TimeFunc where the return value is a time after FrozenTime.Now() by a given duration.
func (f *FrozenTime) AfterFunc(d time.Duration) TimeFunc {
	return func() time.Time { return f.After(d) }
}

func (f *FrozenTime) MetaV1Now() *metav1.Time { t := metav1.NewTime(f.Now()); return &t }
func (f *FrozenTime) MetaV1Before(d time.Duration) *metav1.Time {
	t := metav1.NewTime(f.Before(d))
	return &t
}

func (f *FrozenTime) MetaV1After(d time.Duration) *metav1.Time {
	t := metav1.NewTime(f.After(d))
	return &t
}

func (f *FrozenTime) MetaV1BeforeSec(s int64) *metav1.Time {
	t := metav1.NewTime(f.BeforeSec(s))
	return &t
}

func (f *FrozenTime) MetaV1AfterSec(s int64) *metav1.Time {
	t := metav1.NewTime(f.AfterSec(s))
	return &t
}

func (f *FrozenTime) MetaV1BeforeFunc(d time.Duration) MetaV1TimeFunc {
	return func() *metav1.Time { return f.MetaV1Before(d) }
}

func (f *FrozenTime) MetaV1AfterFunc(d time.Duration) MetaV1TimeFunc {
	return func() *metav1.Time { return f.MetaV1After(d) }
}

func (f *FrozenTime) PbNow() *timestamppb.Timestamp { return timestamppb.New(f.Now()) }
func (f *FrozenTime) PbBefore(d time.Duration) *timestamppb.Timestamp {
	return timestamppb.New(f.Before(d))
}

func (f *FrozenTime) PbAfter(d time.Duration) *timestamppb.Timestamp {
	return timestamppb.New(f.After(d))
}

func (f *FrozenTime) PbBeforeSec(s int64) *timestamppb.Timestamp {
	return timestamppb.New(f.BeforeSec(s))
}

func (f *FrozenTime) PbAfterSec(s int64) *timestamppb.Timestamp {
	return timestamppb.New(f.AfterSec(s))
}

func (f *FrozenTime) PbBeforeFunc(d time.Duration) ProtobufTimeFunc {
	return func() *timestamppb.Timestamp { return f.PbBefore(d) }
}

func (f *FrozenTime) PbAfterFunc(d time.Duration) ProtobufTimeFunc {
	return func() *timestamppb.Timestamp { return f.PbAfter(d) }
}
