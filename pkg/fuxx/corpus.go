package fuxx

import (
	"crypto/md5"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fuxxcss/redi2fuxx/pkg/db"
)

// Command
type Command map[CommandFeature]interface{}

// Command features
type CommandFeature int

const (
	CMD_TEXT CommandFeature = iota
	CMD_TOKEN
	CMD_ACTION
)

type Testcase struct {
	hash     string
	graph    []*Graph
	weight   float32
	commands []Command
}

// corpus max
const (
	CORPUS_MINLEN int = 15
	CORPUS_MAXLEN int = 45
)

// corpus factor
const (
	CORPUS_FACTOR_LEN int = 1
	CORPUS_FACTOR_COV int = 1

	CORPUS_FACTOR_KEEP   int = 1
	CORPUS_FACTOR_CREATE int = 2
	CORPUS_FACTOR_DELETE int = 2
	CORPUS_FACTOR_MIX    int = 3
	CORPUS_FACTOR_CRASH  int = 100
)

type Corpus struct {
	hashset map[string]bool
	order   []*Testcase
}

// export
func NewTestcase(redi *db.Redi, testcase,hash string) *Testcase {

	testPtr := new(Testcase)

	// split
	sliceStr := redi.SplitLine(testcase)
	sliceSize := len(sliceStr)
	testPtr.commands = make([]Command, sliceSize)

	for i, str := range sliceStr {

		testPtr.commands[i] = make(Command, 3)

		// text
		testPtr.commands[i][CMD_TEXT] = str

		// args
		tokens := redi.SplitToken(str)
		testPtr.commands[i][CMD_TOKEN] = tokens

		// action
		testPtr.commands[i][CMD_ACTION] = 0
	}

	testPtr.hash = hash
	testPtr.graph = make([]*Graph, sliceSize)
	testPtr.weight = float32(0)

	return testPtr
}

// public
func (self *Testcase) BuildGraph(index int) error {

	redi := db.SingleRedi("")
	snapshots, err := redi.Diff()
	/*
	log.Println("snapshot:")
	db.DiffDebug(snapshots)
	*/
	if err != nil {
		return err
	}

	// update weight
	createLen := len(snapshots[0])
	deleteLen := len(snapshots[1])

	if createLen > 0 && deleteLen > 0 {
		self.commands[index][CMD_ACTION] = CORPUS_FACTOR_MIX
	} else if createLen > 0 {
		self.commands[index][CMD_ACTION] = CORPUS_FACTOR_CREATE
	} else if deleteLen > 0 {
		self.commands[index][CMD_ACTION] = CORPUS_FACTOR_DELETE
	} else {
		self.commands[index][CMD_ACTION] = CORPUS_FACTOR_KEEP
	}

	// build graph
	command := self.commands[index][CMD_TEXT].(string)
	self.graph[index] = NewGraph()
	self.graph[index].Build(snapshots, command)
	/*
	log.Println("graph:")
	self.graph[index].Debug()
	*/
	return nil
}

// public
func (self *Testcase) Crash(index int) {

	self.commands[index][CMD_ACTION] = CORPUS_FACTOR_CRASH
}

// mutate str, int
func (self *Testcase) Mutate(r *rand.Rand, index int) {

	sliceToken := self.commands[index][CMD_TOKEN].([]string)

	for i, token := range sliceToken {

		// int mutate
		_, errI := strconv.Atoi(token)
		_, errF := strconv.ParseFloat(token, 32)

		if errI == nil && errF == nil {
			chosen := r.Intn(len(InterestingNum))
			sliceToken[i] = InterestingNum[chosen]
		}

		// str mutate
		sliceStr := strings.Split(token, db.RediStrSep)

		if len(sliceStr) >= 3 {
			sliceStr[1] = MutateStr(r, sliceStr[1])
			mutatedStr := ""

			// assemble
			for _, str := range sliceStr {
				mutatedStr += str + db.RediStrSep
			}
			sliceToken[i] = mutatedStr
		}
	}

	// assemble
	mutatedToken := ""

	for _, token := range sliceToken {
		mutatedToken += token + db.RediTokenSep
	}

	self.graph[index].cmdV.vdata = mutatedToken

}

