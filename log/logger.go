package log

import (
	"time"
	"net/http"
	"github.com/golang/glog"
)

/* API 请求操作日志 */
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)
		glog.Infof("%s\t%s\t%s\t%s", r.Method, r.RequestURI, name, time.Since(start))
	})
}
