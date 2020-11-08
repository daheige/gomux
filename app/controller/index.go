package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/daheige/gomux/app/config"
	"github.com/daheige/gomux/app/helper"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

// IndexController index ctrl.
type IndexController struct{}

// Home home page.
func (ctrl *IndexController) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello hg-mux"))
}

// ArtBook art book.
func (ctrl *IndexController) ArtBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	log.Println(vars["id"])
	fmt.Fprintf(w, "Category: %v\n", vars["category"])
}

func (ctrl *IndexController) ArtName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	log.Println(vars["name"])
	fmt.Fprintf(w, "Category: %v\n", vars["category"])
}

// Category url restful param
func (ctrl *IndexController) Category(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // 从url上获取的{category}变量放在一个map[string]string中
	log.Println("category: ", vars["category"])

	// common.InfoLog(1111)
	// log.Println(r.Context())
	// log.Println(r.RemoteAddr)
	log.Println(r.Header.Get("Remote_addr"))
	w.Write([]byte("hello world," + vars["category"]))
}

// Test test redis.
func (ctrl *IndexController) Test(w http.ResponseWriter, r *http.Request) {
	// log.Println("log_id: ", r.Context().Value("log_id"))
	conn, err := config.GetRedisObj("default")
	log.Println("err: ", err)
	if err != nil {
		helper.ApiError(w, 1001, "redis connection error", nil)
		return
	}

	defer conn.Close()

	v, err := redis.String(conn.Do("get", "myname"))
	log.Println(v, err)
	helper.ApiSuccess(w, "ok: "+v, nil)
}

// Info method.
func (ctrl *IndexController) Info(w http.ResponseWriter, r *http.Request) {
	go ctrl.task(r.Context())

	time.Sleep(time.Second * 10)

	helper.ApiSuccess(w, "hello world", nil)
}

func (ctrl *IndexController) task(ctx context.Context) {
	ch := make(chan struct{}, 1)
	go func() {
		// 模拟4秒耗时任务
		time.Sleep(time.Second * 4)
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		log.Println("done")
	case <-ctx.Done():
		log.Println("timeout")
	}
}

// MockPanic 模拟发生panic操作
func (ctrl *IndexController) MockPanic(w http.ResponseWriter, r *http.Request) {
	panic(111)
}

// Hello 压测
func (ctrl *IndexController) Hello(w http.ResponseWriter, r *http.Request) {
	// log.Println(111)
	userInfo := getUserInfo()

	// 模拟赋值操作
	info := make([]UserInfo, 0, len(userInfo))
	for k, _ := range userInfo {
		info = append(info, UserInfo{
			Id:      userInfo[k].Id,
			Name:    userInfo[k].Name,
			Age:     userInfo[k].Age,
			Content: userInfo[k].Content,
		})
	}

	// 对map进行主动gc
	userInfo = nil

	b, _ := json.Marshal(info)

	w.Write(b)
}

type UserInfo struct {
	Id      int64
	Name    string
	Age     int
	Content string
}

func getUserInfo() map[string]UserInfo {
	user := make(map[string]UserInfo, 500)
	str := `What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
	What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
	What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
	What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
What to profile is controlled by config value passed to profile.Start. By default CPU profiling is enabled.
`

	for i := 0; i < 100; i++ {
		nick := "hello_" + strconv.Itoa(i)
		user[nick] = UserInfo{
			Id:      int64(i),
			Name:    nick,
			Age:     i + 10,
			Content: str,
		}
	}

	return user
}
