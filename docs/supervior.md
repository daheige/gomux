# supervisor使用
    注意:以下代码,命令和配置在ubuntu18.04LTS上实际测试,其他发行版或ubuntu其他版本仅供参考
#安装 
    安装可以使用一下命令:
        sudo apt-get install supervisor

        安装成功后,supervisor就会默认启动

# 使用 
        有很多方法添加进程,看了很多博客上的介绍和加上我的实际使用,我认为一下方法最好用:

        将每个进程的配置文件单独拆分,放在/etc/supervisor/conf.d/目录下,以.conf作为扩展名,例如test.conf定义的一个简单的HTTP服务器:
# 配置demo
[program:test]
    command=python -m SimpleHTTPServer
    重启supervisor,让配置文件生效,然后启动test进程:

    supervisorctl reload
    supervisorctl start test

    如果要停止进程,就用stop 
# 其他一些配置,通过这个例子讲解
[program:hg-mux]
    command=/data/www/hg-mux/hg-mux -log_dir=/data/www/hg-mux/logs -port=8080 > /dev/null 2>&1
    numprocs=1                    ; 启动几个进程
    directory=/cas/bin                ; 执行前要不要先cd到目录去，一般不用
    autostart=true                ; 随着supervisord的启动而启动
    autorestart=true              ; 自动重启。。当然要选上了
    startretries=10               ; 启动失败时的最多重试次数
    startsecs = 5                 ; 启动 5 秒后没有异常退出，就当作正常启动了
    exitcodes=0                 ; 
    正常退出代码（是说退出代码是这个时就不再重启了吗？待确定）
    stopsignal=KILL               ; 用来杀死进程的信号
    stopwaitsecs=10               ; 发送SIGKILL前的等待时间
    redirect_stderr=true          ; 重定向stderr到stdout
    stdout_logfile=logfile        ; 指定日志文件

# 常用命令: 
    supervisorctl start programxxx，启动某个进程

    supervisorctl restart programxxx，重启某个进程

    supervisorctl stop groupworker: ，重启所有属于名为groupworker这个分组的进程(start,restart同理)

    supervisorctl stop all，停止全部进程，注：start、restart、stop都不会载入最新的配置文件。

    supervisorctl reload
    载入最新的配置文件，停止原有进程并按新的配置启动、管理所有进程。

    supervisorctl update
    根据最新的配置文件，启动新配置或有改动的进程，配置没有改动的进程不会受影响而重启。

    supervisor启动和停止的日志文件存放在/var/log/supervisor/supervisord.log

    注意：显式用stop停止掉的进程，用reload或者update都不会自动重启
    
