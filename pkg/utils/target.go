package utils

// traget state
type TargetState int
const (
	STATE_OK  TargetState = iota
	STATE_ERR
	STATE_CRASH
)

// fuzz targets
type TargetType int
const (
	// Redi
	REDI_REDIS TargetType = iota
	REDI_KEYDB 
	REDI_STACK
	// TS
	TS_IOTDB
)

// target feature type
type TargetFeatureType int

const (
	TARGET_PORT TargetFeatureType = iota
	TARGET_PATH
	QUEUE_PATH
)

type TargetFeature map[TargetFeatureType]string

var Targets = map[TargetType]TargetFeature {
	// Redis
	REDI_REDIS : {
		TARGET_PORT : "6379",
		TARGET_PATH : "/usr/local/redis/src/redis-server",
		QUEUE_PATH : "queue/redis",
	},
	// KeyDB
	REDI_KEYDB : {
		TARGET_PORT : "6380",
		TARGET_PATH : "/usr/local/keydb/src/keydb-server",
		QUEUE_PATH : "queue/redis",
	},
	// RediStack
	REDI_STACK : {
		TARGET_PORT : "6381",
		TARGET_PATH : "/usr/local/redis/src/redis-stack-server",
		QUEUE_PATH : "queue/redis-stack",
	},
}





