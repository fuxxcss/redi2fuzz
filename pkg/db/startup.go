package db

import (
	"fmt",
	"os",
	"bytes",
	"strings",
	"os/exec",
	"syscall",
	"strconv"
)

fuxxTarget := map[string]bool{
	"redis":true,
	"keydb":true,
	"redis-stack":true,
}

func StartUp(target string){

	var path,port string

	// Fuxx Target (redis, keydb, redis-stack)
	_,ok := fuxxTarget[target]
	if ok {
		switch target{

		// redis-server path
		case "redis", "redis-stack":
			port = "6379"
			path = "/usr/local/redis/src/redis-server"

		// keydb-server path
		case "keydb":
			port = "6380"
			path = "/usr/local/keydb/src/keydb-server"
		}

		StartUp_Core(path,port)

	// target not support 
	}else { 
		fmt.Printf("err: %v is not support\n",target)
		os.Exit(1)
	}

}

func StartUp_Core(path,port string){

	// AFL ENVs
	AFL_DEBUG := "AFL_DEBUG"
	AFL_MAP_SIZE := "__afl_map_size"
	AFL_SHM_ID := "__AFL_SHM_ID"

	// cannot find path
	_,err := os.Stat(path)
	if err != nil {
		fmt.Printf("err: %v %v",path,err)
		os.Exit(1)
	}

	// AFL_DEBUG get AFL_MAP_SIZE
	var stdout bytes.Buffer
	os.Setenv(AFL_DEBUG,"1")

	cmd := exec.Command(path)
	cmd.Stdout = &stdout
	err = cmd.Run()

	// cannot run path
	if err != nil {
		fmt.Printf("err: %v %v\n",path,err)
		os.Exit(1)
	}

	// loop stdout
	for {
		if strings.Contains(string(stdout),AFL_MAP_SIZE){
			break;
		}
	}
	cmd.Process.Kill()

	// get AFL_MAP_SIZE
	var max_size string
	index := strings.Index(string(stdout),AFL_MAP_SIZE)
	for char := stdout[index] ; char != ',' {
		if char >= '0' && char <= '9' {
			max_size += char
		}
	}

	shmid := StartUp_Shm(max_size)
	
	... TODO Startup_DB
	// startup db
	os.Setenv(AFL_DEBUG,"0")
	os.Setenv(AFL_MAP_SIZE,max_size)
	os.Setenv(AFL_SHM_ID,shmid)
	arg1 := "--port " + port
	arg2 := "&"
	cmd = exec.Command(path,arg1,arg2)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true

}

func StartUp_Shm(size string) string {

	... TODO
	// ipcmk 
	shmkey := uintptr(syscall.IPC_PRIVATE)
	shmsize := uintptr(strconv.Atoi(max_size))
	shmflag := uintptr(0666)
	shmid,_,err := syscall.Syscall(syscall.SYS_SHMGET,shmkey,shmsize,shmflag)

	// ipcmk failed
	if err != nil {
		fmt.Printf("err: ipcmk %v\n",err)
		os.Exit(1)
	}

	// ipcmk succeed

	fmt.Printf("[*] Shared Mem ID.%v StartUp.",shmid)


}