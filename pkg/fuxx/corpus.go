package fuxx

import (
	"crypto/md5"
	"errors"
	"log"
	"math/rand"
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
	CORPUS_MINLEN    int = 15
	CORPUS_MAXLEN    int = 45
	CORPUS_THRESHOLD int = 50
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
func NewTestcase(testcase, hash string) *Testcase {

	testPtr := new(Testcase)

	// redis split
	sliceStr := strings.Split(testcase, db.RediSep)
	sliceSize := len(sliceStr)
	testPtr.commands = make([]Command, sliceSize)

	for i, str := range sliceStr {

		testPtr.commands[i] = make(Command, 3)

		// text
		testPtr.commands[i][CMD_TEXT] = str

		// args
		sliceToken := strings.Split(str, db.RediTokenSep)
		tokens := make([]string, 0)

		for _, token := range sliceToken {
			if token != "" {
				tokens = append(tokens, token)
			}
		}
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
	log.Println(snapshots)
	if err != nil {
		return err
	}

	// update weight
	createLen := len(snapshots[0])
	deleteLen := len(snapshots[1])
	log.Println(createLen, deleteLen)
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
func NewCorpus() *Corpus {

	corpus := new(Corpus)
	corpus.hashset = make(map[string]bool, CORPUS_THRESHOLD)
	corpus.order = make([]*Testcase, 0)

	return corpus
}

// public
// if exist return nil,err
// else     return ptr,nil
func (self *Corpus) AddSet(testcase string) (*Testcase, error) {

	// repeat testcase
	sum := md5.Sum([]byte(testcase))
	hash := string(sum[:])

	_, ok := self.hashset[hash]
	if ok {
		return nil, errors.New("Repeat Testcase.")
	}

	// new testcase
	self.hashset[hash] = true

	return NewTestcase(testcase, hash), nil

}

// public
func (self *Corpus) DropSet(testPtr *Testcase) {

	delete(self.hashset, testPtr.hash)
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

	// threshold
	if orderSize > CORPUS_THRESHOLD {
		self.order = self.order[1:]
	}

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

	for i := 0; i < length; i++ {

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
	}

	return mutated

}
