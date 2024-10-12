package journal

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestJournal(t *testing.T) {
	type input struct {
		msg  string
		args []any
	}
	tests := map[string]struct {
		want   []Entry
		inputs []input
	}{
		"empty": {
			want: []Entry{},
		},
		"single": {
			want: []Entry{
				{
					Msg:  "one",
					Args: map[string]any{"key": "value"},
					Source: slog.Source{
						File:     "journal_test.go",
						Function: "func1()",
					},
				},
			},
			inputs: []input{
				{msg: "one", args: []any{"key", "value"}},
			},
		},
		"non normal key": {
			want: []Entry{
				{
					Msg:  "msg",
					Args: map[string]any{"1.1": "value"},
					Source: slog.Source{
						File:     "journal_test.go",
						Function: "func1()",
					},
				},
			},
			inputs: []input{
				{msg: "msg", args: []any{1.1, "value"}},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := New(context.Background())
			for _, input := range tc.inputs {
				Log(ctx, input.msg, input.args...)
			}
			got := Journal(ctx)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreFields(Entry{}, "Time"), cmpopts.IgnoreFields(slog.Source{}, "Line")); diff != "" {
				t.Errorf("unexpected journal (-want +got):\n%s", diff)
			}
		})
	}
}
