package cache

import (
	"os"
	"go-api/util"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	localcache "github.com/patrickmn/go-cache"
)

// RedisClient Redis缓存客户端单例
var RedisClient *redis.Client

// Redis 在中间件中初始化redis链接
func Redis() {
	db, _ := strconv.ParseUint(os.Getenv("REDIS_DB"), 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PW"),
		DB:       int(db),
	})

	_, err := client.Ping().Result()

	if err != nil {
		util.Log().Panic("连接Redis不成功", err)
	}

	RedisClient = client
}

var LocalCacheClient *localcache.Cache

func LocalCache() {
	LocalCacheClient =  localcache.New(5*time.Minute, 10*time.Minute)
}
