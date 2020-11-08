package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/daheige/gomux/app/controller"
	"github.com/daheige/gomux/app/middleware"

	"github.com/daheige/thinkgo/monitor"

	"github.com/gorilla/mux"
)

// RouterHandler router list
func RouterHandler(router *mux.Router) {

	reqWare := &middleware.RequestWare{}

	// 全局中间件 记录日志和异常捕获处理,ip地址记录
	router.Use(middleware.RealIP, reqWare.AccessLog, reqWare.Recover)

	// 服务端超时中间件，可以根据实际情况添加这个中间件
	router.Use(middleware.TimeoutHandler(3 * time.Second))

	// prometheus性能监控打点
	router.Use(monitor.MonitorHandler)

	// 404处理
	router.NotFoundHandler = http.HandlerFunc(reqWare.NotFoundHandler)

	indexCtrl := &controller.IndexController{}
	router.HandleFunc("/", indexCtrl.Home)

	// 对单个接口做性能监控打点
	// router.HandleFunc("/index", monitor.MonitorHandlerFunc(indexCtrl.Home)).Methods("GET")

	router.HandleFunc("/index", indexCtrl.Home).Methods("GET")

	router.HandleFunc("/category", indexCtrl.Category)
	router.HandleFunc("/test", indexCtrl.Test)
	router.HandleFunc("/info", indexCtrl.Info)
	router.HandleFunc("/api/{category}", indexCtrl.Category)
	router.HandleFunc("/book/{category}/{id:[0-9]+}", indexCtrl.ArtBook) // 占位符和正则相结合的路由
	router.HandleFunc("/book/{category}/{name:\\w+}", indexCtrl.ArtName)

	// 模拟panic操作
	// http://localhost:8080/mock-panic
	router.HandleFunc("/mock-panic", indexCtrl.MockPanic)

	// 测试get/post接收数据
	homeCtrl := &controller.HomeController{}
	router.HandleFunc("/home", homeCtrl.Test).Methods("GET")
	router.HandleFunc("/home/test", homeCtrl.Test).Methods("GET")
	router.HandleFunc("/home/post", homeCtrl.Post).Methods("POST")

	router.HandleFunc("/long-async", homeCtrl.LongAsync)
	router.HandleFunc("/get-info", homeCtrl.Info)

	// 子路由分组
	v1 := router.PathPrefix("/api/v1").Subrouter()

	// http://localhost:1338/api/v1/get-info?uid=1&name=daheige&limit=12
	v1.HandleFunc("/get-info", homeCtrl.Info).Methods("GET")
	v1.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// 压力测试
	v1.HandleFunc("/hello", indexCtrl.Hello)

	v1.HandleFunc("/info", indexCtrl.Info)

	// 单独对接口进行设置超时
	v1.HandleFunc("/info2", indexCtrl.Info)

	v1.HandleFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		log.Println(111)
		time.Sleep(3 * time.Second)

		// w.Header().Set("connection", "close")

		flushBody(w) // 立即flush并断开连接

		log.Println("server has finish")
	})

	router.HandleFunc("/polling", func(w http.ResponseWriter, r *http.Request) {
		log.Println("polling")
		time.Sleep(10 * time.Second)

		select {
		case <-w.(http.CloseNotifier).CloseNotify(): // 连接已经关闭
			log.Println("connection closed")
		// case <-time.After(5 * time.Second):
		//     w.Write([]byte("server timeout"))
		default:
			log.Println("send client response")
			w.Write([]byte("I am OK"))
		}
	})

}

/**
参考: http://bewithyou.me/archive/detail/76

http/server.go 源码
func (c *conn) serve(ctx context.Context)方法中
if !c.hijacked() {
	c.close()
	c.setState(c.rwc, StateClosed)
}
链接的断开是交给底层HTTP库来操作的
*/

func flushBody(w http.ResponseWriter) bool {
	// if f, ok := w.(http.Flusher); ok {
	// f.Flush()
	if hj, ok := w.(http.Hijacker); ok { // 从ResponseWriter获取链接控制权
		if conn, _, err := hj.Hijack(); err == nil {
			if err := conn.Close(); err == nil {

				log.Println("client has close,server will end this request")
				return true
			}
		}
	}
	// }

	return false
}

// HealthCheck A very simple health check.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	w.Write([]byte(`{"alive": true}`))
}