// export
func NewCorpus(redi *db.Redi, path string) *Corpus {

	corpus := new(Corpus)
	corpus.hashset = make(map[string]bool, 0)
	corpus.order = make([]*Testcase, 0)

	// add all testcase
	filepath.Walk(path, func(file string, info os.FileInfo, err error) error {

		if err != nil {
			log.Fatalln("err: wrong queue path.")
		}

		if info.IsDir() {
			return nil
		}

		// read file
		content, err := os.ReadFile(file)

		if err != nil {
			log.Println("err: read queue failed.",file)
		}

		corpus.AddSet(redi, content)

		return nil
	})

	return corpus
}

// public
// if exist return nil,err
// else     return ptr,nil
func (self *Corpus) AddSet(redi *db.Redi, testcase []byte) {

	// repeat testcase
	sum := md5.Sum(testcase)
	hash := string(sum[:])

	_, ok := self.hashset[hash]

	// new testcase
	if !ok {

		self.hashset[hash] = true
		self.order = append(self.order, NewTestcase(redi, string(testcase), hash))
	}

}

// public
func (self *Corpus) UpdateWeight(testPtr *Testcase) {

	orderSize := len(self.order)

	length := len(testPtr.commands) * CORPUS_FACTOR_LEN
	actions := orderSize * CORPUS_FACTOR_COV

	for _, cmd := range testPtr.commands {
		actions += cmd[CMD_ACTION].(int)
	}

	// calc weight
	testPtr.weight = float32(actions) / float32(length)

	// insert testPtr
	pos := sort.Search(orderSize, func(i int) bool { return self.order[i].weight >= testPtr.weight })
	self.order = slices.Insert(self.order, pos, testPtr)

}

// public
func (self *Corpus) Select(r *rand.Rand) (*Testcase, int) {

	// roulette wheel selection
	sumFloat := float32(0)

	for _, testPtr := range self.order {
		sumFloat += testPtr.weight
	}

	randFloat := r.Float32() * sumFloat
	sumFloat = float32(0)

	// selected testcase
	var testSelect *Testcase = nil
	for _, testPtr := range self.order {
		sumFloat += testPtr.weight
		if sumFloat > randFloat {
			testSelect = testPtr
			break
		}
	}

	sumInt := 0
	for _, command := range testSelect.commands {
		sumInt += command[CMD_ACTION].(int)
	}

	randInt := r.Intn(sumInt)
	sumInt = 0

	// selected command
	cmdSelect := 0
	for index, command := range testSelect.commands {
		sumInt += command[CMD_ACTION].(int)
		if sumInt > randInt {
			cmdSelect = index
			break
		}
	}

	return testSelect, cmdSelect

}

func (self *Corpus) Mutate() string {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// mutated len
	length := r.Intn(CORPUS_MAXLEN-CORPUS_MINLEN) + CORPUS_MINLEN

	mutated := ""

	sliceGraph := make([]*Graph, 0)

	for i := 0; i < length; {

		// select one command
		testPtr, cmdIndex := self.Select(r)
		graph := testPtr.graph[cmdIndex]

		// match command
		if len(graph.cmdV.prev) > 0 {

			isMatched := false

			for _, g := range sliceGraph {

				// match succeed
				if isMatched = graph.Match(g); isMatched {
					break
				}
			}

			// match failed
			if !isMatched {
				continue
			}
		} else {

			// mutate graph
			graph.Mutate(r)
		}

		// mutate testcase
		testPtr.Mutate(r, cmdIndex)

		sliceGraph = append(sliceGraph, graph)
		mutated += graph.cmdV.vdata + db.RediSep

		// one line is ready
		i++
	}

	return mutated

}
