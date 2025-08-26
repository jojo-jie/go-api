package util

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisLock Redis锁结构
type RedisLock struct {
	client        *redis.Client
	key           string
	value         string // 通常使用UUID，用于安全释放
	expiration    time.Duration
	cancelRenewal context.CancelFunc
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

// LockAndRenewal 获取锁并启动协程自动续期
func (lock *RedisLock) LockAndRenewal(ctx context.Context) (bool, error) {
	// 1. 尝试获取锁
	result, err := lock.client.SetNX(ctx, lock.key, lock.value, lock.expiration).Result()
	if err != nil || !result {
		// 2. 获取锁失败
		return result, err
	}
	// 3. 成功获取锁，启动协程进行自动续期
	renewalCtx, cancelFunc := context.WithCancel(ctx)
	lock.cancelRenewal = cancelFunc
	go lock.renewal(renewalCtx)

	return true, nil
}

func (lock *RedisLock) renewal(ctx context.Context) {
	ticker := time.NewTicker(lock.expiration / 3) // 在过期时间 1/3 时续期
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// 主上下文已结束，停止续期
			return
		case <-ticker.C:
			// 续期 lua 脚本：只有锁的持有者才能续期
			// PEXPIRE 重新设置为固定毫秒数
			renewalScript := `
			if redis.call("GET", KEYS[1]) == ARGV[1] then
				return redis.call("PEXPIRE", KEYS[1], ARGV[2]) 
			else
				return 0
			end
			`
			script := redis.NewScript(renewalScript)
			// 每次续期都将其重置为原过期时间 lock.expiration
			err := script.Run(ctx, lock.client, []string{lock.key}, lock.value, lock.expiration.Milliseconds()).Err()
			if err != nil {
				// 续期失败，记录日志并推出
				return
			}
			// fmt.Printf("续期成功\n")
		}
	}
}
