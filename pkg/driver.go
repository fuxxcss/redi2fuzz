package main

import (
	"log"
	"github.com/fuxxcss/redi2fuxx/pkg/db"
)

// Fuxxer Server File
const (
	FDRIVER_R int = iota +  3
	FDRIVER_W
	FMUTATOR_R
	FMUTATOR_W
)

// Fuxxer Server phone string
const (
	FSERVER_OK string = "ok"
	FSERVER_BAD string = "bad"
	FSERVER_ERR string = "err"
)

const (
	MaxSize int = 0x100000
)

func main(){

	// pipes
	pipeR := os.NewFile(FDRIVER_R, "Read")
	pipeW := os.NewFile(FDRIVER_W, "Write")

	tool,err := io.ReadAll(pipeR)

	if !err { 
		log.Fatalln("fuxxer io failed")
	}

	pipeW.WriteString(FSERVER_OK)

	port,err := io.ReadAll(pipeR)

	if !err { 
		log.Fatalln("fuxxer io failed")
	}

	// init FIO
	fio := SingleFIO(tool)

	// start forkserver
	fio.Start()

	// avoid: out of memory
	buffer := make([]byte,MaxSize)

	for {

		// clean up database
		redi := db.SingleRedi(port)
		err := redi.CleanUp()

		if !err { 
			log.Fatalln("clean up failed")
		}
		
		// get one testcase
		length := fio.Read(&buffer[0],MaxSize)

		if length <= 0 { 
			log.Fatalln("next testcase failed")
		}

		trimed := bytes.TrimRight(buffer,"\x00")
		testcase := string(trimed)

		// phone fuxxer server
		pipeW.WriteString(testcase)
		
		recv,err := io.ReadAll(pipeR)

		if err != nil || recv == FSERVER_ERR {
			log.Fatalln("fuxxer io failed")
		}

		if recv == FSERVER_BAD {

			fio.Write()
			continue
		}

		// execute command
		state := db.REDI_OK
		sliceCmd := strings.Split(testcase,db.RediSep)

		for _,command := range sliceCmd {

			state = redi.Execute(command)

			pipeW.WriteString(state)

			recv,err := io.ReadAll(fcRead)

			if err != nil || recv == FSERVER_ERR {
				log.Fatalln("fuxxer io failed")
			}
		}
		
		// report
		fio.Write()
	}
}