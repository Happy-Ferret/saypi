package log

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/metcalf/saypi/mux"
	"github.com/zenazn/goji/web/mutil"
	"golang.org/x/net/context"
)

const (
	httpDateFormat = "2006-01-02 15:04:05.000000"
	ctxKey         = "log.Extra"
)

// WrapC wraps a mux.HandlerC to log to the standard logger after the
// request completes.
func WrapC(handler mux.HandlerC) mux.HandlerC {
	return logger.WrapC(handler)
}

// SetContext adds the key-value pair to the log data stored in the
// provided context.  It overwrites existing values and returns false
// if no log data is available in the context.
func SetContext(ctx context.Context, key string, value interface{}) bool {
	extra, ok := ctx.Value(ctxKey).(map[string]interface{})
	if ok {
		extra[key] = value
	}

	return ok
}

// WrapC wraps a mux.HandlerC to log to l after the request completes.
func (l *Logger) WrapC(h mux.HandlerC) mux.HandlerC {
	// this takes the request and response, and tees off a copy of both
	// (truncated to a configurable length), and stores them in the request context
	// for later logging
	return mux.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		extra, ok := ctx.Value(ctxKey).(map[string]interface{})
		if !ok {
			extra = make(map[string]interface{})
			ctx = context.WithValue(ctx, ctxKey, extra)
		}

		w2 := mutil.WrapWriter(w)
		h.ServeHTTPC(ctx, w2, r)

		end := time.Now()

		remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)
		reqTime := float64(end.Sub(start).Nanoseconds()) / float64(time.Millisecond)

		data := map[string]interface{}{
			"time":            start.In(time.UTC).Format(httpDateFormat),
			"remote_address":  remoteAddr,
			"http_path":       r.URL.Path,
			"http_method":     r.Method,
			"http_status":     strconv.Itoa(w2.Status()),
			"bytes_written":   strconv.Itoa(w2.BytesWritten()),
			"http_user_agent": r.UserAgent(),
			"request_time":    strconv.FormatFloat(reqTime, 'f', 3, 64),
		}

		for k, v := range extra {
			data[k] = v
		}

		l.Print("http_response", "", data)
	})
}