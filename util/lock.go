package util

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisLock Redis锁结构
type RedisLock struct {
	client     *redis.Client
	key        string
	value      string // 通常使用UUID，用于安全释放
	expiration time.Duration
}

// NewRedisLock 创建一个锁实例
func NewRedisLock(client *redis.Client, key string, expiration time.Duration) *RedisLock {
	return &RedisLock{
		client:     client,
		key:        key,
		value:      uuid.New().String(),
		expiration: expiration,
	}
}

func (lock *RedisLock) Acquire(ctx context.Context) (bool, error) {
	result, err := lock.client.SetNX(ctx, lock.key, lock.value, lock.expiration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

func (lock *RedisLock) Release(ctx context.Context) (bool, error) {
	// 1. Lua脚本：如果Redis中存储的值等于传入的 value，则执行删除操作
	cadScript := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`
	script := redis.NewScript(cadScript)
	// 2. 执行Lua脚本，传入锁的 key 和唯一的 value
	result, err := script.Run(ctx, lock.client, []string{lock.key}, lock.value).Int64()
	if err != nil {
		return false, err
	}
	// 如果返回结果 > 0，代表删除成功（释放了锁）
	return result > 0, nil
}
