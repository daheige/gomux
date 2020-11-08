# gomux

    基于gorilla/mux构建

# 版权

    MIT，可用个人项目或商业项目

# 关于包的版本管理

    采用golang1.11版本 go mod机制（参考docs)
    对于golang1.11以下的版本，请用govendor进行包管理
    
# 关于参数校验

    采用validator库 github.com/go-playground/validator

# 设置 goproxy 代理

    go1.13以下版本：
    设置golang proxy
    vim ~/.bashrc添加如下内容：
    export GOPROXY=https://goproxy.io
    或者
    export GOPROXY=https://athens.azurefd.net

    或者
    export GOPROXY=https://mirrors.aliyun.com/goproxy/

    让bashrc生效
    source ~/.bashrc

    对于golang1.13+
    #Go version >= 1.13
    export GOPROXY=https://goproxy.io,direct
    或者
    export GOPROXY=https://goproxy.cn,direct

    让bashrc生效
    source ~/.bashrc

# 开始运行

    $ go mod tidy
    $ go run main.go
    访问localhost:1338

# 线上部署

    两种方式：
    1、采用docker
    2、采用supervior 参考 hg-mux.conf配置文件

# PProf性能监控和prometheus监控

    采用net/http/pprof包
        浏览器访问http://localhost:2338/debug/pprof，就可以查看
    在命令终端查看：
        安装graphviz
            $ apt install graphviz
        查看profile
            go tool pprof http://localhost:2338/debug/pprof/profile?seconds=60
            (pprof) top 10 --cum --sum
            (pprof) web  #web页面查看cpu使用情况

            每一列的含义：
            flat：给定函数上运行耗时
            flat%：同上的 CPU 运行耗时总比例
            sum%：给定函数累积使用 CPU 总比例
            cum：当前函数加上它之上的调用运行总耗时
            cum%：同上的 CPU 运行耗时总比例

        它会收集30s的性能profile,可以用go tool查看
            go tool pprof profile /home/heige/pprof/pprof.go-api.samples.cpu.002.pb.gz
        查看heap和goroutine
            查看活动对象的内存分配情况
            go tool pprof http://localhost:2338/debug/pprof/heap
            go tool pprof http://localhost:2338/debug/pprof/goroutine

        prometheus性能监控
        http://localhost:2338/metrics
        
    测试过程捕捉pprof
    $ curl http://localhost:2338/debug/pprof/profile?seconds=80 --out pprof.cpu
      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                     Dload  Upload   Total   Spent    Left  Speed
    100 46007    0 46007    0     0    574      0 --:--:--  0:01:20 --:--:-- 12400
    $ go tool pprof -http=:6060 pprof.cpu
    Serving web UI on http://localhost:6060
    json序列化比较耗时
    encoding/json.(*encodeState).string
    /usr/local/go/src/encoding/json/encode.go
    
    Total:       9.15s     10.47s (flat, cum) 41.93%

# wrk 工具压力测试

    https://github.com/wg/wrk

    ubuntu系统安装如下
    1、安装wrk
        # 安装 make 工具
        sudo apt-get install make git

        # 安装 gcc编译环境
        sudo apt-get install build-essential
        sudo mkdir /web/
        sudo chown -R $USER /web/
        cd /web/
        git clone https://github.com/wg/wrk.git
        # 开始编译
        cd /web/wrk
        make
    2、wrk压力测试(测试gorilla/mux 1.7.3版本)
        $ wrk -c 100 -t 8 -d 2m http://localhost:1338/index
        Running 2m test @ http://localhost:1338/index
        8 threads and 100 connections
        Thread Stats   Avg      Stdev     Max   +/- Stdev
            Latency    16.49ms   36.04ms 798.76ms   96.97%
            Req/Sec     1.06k   197.06     6.79k    72.90%
        1006361 requests in 2.00m, 123.81MB read
        Requests/sec:   8379.74
        Transfer/sec:      1.03MB
    3、metrics监控
    访问http://localhost:2338/metrics

    对api/v1/hello进行压力测试，服务端设置超时3s
    $ wrk -t 8  -c 4000 -d 2m --timeout 2 --latency http://localhost:1338/api/v1/hello
    Running 2m test @ http://localhost:1338/api/v1/hello
      8 threads and 4000 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency   666.73ms  255.09ms   2.00s    81.99%
        Req/Sec   189.97    121.79     1.01k    73.97%
      Latency Distribution
         50%  704.58ms
         75%  772.09ms
         90%  859.34ms
         99%    1.31s 
      180802 requests in 2.00m, 39.62GB read
      Socket errors: connect 2987, read 0, write 0, timeout 113
    Requests/sec:   1505.45
    Transfer/sec:    337.78MB
    
    同样的业务，用gin1.4.0进行测试
    代码： https://github.com/daheige/go-api/blob/master/app/controller/IndexController.go#L14
    $ wrk -t 8  -c 4000 -d 2m --timeout 2 --latency http://localhost:1338/v1/hello
    Running 2m test @ http://localhost:1338/v1/hello
      8 threads and 4000 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency   775.08ms  341.12ms   2.00s    77.79%
        Req/Sec   200.95    172.98     1.53k    72.91%
      Latency Distribution
         50%  834.83ms
         75%  948.03ms
         90%    1.09s 
         99%    1.60s 
      153000 requests in 2.00m, 33.55GB read
      Socket errors: connect 2987, read 0, write 0, timeout 773
    Requests/sec:   1274.01
    Transfer/sec:    286.11MB
    相比而言，mux少了一些gin render/json.go的一些处理，相对来说速度要快一些
    看了源码主要是gc,context不一样，gin是用自己的上下文，把http response,request
    进行了包装，每个请求开辟一个 gin context，响应完毕后，然后把context request,response进行reset，还有一些资源的清理，gc
    消耗的时间，相比mux用的标准上下文来说，mux更胜一筹。
    
    对比gin 和 gorilla/mux对一个不包含业务的接口进行测试
    gorilla/mux的情况
    $ wrk -t 8  -c 100 -d 1m --timeout 2 --latency http://localhost:1338/api/v1/info2
    Running 1m test @ http://localhost:1338/api/v1/info2
      8 threads and 100 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    15.70ms   14.45ms 236.26ms   80.49%
        Req/Sec     0.87k   160.88     1.67k    74.42%
      Latency Distribution
         50%   14.38ms
         75%   17.41ms
         90%   26.50ms
         99%   71.94ms
      413867 requests in 1.00m, 74.99MB read
    Requests/sec:   6889.47
    Transfer/sec:      1.25MB

    gin框架的压力测试
    $ wrk -t 8 -c 100 -d 1m --latency http://localhost:1338/api/info
       Running 1m test @ http://localhost:1338/api/info
         8 threads and 100 connections
         Thread Stats   Avg      Stdev     Max   +/- Stdev
           Latency    19.02ms   34.69ms 518.44ms   98.19%
           Req/Sec   799.69    139.57     1.53k    75.05%
         Latency Distribution
            50%   15.86ms
            75%   18.55ms
            90%   24.50ms
            99%  175.30ms
         377484 requests in 1.00m, 67.32MB read
       Requests/sec:   6283.54
       Transfer/sec:      1.12MB
    发现gin响应时间要比gorilla/mux要长，而且qps要明显低一些
    
# 参考文档

    https://github.com/gorilla/mux#middleware
    https://github.com/gorilla/mux#examples
