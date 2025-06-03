package db

import (
	"log"
	"os"
	"os/exec"

	"github.com/fuxxcss/redi2fuxx/pkg/utils"
)

// export
func StartUp(target utils.TargetsType)  {

	var path, port string

	// path, port
	path = target[utils.TARGET_PATH]
	port = target[utils.TARGET_PORT]

	redi := SingleRedi(port)

	alive := redi.CheckAlive()

	// already startup, shutdown first
	if alive {
		redi.client.Do(redi.ctx,"shutdown")
	}

	// startup redi
	startupRedi(path,port)
}

// static
func startupRedi(path, port string) {

	// cannot find path
	_, err := os.Stat(path)
	if err != nil {
		log.Fatalf("err: %v %v", path, err)
	}

	// DB args
	args := []string{
		// port
		RediPort + " " + port,
	}
	rediProc := exec.Command(path, args...)
	err = rediProc.Start()

	// db failed
	if err != nil {
		log.Fatalln("err: redi failed.")
	}

	redi := SingleRedi(port)

	// waiting redi startup
	for {
		alive := redi.CheckAlive()
		if alive {
			break
		}
	}

	// db succeed
	redi.Proc = rediProc
	redi.args = args
	redi.path = path
	log.Printf("[*] DB %v StartUp.\n", path)

}

func ShutDown() {

	// kill redis
	redi := SingleRedi("")
	redi.Proc.Process.Kill()

}
