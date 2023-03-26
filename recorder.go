package otelfox

import (
	"github.com/felixge/httpsnoop"
	"net/http"
	"sync"
)

var rwrPool = &sync.Pool{
	New: func() interface{} {
		return &responseWriterRecorder{}
	},
}

func newResponseStatusRecorder(w http.ResponseWriter) *responseWriterRecorder {
	rwr := rwrPool.Get().(*responseWriterRecorder)
	rwr.status = http.StatusOK
	rwr.written = false
	rwr.w = httpsnoop.Wrap(w, httpsnoop.Hooks{
		Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(b []byte) (int, error) {
				if !rwr.written {
					rwr.written = true
				}
				return next(b)
			}
		},
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				if !rwr.written {
					rwr.written = true
					rwr.status = statusCode
				}
				next(statusCode)
			}
		},
	})

	return rwr
}

type responseWriterRecorder struct {
	w       http.ResponseWriter
	status  int
	written bool
}

func (rwr *responseWriterRecorder) free() {
	rwr.w = nil
	rwrPool.Put(rwr)
}
