package db

/*
#include <afl/config.h>
*/
import "C"

import (
	"encoding/binary"
	"log"
	"os"
)

var (
	aflControl *os.File = os.NewFile(C.FORKSRV_FD, "afl control")
	aflWrite   *os.File = os.NewFile(C.FORKSRV_FD+1, "afl write")
)

func afl_forkserver_start() {

	var phone uint32 = 0

	// Phone home and tell the parent that we're OK.
	binary.Write(aflWrite, binary.LittleEndian, phone)

}

func afl_next_testcase(buf []byte) int {

	var phone uint32 = 0xffffff
	recv := make([]byte, 4)

	// Wait for parent by reading from the pipe.
	aflControl.Read(recv)

	// we have a testcase - read it
	size, err := os.Stdin.Read(buf)

	if err != nil {
		log.Println("here", err)
		return 0
	}

	// report that we are starting the target
	binary.Write(aflWrite, binary.LittleEndian, phone)

	return size

}

func afl_end_testcase() {

	var phone uint32 = 0xffffff

	// next one
	binary.Write(aflWrite, binary.LittleEndian, phone)

}
