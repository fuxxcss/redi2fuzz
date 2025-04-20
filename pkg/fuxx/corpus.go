package fuxx

import (
	"crypto/md5"
	"errors"
	"math/rand"
	"sort"
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
	CMD_TIME
	CMD_ACTION
)

type Testcase struct {
	hash   string
	graph  []*Graph
	weight float32
	// unused
	time     int
	commands []Command
}

// corpus max
const (
	CORPUS_MINLEN    int = 20
	CORPUS_MAXLEN    int = 60
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
	testPtr.commands = make([]Command, len(sliceStr))

	for i, str := range sliceStr {

		testPtr.commands[i] = make(Command, 3)
		// text
		testPtr.commands[i][CMD_TEXT] = str
		// unused
		testPtr.commands[i][CMD_TIME] = 0
		// action
		testPtr.commands[i][CMD_ACTION] = 0
	}

	testPtr.hash = hash
	testPtr.graph = make([]*Graph, 1)
	testPtr.weight = float32(0)

	return testPtr
}

// public
func (self *Testcase) BuildGraph(index int) error {

	redi := db.SingleRedi("")
	snapshots, err := redi.Diff()

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
	self.graph[index].Build(snapshots, command)

	return nil
}

// public
func (self *Testcase) Crash(index int) {

	self.commands[index][CMD_ACTION] = CORPUS_FACTOR_CRASH
}

// export
func NewCorpus() *Corpus {

	corpus := new(Corpus)
	corpus.hashset = make(map[string]bool, CORPUS_THRESHOLD)
	corpus.order = make([]*Testcase, CORPUS_THRESHOLD)

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

	length := len(testPtr.commands) * CORPUS_FACTOR_LEN
	actions := len(self.order) * CORPUS_FACTOR_COV

	for _, cmd := range testPtr.commands {
		actions += cmd[CMD_ACTION].(int)
	}

	// calc weight
	testPtr.weight = float32(actions) / float32(length)

	// insert testPtr
	pos := sort.Search(len(self.order), func(i int) bool { return self.order[i].weight >= testPtr.weight })
	copy(self.order[(pos+1):], self.order[pos:])
	self.order[pos] = testPtr

	// threshold
	if len(self.order) > CORPUS_THRESHOLD {
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

	sliceGraph := make([]*Graph, 1)
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
			graph.MutateGraph(r)
		}

		// mutate str, int
		graph.cmdV.vdata = MutateToken(r,graph.cmdV.vdata)

		sliceGraph = append(sliceGraph, graph)
		mutated += graph.cmdV.vdata + db.RediSep
	}

	return mutated

}
