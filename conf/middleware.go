package conf

import (
	"easymonitor/infra"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"sync/atomic"
	"time"

	"github.com/juju/ratelimit"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var (
	visitors = sync.Map{}
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

type visitorData struct {
	requests int64
	lastSeen time.Time
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Status() int {
	return rw.statusCode
}

type Middleware func(http.Handler) http.Handler

func MetricMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w}
		// 在处理请求之前执行的逻辑
		// 可以在这里进行请求验证、日志记录等操作
		reqBody, err := httputil.DumpRequest(r, true) // 解析读取请求内容
		if err != nil {
			infra.Logger.WithError(err).Errorf("dump request fail")
		}
		// 调用下一个处理程序
		// MetricMonitor.RecordServerCount(TypeHTTP, r.Method, r.URL.Path)
		next.ServeHTTP(rw, r)
		// MetricMonitor.RecordServerHandlerSeconds(TypeHTTP, r.Method, rw.Status(), r.URL.Path, time.Now().Sub(now).Seconds())
		if rw.Status() != http.StatusOK {
			infra.Logger.WithFields(log.Fields{
				"request": string(reqBody),
				"code":    rw.statusCode,
			}).Warn("request fail ")
		}
	})
}

func LimitHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		visitor, loaded := visitors.LoadOrStore(ip, &visitorData{requests: 1, lastSeen: time.Now()})
		if loaded { // 如果已经存在，增加请求数
			v := visitor.(*visitorData)
			atomic.AddInt64(&v.requests, 1)
			v.lastSeen = time.Now()
		}

		// 每隔一分钟重置请求计数
		go func() {
			time.Sleep(1 * time.Minute)
			v := visitor.(*visitorData)
			atomic.StoreInt64(&v.requests, 0)
		}()

		v := visitor.(*visitorData)
		if atomic.LoadInt64(&v.requests) > int64(AppConf.Server.LimitSize) {
			infra.Logger.WithFields(log.Fields{
				"ip":      ip,
				"request": r.URL.Path,
			}).Warn("Rate limit exceeded")
			http.Error(w, fmt.Sprintf("Ip %s Rate limit exceeded", ip), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func LimitHandler2(next http.HandlerFunc) http.HandlerFunc {
	limiter := rate.NewLimiter(100, 100)

	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "系统繁忙 请稍后重试", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func LimitHandler3(next http.HandlerFunc, fillInterval time.Duration, cap, quantum int64) http.HandlerFunc {
	bucket := ratelimit.NewBucketWithQuantum(fillInterval, cap, quantum)

	return func(w http.ResponseWriter, r *http.Request) {
		if bucket.TakeAvailable(1) < 1 {
			http.Error(w, "系统繁忙 请稍后重试", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}
