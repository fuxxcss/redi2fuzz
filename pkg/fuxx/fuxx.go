package fuxx

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fuxxcss/redi2fuxx/pkg/db"
	"github.com/fuxxcss/redi2fuxx/pkg/utils"
)

// export
func Fuxx(target string) {

	// Fuxx Target (redis, keydb, redis-stack)
	ftarget, ok := utils.Targets[target]

	if !ok {
		log.Fatalf("err: %v target is not support\n", target)
	}

	// StartUp target first
	db.StartUp(ftarget)
	defer db.ShutDown()

	chanExit := make(chan struct{})

	// exit control
	go signalCtl(chanExit)

	// fuxx server
	go fuxxServer(target)

	// exit
	<-chanExit

}

// static
func signalCtl(chanExit chan<- struct{}) {

	chanSig := make(chan os.Signal, 1)
	signal.Notify(chanSig, os.Interrupt, syscall.SIGTERM)
	<-chanSig

	fmt.Println("[*] Fuxx Proc is Killed.")
	chanExit <- struct{}{}

}

// static
func fuxxLoop(redi *db.Redi, ptr *Testcase) {

	// fuxx command loop
	var rediState string
	var err error

	length := len(ptr.commands)
	okCnt := length

	for index := 0; index < length; index++ {

		// execute command
		tokens := ptr.commands[index][CMD_TOKEN]
		rediState, err = redi.Execute(tokens.([]string))

		switch rediState {

		// command ok
		case utils.STATE_OK:
			err := ptr.BuildGraph(index)

			if err != nil {
				log.Printf("err: Build Gragh %v.", err)
			}

		// command has fault
		case utils.STATE_BAD:
			fmt.Println("[*] bad cmd:", ptr.commands[index][CMD_TEXT])
			fmt.Println(err)
			okCnt--

		// command crash
		case utils.STATE_ERR:
			file, err := os.OpenFile(ptr.hash, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
			crash := "[*] Found a crash :)"

			// don't miss crash
			if err != nil {

				var lover []string

				for _, cmd := range ptr.commands {
					lover = append(lover, cmd[CMD_TEXT].(string))
				}

				fmt.Println(crash)
				log.Fatalln(lover)
			}

			// log crash
			ptr.Crash(index)
			indexStr := strconv.Itoa(index)

			file.WriteString(crash + "\n")
			file.WriteString("index ==> " + indexStr + "\n")

			for _, cmd := range ptr.commands {

				file.WriteString(cmd[CMD_TEXT].(string))
			}

			// restart
			redi.Restart()

		}

	}

	// bad testcase
	if okCnt < CORPUS_MINLEN {

		fmt.Println("[*] bad queue.")
	}

}

// static
func fuxxServer(target string) {

	ftarget := utils.Targets[target]

	// init redi
	redi := db.SingleRedi(ftarget[utils.TARGET_PORT])

	// init corpus
	corpus := NewCorpus(redi, ftarget[utils.QUEUE_PATH])
	fmt.Println("[*] init corpus.")

	for _, testPtr := range corpus.order {

		// clean up database and redi snapshot
		err := redi.CleanUp()

		if err != nil {
			log.Fatalln("err: clean up failed")
		}

		fuxxLoop(redi, testPtr)
		corpus.UpdateWeight(testPtr)
	}

	// fuxx loop
	var mutated string
	var lines []string
	var tokens []string
	tryCnt := 0

	for {

		var rediState string

		// mutate
		mutated = corpus.Mutate()

		// clean up database
		err := redi.CleanUp()

		if err != nil {
			log.Fatalln("clean up failed")
		}

		lines = redi.SplitLine(mutated)

		for index, line := range lines {

			tokens = redi.SplitToken(line)
			rediState, err = redi.Execute(tokens)

			// print
			fmt.Printf("fuxx count: %d\n", tryCnt)
			fmt.Printf("fuxx line: %s\n", line)
			if err != nil {
				fmt.Println(err)
			}

			// command crash
			if rediState == utils.STATE_ERR {

				sum := md5.Sum([]byte(mutated))
				hash := string(sum[:])
				file, err := os.OpenFile(hash, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
				crash := "[*] Found a crash :)"

				// don't miss crash
				if err != nil {

					fmt.Println(crash)
					log.Fatalln(mutated)
				}

				// log crash
				indexStr := strconv.Itoa(index)

				file.WriteString(crash + "\n")
				file.WriteString("index ==> " + indexStr + "\n")
				file.WriteString(mutated)

				// restart
				redi.Restart()

			}
		}

		tryCnt++
	}
}
