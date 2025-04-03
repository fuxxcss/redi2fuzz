package db

import (
	"os",
	"log",
	"bytes",
	"sync",
	"strings",
	"os/exec",
	"syscall",
	"strconv"

	"github.com/fuxxcss/redi2fuxx/pkg/fuxx"
)

const (

)

// export 
func StartUp(target,tool interface{}) *Shm,error {

	var path,port string

	// path, port
	path = target[Path]
	port = target[Port]
	
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
func startupCore(path,port string,tool interface{}) *Shm{

	// cannot find path
	_,err := os.Stat(path)
	if err != nil {
		log.Fatalf("err: %v %v",path,err)
	}

	// set ENV_DEBUG get map size
	var stdout bytes.Buffer
	os.Setenv(tool[TOOLS_ENV_DEBUG],"1")

	debugProc := exec.Command(path)
	debugProc.Stdout = &stdout
	err = debugProc.Run()

	// cannot run path
	if err != nil {
		log.Fatalf("err: %v %v\n",path,err)
	}

	// loop stdout
	log.Println("[*] Loop Get Debug Size.")
	for {
		if strings.Contains(string(stdout),tool[utils.TOOLS_ENV_DEBUG_SIZE]){
			break
		}
	}
	debugProc.Process.Kill()

	// get debug size
	index := strings.Index(string(stdout),tool[utils.TOOLS_ENV_DEBUG_SIZE])
	shmsize := ""
	for char := stdout[index] ; char != ',' {
		if char >= '0' && char <= '9' {
			shmsize += char
		}
	}
	
	// startup shm
	shm := SingleShm(shmsize)

	// clean up shm
	shm.CleanUp()

	// startup db
	// DB ENVs
	os.Setenv(tool[TOOLS_ENV_DEBUG],"0")
	os.Setenv(tool[TOOLS_ENV_MAX_SIZE],shm.ShmSize)
	os.Setenv(tool[TOOLS_ENV_SHM_ID],shm.ShmID)
	// DB args
	args := []string {
		// port
		RediSep + " " + port,
		// daemon
		RediDeamon,
	}
	rediProc = exec.Command(path,args...)
	err := rediProc.Run()
	
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
	
}

func ShutDown() {

	// kill redis
	redi := SingleRedi(nil)
	redi.Proc.Process.Kill()

	// close shm
	shm := SingleShm(nil)
	shm.Close()
}