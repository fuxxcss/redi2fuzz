package fuxx

import (
	"os",
	"log",
	"fmt",
	"bytes",
	"strings",
	"os/exec",
	"os/signal",
	"syscall",
	"strconv"

	// "gopkg.in/yaml.v3"
	"github.com/fuxxcss/redi2fuxx/pkg/db"
)

// Fuxxer Server File
const (
	FDRIVER_R int = iota +  3
	FDRIVER_W
	FMUTATOR_R
	FMUTATOR_W
)

// Fuxxer Server phone string
const (
	FSERVER_OK string = "ok"
	FSERVER_BAD string = "bad"
	FSERVER_ERR string = "err"
)

// export
func Fuxx(target,tool string){

	// Fuxx Tool (afl, honggfuzz)
	ftool,ok := utils.Tools[tool]

	if !ok {
		log.Fatalf("err: %v tool is not support\n",tool)
	}

	// Fuxx Target (redis, keydb, redis-stack)
	ftarget,ok := utils.Targets[target]

	if !ok {
		log.Fatalf("err: %v target is not support\n",target)
	}

	// StartUp target first
	shm,err := db.StartUp(ftarget,ftool)
	defer db.ShutDown()

	if err != nil {
		log.Printf("err: %v",err)
	}
	
	// driver pipe for ipc
	dRead,dWrite,err := os.Pipe()
	if err != nil {
		log.Fatalf("err: testcase pipe failed %v\n",err)
	}
	dPipe := []*os.File {
		dRead,
		dWrite
	}

	// mutator pipe for ipc
	mRead,mWrite,err := os.Pipe()
	if err != nil {
		log.Fatalf("err: control pipe failed %v\n",err)
	}
	mPipe := []*os.File {
		mRead,
		mWrite
	}

	// fuxx with rpipe,wpipe
	exe := ftool[utils.TOOLS_EXE]
	args := []string {

		// to do (YAML)
		// dict
		ftool[utils.TOOLS_DICT] + " " + "fuzz/dumb/" + target + ".dict",
		// timeout
		ftool[utils.TOOLS_TIMEOUT] + " " + "5000",
		// input
		ftool[utils.TOOLS_INPUT] + " " + "fuzz/input/" + target,
		// output
		ftool[utils.TOOLS_OUTPUT] + " " + "fuzz/output/" + target,
		// driver
		ftool[utils.TOOLS_DRIVER] + " " + "build/driver",
	}

	fuxxProc := exec.Command(exe,args...)
	fuxxProc.ExtraFiles = []*os.File{
		// driver pipe
		dRead,
		dWrite,
		// mutator pipe
		mRead,
		mWrite,
	}

	// fuxx envs
	// coverage map env must be set
	os.Setenv(utils.CoverageMap,shm.ShmID)
	// debug env
	os.Setenv(ftool[utils.TOOLS_ENV_DEBUG],"0")
	// max size env
	os.Setenv(ftool[utils.TOOLS_ENV_MAX_SIZE],shm.ShmSize)
	// custom flag env
	os.Setenv(ftool[utils.TOOLS_ENV_CUSTOM_FLAG],"1")
	// custom path env
	os.Setenv(ftool[utils.TOOLS_ENV_CUSTOM_PATH],"build/libmutator.so")
	// skip cpufreq env
	os.Setenv(ftool[utils.TOOLS_ENV_SKIP_CPUFREQ],"1")
	// skip bin check env
	os.Setenv(ftool[utils.TOOLS_ENV_SKIP_BIN_CHECK],"1")
	// use asan env
	os.Setenv(ftool[utils.TOOLS_ENV_USE_ASAN],"1")
	// fast cal env
	os.Setenv(ftool[utils.TOOLS_ENV_FAST_CAL],"1")

	// run fuxxer
	err := fuxxProc.Run()
	defer fuxxProc.Process.Kill()

	// failed
	if err != nil {
		log.Fatalf("err: fuxx proc %v\n",err)
	}

	// succeed
	chanExit := make(chan struct{})
	chanExitPrint := make(chan struct{})

	go signalCtl(chanExit)
	go fuxxPrint(fuxxProc,chanExit,chanExitPrint)
	go fuxxServer(ftarget,ftool,dPipe,mPipe)

	// exit
	<-chanExitPrint
		
}

// static
func signalCtl(chanExit chan<- struct{}){

	chanSig := make(chan os.Signal,1)
	signal.Notify(chan_sig, os.Interrupt, syscall.SIGTERM)
	<-chanSig

	log.Println("[*] Fuxx Proc is Killed.")
	chanExit <- struct{}{}

}

// static
func fuxxPrint(fuxxProc *exec.Cmd,chanExit <-chan struct{},chanExitPrint chan<- struct{}){

	// AFL_Print is exit
	defer chanExitPrint <- struct{}{}

	for {
		select {
		
		// chan exit
		case <-chanExit:
			log.Println("[*] Fuxx Printer exit.")
			return
		
		default:
			stdout,err := fuxxProc.StdoutPipe()
	
			// stdout failed
			if err != nil {
				log.Println("err: Fuxx Printer failed.")
				break
			}

			io.Copy(os.Stdout,stdout)
		}
		
	}
}

// static
func fuxxServer(ftarget,ftool interface{},dPipe,mPipe []*os.File){

	// init corpus
	corpus := NewCorpus()

	// fuxx loop
	for {
		// phone driver: tool
		dPipe[1].WriteString(tool)
		
		io.ReadAll(dPipe[0])

		// phone driver: port
		dPipe[1].WriteString(ftarget[TARGET_PORT])

		// read testcase from driver
		recv,err := io.ReadAll(dPipe[0])

		if err != nil {
			log.Printf("err: Fuxx Server read %v.",err)

			// phone driver: err
			dPipe[1].WriteString(FSERVER_ERR)
			return
		}

		// phone driver: bad
		testcase,err := corpus.AddSet(recv)

		if err != nil {

			log.Printf("bad: %v\n",err)
			dPipe[1].WriteString(FSERVER_BAD)
			continue
		}

		// phone driver: ok
		dPipe[1].WriteString(db.FSERVER_OK)
		
		// fuxx command loop
		okCnt := len(testcase.commands)
		
		for index,_ : range len(testcase.commands){

			recv,err = io.ReadAll(dPipe[0])

			if err != nil {
				log.Printf("err: Fuxx Server read %v.",err)

				// phone driver: err
				dPipe[1].WriteString(FSERVER_ERR)
				return
			}

			switch recv {

			// command ok
			case FSERVER_OK:
				testcase.BuildGraph(index)

			// command has fault
			case FSERVER_BAD:
				-- okCnt

			// command crash
			case db.FSERVER_ERR:
				file,err := os.OpenFile(hash,os.O_CREATE | os.O_WRONLY | os.O_TRUNC,0664)
				crash := "[*] Found a crash :)"

				// don't miss crash
				if err != nil {
					log.Println(crash)
					return
				}

				// log crash
				file.WriteString(crash + "\n")
				file.WriteString("index ==> " + strconv.Itoa(cnt) + "\n")
				file.WriteString(recv)

				// restart
				db.StartUp(ftarget,ftool)

				testcase.Crash(index)
			}

			// phone driver: ok
			dPipe[1].WriteString(FSERVER_OK)

		}

		// drop testcase
		if okCnt < CORPUS_MINLEN {
			corpus.DropSet(testcase)

		// update weight
		}else {
			corpus.UpdateWeight(testcase)
		}

		var mutated string

		mutated = corpus.Mutate()

		// write testcase to mutator
		mPipe[1].WriteString(mutated)
	}
}
