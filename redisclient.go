package redisdb

import (
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
)

var RdsSources cmap.ConcurrentMap[string, *redis.Client] = cmap.New[*redis.Client]()
