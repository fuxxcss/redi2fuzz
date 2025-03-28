package db

import (
	"syscall",
	"strconv",
)

type Shm string

// global
var (
	global_shm Shm
	mutex_shm sync.Mutex
)

// export
func SingleShm(shmsize string) Shm {

	if global_shm == nil {
		mutex_shm.Lock()
		defer mutex_shm.Unlock()
		if global_shm == nil {
			global_shm = NewShm(shmsize)
		}
	}
	return global_shm
}

// export
func NewShm(shmsize string) Shm {

	// ipcmk 
	shmkey := uintptr(syscall.IPC_PRIVATE)
	shmsize := uintptr(strconv.Atoi(shmsize))
	shmflag := uintptr(0666)
	shmid,_,err := syscall.Syscall(syscall.SYS_SHMGET,shmkey,shmsize,shmflag)

	// ipcmk failed
	if err != nil {
		log.Fatalf("err: ipcmk %v\n",err)
	}

	// ipcmk succeed
	log.Printf("[*] Shared Mem ID.%v StartUp.\n",shmid)

	return shmid

}

// public
func (self *Shm) Cleanup_Shm(){

	// attach
	addr,_,err := syscall.Syscall(*self,syscall.SYS_SHMAT,nil,0)

	// attach failed
	if err != nil {
		log.Printf("err: attach shm %v\n",err)
	}
	defer syscall.Syscall(syscall.SYS_SHMDT,addr)

	// cleanup
	clear(addr)

}