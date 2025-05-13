package db

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/fuxxcss/redi2fuxx/pkg/utils"
)

// export
func StartUp(target utils.TargetsType, tool utils.ToolsType) *Shm {

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

	// core func
	return startupCore(path, port, tool)

}

// static
func startupCore(path, port string, tool utils.ToolsType) *Shm {

	// cannot find path
	_, err := os.Stat(path)
	if err != nil {
		log.Fatalf("err: %v %v", path, err)
	}

	// set ENV_DEBUG get map size
	var stderr bytes.Buffer
	os.Setenv(tool[utils.TOOLS_ENV_DEBUG], "1")

	debugProc := exec.Command(path)
	debugProc.Stderr = &stderr
	err = debugProc.Start()

	// cannot run path
	if err != nil {
		log.Fatalf("err: %v %v\n", path, err)
	}

	// deal with stdout
	var originStr string
	toMatch := tool[utils.TOOLS_ENV_DEBUG_SIZE]

	for {
		originStr = stderr.String()
		if strings.Contains(originStr, toMatch) {
			break
		}
	}
	debugProc.Process.Kill()

	// get debug size
	re := regexp.MustCompile(toMatch + ` ([0-9]+)`)
	isMatch := re.FindStringSubmatch(originStr)

	shmsize := isMatch[1]

	// startup shm
	shm := SingleShm(shmsize)

	// clean up shm
	shm.CleanUp()

	// startup db
	// DB ENVs
	os.Setenv(tool[utils.TOOLS_ENV_DEBUG], "0")
	os.Setenv(tool[utils.TOOLS_ENV_MAX_SIZE], shm.ShmSize)
	os.Setenv(tool[utils.TOOLS_ENV_SHM_ID], shm.ShmID)

	// DB args
	args := []string{
		// port
		RediPort + " " + port,
	}
	rediProc := exec.Command(path, args...)
	err = rediProc.Start()

	// db failed
	if err != nil {
		shm.Close()
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
	log.Printf("[*] DB %v StartUp.\n", path)

	return shm

}

func ShutDown() {

	// kill redis
	redi := SingleRedi("")
	redi.Proc.Process.Kill()

	// close shm
	shm := SingleShm("")
	shm.Close()
}
