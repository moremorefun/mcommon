package mcommon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

// baseKey 基础key
var baseKey = ""

// RedisCreate 创建数据库
func RedisCreate(address string, password string, dbIndex int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       dbIndex,  // use default DB
	})
	_, err := client.Ping().Result()
	if err != nil {
		Log.Fatalf("redis ping error: %s", err.Error())
		return nil
	}
	return client
}

// RedisSetBaseKey 设置基础key
func RedisSetBaseKey(v string) {
	baseKey = v
}

// RedisGet 获取
func RedisGet(ctx context.Context, client *redis.Client, key string) (string, error) {
	key = fmt.Sprintf("%s_%s", baseKey, key)
	ret, err := client.WithContext(ctx).Get(key).Result()
	if err != nil {
		// "redis: nil" 不存在
		if !strings.Contains(err.Error(), "redis: nil") {
			return "", err
		}
		return "", nil
	}
	return ret, nil
}

// RedisSet 设置
func RedisSet(ctx context.Context, client *redis.Client, key, value string, du time.Duration) {
	key = fmt.Sprintf("%s_%s", baseKey, key)
	err := client.WithContext(ctx).Set(key, value, du).Err()
	if err != nil {
		return
	}
}
