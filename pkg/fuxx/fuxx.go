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
	FTESTCASE_R int = iota +  3
	FTESTCASE_W
	FCTL_R
	FCTL_W
)

// Fuxxer Server phone string
const (
	FSERVER_OK string = "ok"
	FSERVER_BAD string = "bad"
	FSERVER_ERR string = "err"
)

// export
func Fuxx(target,mode,tool string){

	// Fuxx Tool (afl, honggfuzz)
	ftool,ok := utils.Tools[tool]

	if !ok {
		log.Fatalf("err: %v tool is not support\n",tool)
	}

	// Fuxx Mode (dumb, gramfree, fagent)
	fmode,ok := utils.Modes[mode] 

	if !ok {
		log.Fatalf("err: %v mode is not support\n",mode)
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
	
	// testcase pipe for ipc
	testRead,testWrite,err := os.Pipe()
	if err != nil {
		log.Fatalf("err: testcase pipe failed %v\n",err)
	}
	testPipe := []*os.File {
		testRead,
		testWrite 
	}

	// control pipe for ipc
	ctlRead,ctlWrite,err := os.Pipe()
	if err != nil {
		log.Fatalf("err: control pipe failed %v\n",err)
	}
	ctlPipe := []*os.File {
		ctlRead,
		ctlWrite
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
		// ftestcase_R
		testRead,
		// ftestcase_W
		testWrite,
		// fctl_R
		ctlRead,
		// fctl_W
		ctlWrite,
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
	os.Setenv(ftool[utils.TOOLS_ENV_CUSTOM_PATH],"build/libcustom.so")
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
	go fuxxServer(ftarget,ftool,fmode,testPipe,ctlPipe)

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
func fuxxServer(target,tool,mode interface{},tpipe,cpipe []*os.File){

	// init corpus
	corpus := NewCorpus()

	// fuxx loop
	for {

		// read testcase from driver
		recv,err := io.ReadAll(tpipe[0])
		if err != nil {
			log.Printf("err: Fuxx Server read %v.",err)

			// phone driver: err
			cpipe[1].WriteString(FSERVER_ERR)
			return
		}

		// phone driver: bad
		testcase,err := corpus.AddSet(recv)
		if err != nil {

			log.Printf("bad: %v\n",err)
			cpipe[1].WriteString(FSERVER_BAD)
			continue
		}

		// phone driver: ok
		cpipe[1].WriteString(db.FSERVER_OK)
		
		// fuxx command loop
		okCnt := len(testcase.commands)
		
		for index,_ : range len(testcase.commands){
			recv,err = io.ReadAll(cpipe[0])
			if err != nil {
				log.Printf("err: Fuxx Server read %v.",err)

				// phone driver: err
				cpipe[1].WriteString(FSERVER_ERR)
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
				db.StartUp(target,tool)

				testcase.Crash(index)
			}

			// phone driver: ok
			cpipe[1].WriteString(FSERVER_OK)

		}

		// drop testcase
		if okCnt < CORPUS_MINLEN {
			corpus.DropSet(testcase)

		// update weight
		}else {
			corpus.UpdateWeight(testcase)
		}

		// mutate 
		mutated := mode()

		// write testcase to fuxxer
		tpipe[1].WriteString(mutated)

	}
}
