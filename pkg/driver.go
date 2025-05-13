package main

/*
#include <stdint.h>
*/
import "C"

import (
	"log"
	"os"

	"github.com/fuxxcss/redi2fuxx/pkg/db"
	"github.com/fuxxcss/redi2fuxx/pkg/utils"
)

// Fuxxer Server File
const (
	FDRIVER_R uintptr = iota + 3
	FDRIVER_W
)

func main() {

	// pipes
	pipeR := os.NewFile(FDRIVER_R, "Read")
	pipeW := os.NewFile(FDRIVER_W, "Write")

	// get tool, port
	tool := os.Getenv(utils.BaseTool)

	// init FIO
	fio := db.SingleFIO(tool)

	// start forkserver
	fio.Start()

	// avoid: out of memory
	buf := make([]byte, utils.MaxSize)

	for {

		// get one testcase
		size := fio.Read(buf)

		if size <= 0 {
			log.Fatalln("next testcase failed")
		}

		testcase := string(buf[:size])

		// phone fuxxer server
		pipeW.WriteString(testcase)


		// read answer
		recv := make([]byte,utils.STATE_LEN)

		_, err := pipeR.Read(recv)
		recvStr := string(recv)

		// io failed
		if err != nil || recvStr == utils.STATE_ERR {
			log.Fatalln("fuxxer io failed")
		}

		// STATE_BAD or STATE_OK
		// phone base tool
		fio.Write()
	}
}
