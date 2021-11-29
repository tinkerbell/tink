package tests

import (
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrozenTime is an interface for testing out fake times with additional
// methods for protobuf and Kubernetes time formats.
type FrozenTime interface {
	// Now never changes
	Now() time.Time
	Before(time.Duration) time.Time
	After(time.Duration) time.Time

	// Before Now() by int64 seconds
	BeforeSec(int64) time.Time
	// After Now() by int64 seconds
	AfterSec(int64) time.Time

	MetaV1Now() *metav1.Time
	MetaV1Before(time.Duration) *metav1.Time
	MetaV1After(time.Duration) *metav1.Time
	MetaV1BeforeSec(int64) *metav1.Time
	MetaV1AfterSec(int64) *metav1.Time

	PbNow() *timestamppb.Timestamp
	PbBefore(time.Duration) *timestamppb.Timestamp
	PbAfter(time.Duration) *timestamppb.Timestamp
	PbBeforeSec(int64) *timestamppb.Timestamp
	PbAfterSec(int64) *timestamppb.Timestamp
}

func NewFrozenTimeUnix(unix int64) FrozenTime {
	return &ft{t: time.Unix(unix, 0)}
}

func NewFrozenTime(t time.Time) FrozenTime {
	return &ft{t}
}

type ft struct {
	t time.Time
}

func (f *ft) Now() time.Time                   { return f.t }
func (f *ft) Before(d time.Duration) time.Time { return f.Now().Add(d) }
func (f *ft) After(d time.Duration) time.Time  { return f.Now().Add(-d) }
func (f *ft) BeforeSec(s int64) time.Time      { return f.Now().Add(time.Duration(-s) * time.Second) }
func (f *ft) AfterSec(s int64) time.Time       { return f.Now().Add(time.Duration(s) * time.Second) }

func (f *ft) MetaV1Now() *metav1.Time                   { t := metav1.NewTime(f.Now()); return &t }
func (f *ft) MetaV1Before(d time.Duration) *metav1.Time { t := metav1.NewTime(f.Before(d)); return &t }
func (f *ft) MetaV1After(d time.Duration) *metav1.Time  { t := metav1.NewTime(f.After(d)); return &t }
func (f *ft) MetaV1BeforeSec(s int64) *metav1.Time      { t := metav1.NewTime(f.BeforeSec(s)); return &t }
func (f *ft) MetaV1AfterSec(s int64) *metav1.Time       { t := metav1.NewTime(f.AfterSec(s)); return &t }

func (f *ft) PbNow() *timestamppb.Timestamp                   { return timestamppb.New(f.Now()) }
func (f *ft) PbBefore(d time.Duration) *timestamppb.Timestamp { return timestamppb.New(f.Before(d)) }
func (f *ft) PbAfter(d time.Duration) *timestamppb.Timestamp  { return timestamppb.New(f.After(d)) }
func (f *ft) PbBeforeSec(s int64) *timestamppb.Timestamp      { return timestamppb.New(f.BeforeSec(s)) }
func (f *ft) PbAfterSec(s int64) *timestamppb.Timestamp       { return timestamppb.New(f.AfterSec(s)) }
