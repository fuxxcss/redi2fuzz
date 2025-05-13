package db

import (
	"sync"

	"github.com/fuxxcss/redi2fuxx/pkg/utils"
)

// global
var (
	globalFIO *FIO
	mutexFIO  sync.Mutex
)

// fio
type FIO struct {
	Start func()
	Read  func([]byte) int
	Write func()
}

func SingleFIO(tool string) *FIO {

	if globalFIO == nil {
		mutexFIO.Lock()
		defer mutexFIO.Unlock()
		if globalFIO == nil {
			globalFIO = NewFIO(tool)
		}
	}
	return globalFIO
}

func NewFIO(tool string) *FIO {

	fio := new(FIO)

	// init FIO
	switch tool {

	case utils.AFL:
		fio.Start = afl_forkserver_start
		fio.Read = afl_next_testcase
		fio.Write = afl_end_testcase
	}

	return fio

}
