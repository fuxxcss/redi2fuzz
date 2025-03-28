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

// export 
func StartUp(target string) {

	var path,port string

	// Fuxx Target (redis, keydb, redis-stack)
	t,ok := Targets[target]
	if ok {

		// path, port
		path = t[Path]
		port = t[Port]
		
		// need to startup
		rdb := SingleRdb(port)

		alive := rdb.Check_alive()
		if !alive {
			startup_core(path,port)
		}
		
	// target not support 
	}else {
		log.Fatalf("err: %v is not support\n",target)
	}

}

// static
func startup_core(path,port string){

	// cannot find path
	_,err := os.Stat(path)
	if err != nil {
		log.Fatalf("err: %v %v",path,err)
	}

	// AFL_DEBUG get AFL_MAP_SIZE
	var stdout bytes.Buffer
	os.Setenv(fuxx.AFL_DEBUG,"1")

	cmd := exec.Command(path)
	cmd.Stdout = &stdout
	err = cmd.Run()

	// cannot run path
	if err != nil {
		log.Fatalf("err: %v %v\n",path,err)
	}

	// loop stdout
	log.Println("[*] Loop Get AFL_MAP_SIZE.")
	for {
		if strings.Contains(string(stdout),fuxx.AFL_DEBUG_SIZE){
			break
		}
	}
	cmd.Process.Kill()

	// get AFL_MAP_SIZE
	index := strings.Index(string(stdout),fuxx.AFL_DEBUG_SIZE)
	shmsize := ""
	for char := stdout[index] ; char != ',' {
		if char >= '0' && char <= '9' {
			shmsize += char
		}
	}
	
	// startup shm
	shm := SingleShm(shmsize)

	// clean up shm
	shm.Cleanup_Shm()

	// startup db
	// DB ENVs
	os.Setenv(fuxx.AFL_DEBUG,"0")
	os.Setenv(fuxx.AFL_MAP_SIZE,shmsize)
	os.Setenv(fuxx.AFL_SHM_ID,shm)
	// DB args
	args := []string {
		// port
		"--port " + port,
		// process
		"&",
	}
	cmd = exec.Command(path,args...)
	err := cmd.Run()
	
	rdb := SingleRdb(port)
	alive := rdb.Check_alive()

	// db failed
	if err != nil {
		log.Fatalf("err: db %v\n",err)
	}
	if !alive {
		log.Fatalln("err: rdb failed.")
	}

	// db succeed
	log.Printf("[*] DB %v StartUp.\n",path)
	
}