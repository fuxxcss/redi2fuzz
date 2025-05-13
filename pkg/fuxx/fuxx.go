package fuxx

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	//"gopkg.in/yaml.v3"
	"github.com/fuxxcss/redi2fuxx/pkg/db"
	"github.com/fuxxcss/redi2fuxx/pkg/utils"
)

// export
func Fuxx(target, tool string) {

	// Fuxx Tool (afl, honggfuzz)
	ftool, ok := utils.Tools[tool]

	if !ok {
		log.Fatalf("err: %v tool is not support\n", tool)
	}

	// Fuxx Target (redis, keydb, redis-stack)
	ftarget, ok := utils.Targets[target]

	if !ok {
		log.Fatalf("err: %v target is not support\n", target)
	}

	// StartUp target first
	shm := db.StartUp(ftarget, ftool)
	defer db.ShutDown()

	// driver testcase string pipe for ipc
	strRead, strWrite, err := os.Pipe()
	if err != nil {
		log.Panicf("err: fserver string pipe failed %v\n", err)
	}

	// driver control pipe for ipc
	ctlRead, ctlWrite, err := os.Pipe()
	if err != nil {
		log.Panicf("err: fserver control pipe failed %v\n", err)
	}

	dPipe := []*os.File{
		strRead,
		ctlWrite,
	}

	// mutator pipe for ipc
	mutRead, mutWrite, err := os.Pipe()
	if err != nil {
		log.Panicf("err: fserver mutate pipe failed %v\n", err)
	}

	mPipe := []*os.File{
		mutWrite,
	}

	// fuxx with rpipe,wpipe
	exe := ftool[utils.TOOLS_EXE]
	args := []string{
		// timeout
		ftool[utils.TOOLS_TIMEOUT] + " " + "5000",
		// input
		ftool[utils.TOOLS_INPUT] + "fuzz/input/" + target,
		// output
		ftool[utils.TOOLS_OUTPUT] + "fuzz/output/" + target,
		// driver
		"build/driver",
	}

	fuxxProc := exec.Command(exe, args...)
	fuxxProc.ExtraFiles = []*os.File{
		// driver pipe
		ctlRead,
		strWrite,
		// mutator pipe
		mutRead,
	}

	// fuxx printer
	fuxxProc.Stdout = os.Stdout
	fuxxProc.Stderr = os.Stderr

	// fuxx envs
	// coverage map env must be set
	os.Setenv(utils.CoverageMap, shm.ShmID)
	// fuxx tool
	os.Setenv(utils.BaseTool, tool)
	// debug env
	os.Setenv(ftool[utils.TOOLS_ENV_DEBUG], "0")
	// max size env
	os.Setenv(ftool[utils.TOOLS_ENV_MAX_SIZE], shm.ShmSize)
	// custom flag env
	os.Setenv(ftool[utils.TOOLS_ENV_CUSTOM_FLAG], "1")
	// custom path env
	os.Setenv(ftool[utils.TOOLS_ENV_CUSTOM_PATH], "build/libmutator.so")
	// skip cpufreq env
	os.Setenv(ftool[utils.TOOLS_ENV_SKIP_CPUFREQ], "1")
	// skip bin check env
	os.Setenv(ftool[utils.TOOLS_ENV_SKIP_BIN_CHECK], "1")
	// use asan env
	os.Setenv(ftool[utils.TOOLS_ENV_USE_ASAN], "1")
	// fast cal env
	os.Setenv(ftool[utils.TOOLS_ENV_FAST_CAL], "1")

	// run fuxxer
	err = fuxxProc.Start()
	defer fuxxProc.Process.Kill()

	// failed
	if err != nil {
		log.Panicf("err: fuxx proc %v\n", err)
	}

	// succeed
	chanExit := make(chan struct{})

	go signalCtl(chanExit)
	go fuxxServer(target, tool, dPipe, mPipe)

	// exit
	<-chanExit

}

// static
func signalCtl(chanExit chan<- struct{}) {

	chanSig := make(chan os.Signal, 1)
	signal.Notify(chanSig, os.Interrupt, syscall.SIGTERM)
	<-chanSig

	log.Println("[*] Fuxx Proc is Killed.")
	chanExit <- struct{}{}

}

// static
func fuxxServer(target, tool string, dPipe, mPipe []*os.File) {

	ftarget := utils.Targets[target]
	ftool := utils.Tools[tool]

	// init corpus
	corpus := NewCorpus()

	// init redi
	redi := db.SingleRedi(ftarget[utils.TARGET_PORT])

	// init buffer
	recv := make([]byte, utils.MaxSize)

	// fuxx loop
	for {

		// read testcase from driver
		size, err := dPipe[0].Read(recv)

		if err != nil {
			log.Printf("err: Fuxx Server read %v.", err)

			// phone driver: err
			dPipe[1].WriteString(utils.STATE_ERR)
			return
		}

		// bad testcase, skip it
		origin := string(recv[:size])
		testPtr, err := corpus.AddSet(origin)
		
		if err != nil {
			dPipe[1].WriteString(utils.STATE_BAD)
			continue
		}

		// clean up database
		err = redi.CleanUp()

		if err != nil {
			log.Fatalln("clean up failed")
		}

		// fuxx command loop
		var rediState string

		length := len(testPtr.commands)
		okCnt := length

		for index := 0; index < length; index++ {

			// execute command
			cmd := testPtr.commands[index][CMD_TOKEN]
			rediState = redi.Execute(cmd.([]string))

			switch rediState {

			// command ok
			case utils.STATE_OK:
				err := testPtr.BuildGraph(index)

				if err != nil {
					log.Printf("err: Build Gragh %v.", err)
				}

			// command has fault
			case utils.STATE_BAD:
				log.Println("bad cmd:", index)
				okCnt--

			// command crash
			case utils.STATE_ERR:
				file, err := os.OpenFile(testPtr.hash, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
				crash := "[*] Found a crash :)"

				// don't miss crash
				if err != nil {
					log.Println(crash)
					return
				}

				// log crash
				indexStr := strconv.Itoa(index)

				file.WriteString(crash + "\n")
				file.WriteString("index ==> " + indexStr + "\n")
				file.WriteString(origin)

				// restart
				db.StartUp(ftarget, ftool)

				testPtr.Crash(index)
			}

			// phone driver: ok
			dPipe[1].WriteString(utils.STATE_OK)

		}

		// drop testcase
		if okCnt < CORPUS_MINLEN {
			log.Println("dropped")
			corpus.DropSet(testPtr)

		// update weight
		} else {
			corpus.UpdateWeight(testPtr)
		}

		// mutate
		mutated := corpus.Mutate()

		// write testcase to mutator
		mPipe[1].WriteString(mutated)
	}
}
