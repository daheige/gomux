package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/daheige/gomux/app/config"
	"github.com/daheige/gomux/app/routes"
	"github.com/daheige/thinkgo/gpprof"
	"github.com/daheige/thinkgo/logger"
	"github.com/daheige/thinkgo/monitor"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port      int
	logDir    string
	configDir string
	graceWait time.Duration // 平滑重启的等待时间1s or 1m
)

func init() {
	flag.IntVar(&port, "port", 1338, "app listen port")
	flag.StringVar(&logDir, "log_dir", "./logs", "log dir")
	flag.StringVar(&configDir, "config_dir", "./", "config dir")
	flag.DurationVar(&graceWait, "graceful-timeout", 5*time.Second, "the server gracefully reload. eg: 15s or 1m")
	flag.Parse()

	// 日志文件设置
	logger.SetLogDir(logDir)
	logger.SetLogFile("go-web.log")
	logger.MaxSize(200)

	// zap 底层 const callerSkipOffset = 2
	// 这里的callerSkipOffset默认是2，所以这里InitLogger skip需要初始化为1
	logger.InitLogger(3)

	// 初始化配置文件
	config.InitConf(configDir)
	config.InitRedis()

	// 添加prometheus性能监控指标
	prometheus.MustRegister(monitor.WebRequestTotal)
	prometheus.MustRegister(monitor.WebRequestDuration)

	prometheus.MustRegister(monitor.CpuTemp)
	prometheus.MustRegister(monitor.HdFailures)

	// 性能监控的端口port+1000,只能在内网访问
	httpMux := gpprof.New()

	// 添加prometheus metrics处理器
	httpMux.Handle("/metrics", promhttp.Handler())
	gpprof.Run(httpMux, port+1000)
}

func main() {
	router := mux.NewRouter()

	// 与http.ServerMux不同的是mux.Router是完全的正则匹配
	// 设置路由路径/index/，如果访问路径/idenx/hello会返回404
	// 设置路由路径为/index/访问路径/index也是会报404的
	// 需要设置r.StrictSlash(true), /index/与/index才能匹配
	router.StrictSlash(true)

	// 健康检查
	router.HandleFunc("/check", routes.HealthCheck)

	// 路由设置
	routes.RouterHandler(router)

	// 打印访问的路由
	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}

		var queriesTemplates []string
		queriesTemplates, err = route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}

		var queriesRegexps []string
		queriesRegexps, err = route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}

		var methods []string
		methods, err = route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}

		return nil
	})

	if err != nil {
		fmt.Println("router walk error:", err)
	}

	// 设置服务端超时限制
	server := &http.Server{
		Handler: http.TimeoutHandler(router, time.Second*2, `{code:503,"message":"server timeout"}`),
		// Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 在独立携程中运行
	log.Println("server run on: ", port)

	go func() {
		defer logger.Recover()

		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Println("server listen error:", err)
				logger.Error("server listen error", map[string]interface{}{
					"trace_error": err.Error(),
				})

				return
			}

			log.Println("server will exit...")
		}
	}()

	// mux平滑重启
	ch := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// recivie signal to exit main goroutine
	// window signal
	// signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGHUP)
	// signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2, os.Interrupt, syscall.SIGHUP)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGHUP)

	// Block until we receive our signal.
	<-ch

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), graceWait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// if your application should wait for other services
	// to finalize based on context cancellation.
	go server.Shutdown(ctx) // 在独立的携程中关闭服务器
	<-ctx.Done()

	log.Println("shutting down")
	logger.Info("server shutdown success", nil)
}
