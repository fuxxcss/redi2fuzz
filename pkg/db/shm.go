package db

import (
	"syscall",
	"strconv",
)

type Shm struct {
	ShmID string
	ShmSize string
}

// global
var (
	globalShm Shm
	mutexShm sync.Mutex
)

// export
func SingleShm(shmsize string) *Shm {

	if globalShm == nil {
		mutexShm.Lock()
		defer mutexShm.Unlock()
		if globalShm == nil {
			globalShm = NewShm(shmsize)
		}
	}
	return globalShm
}

// export
func NewShm(shmsize string) *Shm {

	shm := new(Shm)

	// ipcmk 
	key := uintptr(syscall.IPC_PRIVATE)
	size := uintptr(strconv.Atoi(shmsize))
	flag := uintptr(0666)
	shmid,_,err := syscall.Syscall(syscall.SYS_SHMGET,key,size,flag)

	// ipcmk failed
	if err != nil {
		log.Fatalf("err: ipcmk %v\n",err)
	}

	// ipcmk succeed
	log.Printf("[*] Shared Mem ID = %v StartUp, Size = %v.\n",shmid,shmsize)
	shm.ShmID = shmid
	shm.ShmSize = shmsize

	return shm

}

// public
func (self *Shm) CleanUp(){

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

// public
func (self *Shm) Close() {

	_, _, err := syscall.Syscall(syscall.SYS_SHMCTL, uintptr(self.ShmID), syscall.IPC_RMID, 0)

	// free shm failed
	if err != nil {
		log.Println("err: %v",err)
	}
}