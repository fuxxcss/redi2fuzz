package db

import (
	"github.com/fuxxcss/redi2fuzz/pkg/utils"
	"github.com/fuxxcss/redi2fuzz/pkg/model"
)

type DB interface {
	StartUp() error
	Restart() error
	ShutDown()
	CheckAlive() bool
	CleanUp() error
	Execute([]string) (utils.TargetState, error)
	Collect() (model.Snapshot, error)
	Stderr() string
	Debug()
}
