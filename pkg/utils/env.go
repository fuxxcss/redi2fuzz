package utils

// traget state
const (
	STATE_LEN int = 3
	STATE_OK  string = "okk"
	STATE_BAD string = "bad"
	STATE_ERR string = "err"
)

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
	QUEUE_PATH
)

type TargetsType map[TargetFeature]string

var Targets = map[string]TargetsType {
	// Redis
	Redis : {
		TARGET_PORT : "6379",
		TARGET_PATH : "/usr/local/redis/src/redis-server",
		QUEUE_PATH : "queue/redis",
	},
	// KeyDB
	KeyDB : {
		TARGET_PORT : "6380",
		TARGET_PATH : "/usr/local/keydb/src/keydb-server",
		QUEUE_PATH : "queue/redis",
	},
	// RediStack
	RediStack : {
		TARGET_PORT : "6381",
		TARGET_PATH : "/usr/local/redis/src/redis-stack-server",
		QUEUE_PATH : "queue/redis-stack",
	},
}



