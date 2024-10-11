package journal

import (
	"context"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type CtxKey string

const Name CtxKey = "journal"

type Entry struct {
	Msg    string         `json:"msg"`
	Args   map[string]any `json:"args,omitempty"`
	Source slog.Source    `json:"source"`
	Time   string         `json:"time"`
}

// New creates a slice of Entries in the provided context.
func New(ctx context.Context) context.Context {
	e := &[]Entry{}
	return context.WithValue(ctx, Name, e)
}

// Log adds a new Entry to the journal in the provided context.
// Log is not thread-safe.
func Log(ctx context.Context, msg string, args ...any) {
	t := time.Now().UTC().Format(time.RFC3339Nano)
	m := make(map[string]any)
	for i := 0; i < len(args); i += 2 {
		m[args[i].(string)] = args[i+1]
	}
	e, ok := ctx.Value(Name).(*[]Entry)
	if !ok {
		e = &[]Entry{{Msg: msg, Args: m, Source: fileAndLine(), Time: t}}
		return
	}
	*e = append(*e, Entry{Msg: msg, Args: m, Source: fileAndLine(), Time: t})
}

// Journal returns the journal from the provided context.
func Journal(ctx context.Context) []Entry {
	e, ok := ctx.Value(Name).(*[]Entry)
	if !ok {
		return nil
	}
	return *e
}

func fileAndLine() slog.Source {
	pc, file, line, _ := runtime.Caller(2)
	fn := runtime.FuncForPC(pc)
	var fnName string
	if fn == nil {
		fnName = "?()"
	} else {
		fnName = strings.TrimLeft(filepath.Ext(fn.Name()), ".") + "()"
	}

	return slog.Source{
		Function: fnName,
		File:     filepath.Base(file),
		Line:     line,
	}
}
