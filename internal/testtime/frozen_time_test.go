package testtime

import (
	"testing"
	"time"
)

func TestFrozenTime(t *testing.T) {
	cases := []struct {
		name          string
		beginTime     int64
		timeOffsetSec int64
		now           time.Time
		before        time.Time
		after         time.Time
	}{
		{
			name:          "a new hope premier",
			beginTime:     233391600,
			timeOffsetSec: 7260, // 121 minuets
			now:           time.Unix(233391600, 0),
			before:        time.Unix(233384340, 0),
			after:         time.Unix(233398860, 0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ft := NewFrozenTimeUnix(tc.beginTime)
			if !tc.now.Equal(ft.Now()) {
				t.Fatalf("Unexpected now: wanted %#v, got %#v", tc.now, ft.Now())
			}
			if !tc.before.Equal(ft.Before(time.Duration(tc.timeOffsetSec) * time.Second)) {
				t.Fatalf("Unexpected before: wanted %#v, got %#v", tc.before, ft.Before(time.Duration(tc.timeOffsetSec)*time.Second))
			}
			if !tc.after.Equal(ft.After(time.Duration(tc.timeOffsetSec) * time.Second)) {
				t.Fatalf("Unexpected after: wanted %#v, got %#v", tc.after, ft.After(time.Duration(tc.timeOffsetSec)*time.Second))
			}
			if !tc.before.Equal(ft.BeforeSec(tc.timeOffsetSec)) {
				t.Fatalf("Unexpected beforeSec: wanted %#v, got %#v", tc.before, ft.BeforeSec(tc.timeOffsetSec))
			}
			if !tc.after.Equal(ft.AfterSec(tc.timeOffsetSec)) {
				t.Fatalf("Unexpected afterSec: wanted %#v, got %#v", tc.after, ft.AfterSec(tc.timeOffsetSec))
			}
			if !tc.before.Equal(ft.BeforeFunc(time.Duration(tc.timeOffsetSec) * time.Second)()) {
				t.Fatalf("Unexpected beforeSec: wanted %#v, got %#v", tc.before, ft.BeforeFunc(time.Duration(tc.timeOffsetSec)*time.Second)())
			}
			if !tc.after.Equal(ft.AfterFunc(time.Duration(tc.timeOffsetSec) * time.Second)()) {
				t.Fatalf("Unexpected afterSec: wanted %#v, got %#v", tc.after, ft.AfterFunc(time.Duration(tc.timeOffsetSec)*time.Second)())
			}
		})
	}
}
