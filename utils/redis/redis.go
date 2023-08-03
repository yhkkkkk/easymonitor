package redis

import (
	"easymonitor/conf"
	"fmt"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
)

var Client *redisv8.Client

func Setup() {
	conn := conf.AppConf.Redis
	redisDb := redisv8.NewClient(&redisv8.Options{
		Addr:         fmt.Sprintf("%s:%d", conn.Addr, conn.Port),
		Password:     conn.Password,
		DB:           conn.Db,
		MaxRetries:   conn.MaxRetries,
		ReadTimeout:  time.Second * time.Duration(conn.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(conn.WriteTimeout),
		PoolSize:     conn.PoolSize,
		MinIdleConns: conn.MinIdleConns,
		DialTimeout:  time.Second * time.Duration(conn.DialTimeout),
	})
	Client = redisDb
}
