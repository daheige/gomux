package controller

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/daheige/gomux/app/helper"
)

type HomeController struct {
	BaseController
}

// Test test form id.
func (ctrl *HomeController) Test(w http.ResponseWriter, r *http.Request) {
	log.Println(r.FormValue("id"))
	log.Println(r.Form) // 所有的form get数据 //www.hgmux.com/home/test?id=1&name=daheige
	// map[name:[daheige] id:[1]] 类型是 map[string][]string
	w.Write([]byte("ok"))
}

// Post post ctrl.
func (ctrl *HomeController) Post(w http.ResponseWriter, r *http.Request) {
	log.Println(r.PostFormValue("name")) // 会自动调用r.ParseForm()解析header,body

	log.Println(r.PostForm) // 所有的form post数据
	log.Println(r.Body)
	body, err := ioutil.ReadAll(r.Body) // 读取body内容 {"ids":{"a":1}} <nil>
	log.Println(string(body), err)
	w.Write([]byte("ok"))
}

// LongAsync When starting new Goroutines inside a middleware or handler,
// you SHOULD NOT use the original request inside it,
// you have to use a read-only copy.
func (ctrl *HomeController) LongAsync(w http.ResponseWriter, r *http.Request) {
	// create copy to be used inside the goroutine
	cReq := *r // 这里cReq是一个只读request
	go func() {
		// simulate a long task with time.Sleep(). 3 seconds
		time.Sleep(3 * time.Second)

		// note that you are using the copied http.Request "cCp", IMPORTANT
		log.Println("Done! in path " + cReq.URL.Path)
	}()

	helper.ApiSuccess(w, "ok", nil)
}

// InfoReq info request.
type InfoReq struct {
	Uid   int    `json:"uid" validate:"required,min=1"`
	Limit int    `json:"limit" validate:"required,min=1,max=20"`
	Name  string `json:"name" validate:"omitempty,max=10"`
}

// Info 测试参数校验
// http://localhost:1338/get-info?limit=12&uid=1&name=abcdeabcde
func (ctrl *HomeController) Info(w http.ResponseWriter, r *http.Request) {
	// 接收参数
	req := &InfoReq{
		Uid:   ctrl.GetInt(r.FormValue("uid")),
		Limit: ctrl.GetInt(r.FormValue("limit")),
		Name:  r.FormValue("name"),
	}

	// 校验参数
	err := validate.Struct(req)
	if err != nil {
		helper.ApiError(w, 1001, "param error", nil)
		return
	}

	if req.Limit == 0 {
		helper.ApiSuccess(w, "ok", helper.EmptyArray{})
		return
	}

	helper.ApiSuccess(w, "success", helper.H{
		"uid":   req.Uid,
		"limit": req.Limit,
	})
}
