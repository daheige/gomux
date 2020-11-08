# mux介绍

    mux 是一个用来执行http请求的路由和分发的第三方扩展包。
    mux 其名称来源于HTTP request multiplexer，类似于官方包http.ServeMux，mux.Router将会定义一个路由列表，其中每一个路由都会定义对应的请求url，及其处理方法。

# 源码

    第三方库源码：
    https://github.com/gorilla/mux

# 安装

    go get -u github.com/gorilla/mux
    
# 使用

    添加包引用：
    "github.com/gorilla/mux"
    
# 常用方法介绍
    初始化路由
    r := mux.NewRouter()
# 路由注册
    最简单的路由注册：

    r.HandleFunc("/", HomeHandler)
    其中代码中的第一个参数为请求url，第二个参数为请求的处理函数，该函数可简单的定义为以下代码：

    func HomeHandler(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "this is home")
    }
# 带有变量的url路由注册：
    其中参数可使用正则表达式匹配

    r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler)
# 指定Host：
    r.Host("www.example.com")
# 指定http方法：
    r.Methods("GET", "POST")
# 指定URL安全策略：
    r.Schemes("https")
# 增加URL前缀：
    r.PathPrefix("/products/")
# 添加请求头：
    r.Headers("X-Requested-With", "XMLHttpRequest")
# 添加请求参数：
    r.Queries("key", "value")
# 组合使用：

    r.HandleFunc("/products", ProductsHandler).
    Host("www.example.com").
    Methods("GET").
    Schemes("http")
# 子路由的使用：
    r := mux.NewRouter()
    s := r.PathPrefix("/products").Subrouter()
    // "/products/"
    s.HandleFunc("/", ProductsHandler)
    // "/products/{key}/"
    s.HandleFunc("/{key}/", ProductHandler)
    // "/products/{key}/details"
    s.HandleFunc("/{key}/details", ProductDetailsHandler)
# 定义路由别名：

    r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler).Name("article")
# 静态文件路由：

    flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
    flag.Parse()
    r := mux.NewRouter()

    // This will serve files under http://localhost:8000/static/<filename>
        r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
# 生成已注册的URL:
    生成已注册的url需要用到路由的别名，代码如下：

    url, err := r.Get("router_name").URL("key1", "val1", "key2", "val2")
    例如：

    r := mux.NewRouter()
    r.Host("{subdomain}.domain.com").
    Path("/articles/{category}/{id:[0-9]+}").
    Queries("filter", "{filter}").
    HandlerFunc(ArticleHandler).
    Name("article")

    // url.String() will be "http://news.domain.com/articles/technology/42?filter=gorilla"
    url, err := r.Get("article").URL("subdomain", "news",
                                    "category", "technology",
                                    "id", "42",
                                    "filter", "gorilla")
# Walk方法：
    walk方法可以遍历访问所有已注册的路由，例如以下代码：

    func main() {
        r := mux.NewRouter()
        r.HandleFunc("/", handler)
        r.HandleFunc("/products", handler).Methods("POST")
        r.HandleFunc("/articles", handler).Methods("GET")
        r.HandleFunc("/articles/{id}", handler).Methods("GET", "PUT")
        r.HandleFunc("/authors", handler).Queries("surname", "{surname}")
        err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
            pathTemplate, err := route.GetPathTemplate()
            if err == nil {
                fmt.Println("ROUTE:", pathTemplate)
            }
            pathRegexp, err := route.GetPathRegexp()
            if err == nil {
                fmt.Println("Path regexp:", pathRegexp)
            }
            queriesTemplates, err := route.GetQueriesTemplates()
            if err == nil {
                fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
            }
            queriesRegexps, err := route.GetQueriesRegexp()
            if err == nil {
                fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
            }
            methods, err := route.GetMethods()
            if err == nil {
                fmt.Println("Methods:", strings.Join(methods, ","))
            }
            fmt.Println()
            return nil
        })

        if err != nil {
            fmt.Println(err)
        }

        http.Handle("/", r)
    }
# Middleware 中间件
    mux同样也支持为路由添加中间件。
    最简单的中间件定义：

    func loggingMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Do stuff here
            log.Println(r.RequestURI)
            // Call the next handler, which can be another middleware in the chain, or the final handler.
            next.ServeHTTP(w, r)
        })
    }
# 中间件使用：

    r := mux.NewRouter()
    r.HandleFunc("/", handler)
    r.Use(loggingMiddleware)
# 综合示例
```
package th_mux

import(
    "strings"
	"flag"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
)

func Run(){
	var dir string
    flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
    flag.Parse()
	// 初始化Router
	r := mux.NewRouter()
	// 静态文件路由
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))
	// 普通路由
	r.HandleFunc("/", HomeHandler)
	// 指定host
    r.HandleFunc("/host", HostHandler).Host("localhost")
	// 带变量的url路由
	r.HandleFunc("/users/{id}", GetUserHandler).Methods("Get").Name("user")

	url, _ := r.Get("user").URL("id", "5")
	fmt.Println("user url: ", url.String())

	// 遍历已注册的路由
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})

    r.Use(TestMiddleware)
	http.ListenAndServe(":3000", r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "this is home")
}

func HostHandler(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "the host is localhost")
}

func GetUserHandler(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	
	fmt.Fprint(w, "this is get user, and the user id is ", vars["id"])
}


func TestMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Do stuff here
        fmt.Println("middleware print: ", r.RequestURI)
        // Call the next handler, which can be another middleware in the chain, or the final handler.
        next.ServeHTTP(w, r)
    })
}
```
