package utils

// fuxx targets
const (
	Redis string = "redis"
	KeyDB string = "keydb"
	RediStack string = "redis-stack"
)

// target features
type TargetFeature int

const (
	TARGET_PORT TargetFeature = iota
	TARGET_PATH
)

var Targets = map[string]interface{} {
	// Redis
	Redis : map[TargetFeature]interface{} {
		TARGET_PORT : "6379",
		TARGET_PATH : "/usr/local/redis/src/redis-server",
	},
	// KeyDB
	KeyDB : map[TargetFeature]interface{} {
		TARGET_PORT : "6380",
		TARGET_PATH : "/usr/local/keydb/src/keydb-server",
	},
	// RediStack
	RediStack : map[TargetFeature]interface{} {
		TARGET_PORT : "6381",
		TARGET_PATH : "/usr/local/redis/src/redis-stack-server",
	},
}



