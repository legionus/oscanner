package logger

import (
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/distribution/context"

	ocontext "github.com/openshift/oscanner/pkg/context"
)

func NewEntry() *Entry {
	return &Entry{
		Entry: log.NewEntry(log.StandardLogger()),
	}
}

func GetHTTPEntry(ctx context.Context) *Entry {
	e := NewEntry()

	if v := ctx.Value("instance.id"); v != nil {
		e = e.WithField(ocontext.HTTPRequestIDKey, v)
	}

	for _, k := range []string{
		ocontext.HTTPRequestMethodKey,
		ocontext.HTTPRequestURLKey,
		ocontext.HTTPRequestRAddrKey,
		ocontext.HTTPRequestLengthKey,
		ocontext.HTTPRequestTimeKey,
	} {
		if v := ctx.Value(k); v != nil {
			e = e.WithField(k, v)
		}
	}
	return e
}

type Entry struct {
	*log.Entry
}

func (e *Entry) WithFieldsDepth(args log.Fields, depth int) *Entry {
	e.Entry = e.Entry.WithFields(args)

	if file, line, ok := GetFileLine(depth); ok {
		e.Entry = e.Entry.WithField("file", file)
		e.Entry = e.Entry.WithField("fileline", line)
	}
	return e
}

func (e *Entry) WithFields(args log.Fields) *Entry {
	return e.WithFieldsDepth(args, 3)
}

func (e *Entry) WithField(key string, value interface{}) *Entry {
	e.Entry = e.Entry.WithField(key, value)
	return e
}

func GetFileLine(depth int) (file string, line int, ok bool) {
	_, file, line, ok = runtime.Caller(depth)

	if ok {
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}
	} else {
		file = "???"
		line = 0
	}
	return
}
