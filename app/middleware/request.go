package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/daheige/gomux/app/extensions/logger"
	"github.com/daheige/gomux/app/helper"

	"github.com/daheige/thinkgo/grecover"
	"github.com/daheige/thinkgo/gutils"
)

// RequestWare request middleware.
type RequestWare struct{}

// AccessLog 访问日志
func (reqWare *RequestWare) AccessLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		// wrk测试发现log.Println会枷锁，将内容输出到终端的时候，每次都会sync.Mutex
		// lock,unlock操作
		// log.Println("request before")
		// log.Println("request uri: ", r.RequestURI)

		// 设置一些请求的公共参数到上下文上
		reqId := r.Header.Get("x-request-id")
		if reqId == "" {
			reqId = gutils.RndUuidMd5()
		}

		// log.Println("log_id: ", reqId)
		// 将requestId 写入当前上下文中
		r = helper.ContextSet(r, "log_id", reqId)

		// 通过ClientIpWare之后，这里的r.RemoteAddr就是客户端的ip真实地址
		r = helper.ContextSet(r, "client_ip", r.RemoteAddr)

		r = helper.ContextSet(r, "request_method", r.Method)
		r = helper.ContextSet(r, "request_uri", r.RequestURI)
		r = helper.ContextSet(r, "user_agent", r.Header.Get("User-Agent"))

		logger.Info(r.Context(), "exec begin", nil)

		h.ServeHTTP(w, r)

		// log.Println("request end")
		// 请求结束记录日志
		logger.Info(r.Context(), "exec end", map[string]interface{}{
			"exec_time": time.Now().Sub(t).Seconds(),
		})

	})
}

// Recover 当请求发生了异常或致命错误，需要捕捉r,w执行上下文的错误
// 该Recover设计灵感来源于golang gin框架的WriteHeaderNow()设计
func (reqWare *RequestWare) Recover(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(r.Context(), "exec panic error", map[string]interface{}{
					"trace_error": string(grecover.CatchStack()),
				})

				// 当http请求发生了recover或异常就直接终止
				helper.HttpCode(w, http.StatusInternalServerError, "server error!")
				return
			}
		}()

		h.ServeHTTP(w, r)
	})
}

// NotFoundHandler 404处理函数
func (reqWare *RequestWare) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	helper.HttpCode(w, http.StatusNotFound, "404 - page not found")
}

// TimeoutHandler server timeout middleware wraps the request context with a timeout
// 中间件参考go-chi/chi库 https://github.com/go-chi/chi/blob/master/middleware/timeout.go
// return http.Handler处理器，服务端超时中间件，可以根据实际情况添加这个中间件
func TimeoutHandler(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// wrap the request context with a timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer func() {
				// cancel to clear resources after finished
				cancel()

				// check if context timeout was reached
				if ctx.Err() == context.DeadlineExceeded {

					// 记录操作日志
					logger.Error(r.Context(), "server timeout", nil)

					// write response and abort the request
					w.WriteHeader(http.StatusGatewayTimeout)
					w.Write([]byte(`{code:504,"message":"gateway timeout"}`))

					return
				}

			}()

			// 继续往后执行
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}

}

// TimeoutHandlerFunc 单独对处理器函数做超时限制
// return http.HandlerFunc，服务端超时中间件，可以根据实际情况添加这个中间件
func TimeoutHandlerFunc(timeout time.Duration, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer func() {
			// cancel to clear resources after finished
			cancel()

			// check if context timeout was reached
			if ctx.Err() == context.DeadlineExceeded {

				// 记录操作日志
				logger.Error(r.Context(), "server timeout", nil)

				// write response and abort the request
				w.WriteHeader(http.StatusGatewayTimeout)
				w.Write([]byte(`{code:504,"message":"gateway timeout"}`))

				return
			}

		}()

		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}

}
