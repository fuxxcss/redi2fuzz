package db

/*
#include "afl.h"
#include "honggfuzz.h"
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/fuxxcss/redi2fuxx/pkg/utils"
)

// global
var (
	globalFIO *FIO
	mutexFIO sync.Mutex
)

// fio type
type Cuint8 C.uint8_t
type Csize C.size_t

type FIO struct {
	Start func()
	Read func(*Cuint8,Csize) Csize
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

	var cStart,cRead,cWrite unsafe.Pointer

	// init FIO
	switch tool {

	case utils.AFL:
		cStart = C.afl_forkserver_start
		cRead = C.afl_next_testcase
		cWrite = C.afl_end_testcase
	}

	fio.Start = *(*func())(unsafe.Pointer(&cStart))
	fio.Read = *(*func(*Cuint8,Csize) Csize)(unsafe.Pointer(&cRead))
	fio.Write = *(*func())(unsafe.Pointer(&cWrite))

	return fio

}

