package config

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/daheige/thinkgo/gredigo"
	"github.com/daheige/thinkgo/yamlconf"
	"github.com/gomodule/redigo/redis"
)

var (
	// AppEnv app_env
	AppEnv string
	conf   *yamlconf.ConfigEngine
)

// InitConf init config.
func InitConf(path string) {
	dir, err := filepath.Abs(path)
	if err != nil {
		log.Fatalln("config dir path error: ", err)
	}

	conf = yamlconf.NewConf()
	conf.LoadConf(filepath.Join(dir, "app.yaml"))
}

// InitRedis 初始化redis
func InitRedis() {
	redisConf := &gredigo.RedisConf{}
	conf.GetStruct("RedisCommon", redisConf)

	// log.Println(redisConf)
	redisConf.SetRedisPool("default")
}

// GetRedisObj 从连接池中获取redis client
func GetRedisObj(name string) (redis.Conn, error) {
	conn := gredigo.GetRedisClient(name)
	if conn == nil || conn.Err() != nil {
		return nil, errors.New("get redis client error")
	}

	return conn, nil
}
