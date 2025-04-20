package db

import (
	"os"
	"log"
	"bytes"
	"errors"
	"strings"
	"os/exec"
	"regexp"

	"github.com/fuxxcss/redi2fuxx/pkg/utils"
)

// export 
func StartUp(target utils.TargetsType,tool utils.ToolsType) (*Shm,error) {

	var path,port string

	// path, port
	path = target[utils.TARGET_PATH]
	port = target[utils.TARGET_PORT]
	
	redi := SingleRedi(port)
	alive := redi.CheckAlive()

	// need to startup
	if !alive {
		return startupCore(path,port,tool),nil

	// already startup
	}else {
		return nil,errors.New("Already StartUp.")
	}
		
}

// static
func startupCore(path,port string,tool utils.ToolsType) *Shm {

	// cannot find path
	_,err := os.Stat(path)
	if err != nil {
		log.Fatalf("err: %v %v",path,err)
	}

	// set ENV_DEBUG get map size
	var stdout bytes.Buffer
	os.Setenv(tool[utils.TOOLS_ENV_DEBUG],"1")

	debugProc := exec.Command(path)
	debugProc.Stdout = &stdout
	err = debugProc.Run()

	// cannot run path
	if err != nil {
		log.Fatalf("err: %v %v\n",path,err)
	}

	// deal with stdout
	log.Println("[*] Get Debug Size.")
	var originStr string
	toMatch := tool[utils.TOOLS_ENV_DEBUG_SIZE]

	for {
		originStr = stdout.String()
		if strings.Contains(originStr,toMatch) {
			break
		}
	}
	debugProc.Process.Kill()

	// get debug size
	re := regexp.MustCompile(toMatch + `=(\S+)`)
	isMatch := re.FindStringSubmatch(originStr)

	shmsize := isMatch[1]
	
	// startup shm
	shm := SingleShm(shmsize)

	// clean up shm
	shm.CleanUp()

	// startup db
	// DB ENVs
	os.Setenv(tool[utils.TOOLS_ENV_DEBUG],"0")
	os.Setenv(tool[utils.TOOLS_ENV_MAX_SIZE],shm.ShmSize)
	os.Setenv(tool[utils.TOOLS_ENV_SHM_ID],shm.ShmID)
	// DB args
	args := []string {
		// port
		RediSep + " " + port,
		// daemon
		RediDeamon,
	}
	rediProc := exec.Command(path,args...)
	err = rediProc.Run()
	
	redi := SingleRedi(port)
	alive := redi.CheckAlive()

	// db failed
	if err != nil {
		log.Fatalf("err: db %v\n",err)
	}
	if !alive {
		log.Fatalln("err: redi failed.")
	}

	// db succeed
	redi.Proc = rediProc
	log.Printf("[*] DB %v StartUp.\n",path)

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