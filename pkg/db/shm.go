package db

import (
	"log"
	"sync"
	"strconv"
	"golang.org/x/sys/unix"
)

type Shm struct {
	ShmID string
	ShmSize string
}

// global
var (
	globalShm *Shm
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
	size, _ := strconv.Atoi(shmsize)
	id, err := unix.SysvShmGet(unix.IPC_PRIVATE, size,0666)

	// ipcmk failed
	if err != nil {
		log.Fatalf("err: ipcmk %v\n", err)
	}

	// ipcmk succeed
	log.Printf("[*] Shared Mem ID = %v StartUp, Size = %v.\n", id, shmsize)

	shmid := strconv.Itoa(id)

	shm.ShmID = shmid
	shm.ShmSize = shmsize

	return shm

}

// public
func (self *Shm) CleanUp(){

	// attach
	id, _ := strconv.Atoi(self.ShmID)
	addr, err := unix.SysvShmAttach(id, 0, 0)

	// attach failed
	if err != nil {
		log.Printf("err: attach shm %v\n", err)
	}
	defer unix.SysvShmDetach(addr)

	// cleanup
	clear(addr)

}

// public
func (self *Shm) Close() {

	id, _ := strconv.Atoi(self.ShmID)
	_, err := unix.SysvShmCtl(id, unix.IPC_RMID, nil)

	// free shm failed
	if err != nil {
		log.Println("err:", err)
	}
}