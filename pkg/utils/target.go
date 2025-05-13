package utils

// traget state
const (
	STATE_LEN int = 3
	STATE_OK  string = "okk"
	STATE_BAD string = "bad"
	STATE_ERR string = "err"
)

// testcase maxsize
const (
	MaxSize int = 0x100000
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
)

type TargetsType map[TargetFeature]string

var Targets = map[string]TargetsType {
	// Redis
	Redis : TargetsType {
		TARGET_PORT : "6379",
		TARGET_PATH : "/usr/local/redis/src/redis-server",
	},
	// KeyDB
	KeyDB : TargetsType {
		TARGET_PORT : "6380",
		TARGET_PATH : "/usr/local/keydb/src/keydb-server",
	},
	// RediStack
	RediStack : TargetsType {
		TARGET_PORT : "6381",
		TARGET_PATH : "/usr/local/redis/src/redis-stack-server",
	},
}



