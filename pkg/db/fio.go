package db

/*
#include "afl.h"
#include "honggfuzz.h"
*/
import "C"

// global
var (
	globalFIO *FIO
	mutexFIO sync.Mutex
)

type FIO struct {
	Start func()
	Read func(C.u8,C.size_t)
	Write func(void)
}

func SingleFIO(tool string) *FIO {

	if globalFIO == nil {
		mutexFIO.Lock()
		defer mutexFIO.Unlock()
		if globalFIO == nil {
			globalFIO = NewRedi(tool)
		}
	}
	return globalFIO
}

func NewFIO(tool string) *FIO {

	fio := new(IO)

	// init IO
	switch tool {

	case utils.AFL:
		fio.Start = C.afl_forkserver_start
		fio.Read = C.afl_next_testcase
		fio.Write = C.afl_end_testcase
	}
	return fio

}

