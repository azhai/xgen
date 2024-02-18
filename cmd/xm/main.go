package main

import (
	"context"
	"strconv"

	"github.com/azhai/xgen/utils/cache"
	"github.com/muyo/sno"
)

func main() {
	ctx := context.Background()
	counter := cache.NewRedisString(ctx, "counter", 0)
	no := counter.Incr(1)
	tasks := cache.NewRedisHash(ctx, strconv.Itoa(no), "tasks", 0)
	tasks.Merge(cache.Dict{"id": "Echo" + sno.New(0).String()})
	// mq := cache.NewRedisMQ(ctx, "job", "jobGroup")
}
