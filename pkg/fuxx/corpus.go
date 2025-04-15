package fuxx

import (
    "sort",
    "errors"
    "crypto/md5"
    "strings"

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

    hash string
    graph []Graph
    weight float
    // unused
    time int
    commands []Command
}

// corpus max
const (
    CORPUS_MINLEN int = 20
    CORPUS_MAXLEN int = 60
    CORPUS_THRESHOLD int = 50
)

// corpus factor
const (
    CORPUS_FACTOR_LEN int = 1
    CORPUS_FACTOR_COV int = 1
    
    CORPUS_FACTOR_KEEP int = 1
    CORPUS_FACTOR_CREATE int = 2
    CORPUS_FACTOR_DELETE int = 2
    CORPUS_FACTOR_MIX int = 3
    CORPUS_FACTOR_CRASH int = 100
)

type Corpus struct {

    hashset map[string]bool
    order []*Testcase
}

// export
func NewTestcase(testcase,hash string) *Testcase {

    testPtr := new(Testcase)

    // redis split 
    slice := strings.Split(testcase,db.RediSep)
    testPtr.commands = make([]Command,len(slice))

    for i,str := range slice {
        
        testPtr.commands[i] = make(map[string]int,3)
        // text 
        testPtr.commands[i][CMD_TEXT] = str
        // unused
        testPtr.commands[i][CMD_TIME] = 0
        // action 
        testPtr.commands[i][CMD_ACTION] = 0
    }

    testPtr.hash = hash
    testPtr.time = 0

    return testPtr
}

// public
func (self *Testcase) BuildGraph(index int) error {

    redi := db.SingleRedi(nil)
    snapshots,err = redi.Diff()

	if err != nil {
		return err
	}

    // update weight
    createLen := len(snapshots[0])
    deleteLen := len(snapshots[1])

    if createLen && deleteLen {
        self.commands[index][CMD_ACTION] = CORPUS_FACTOR_MIX
    }else if createLen {
        self.commands[index][CMD_ACTION] = CORPUS_FACTOR_CREATE
    }else if deleteLen {
        self.commands[index][CMD_ACTION] = CORPUS_FACTOR_DELETE
    }else {
        self.commands[index][CMD_ACTION] = CORPUS_FACTOR_KEEP
    }

    // build graph
    command := self.commands[index][CMD_TEXT]
    self.graph[index].Build(snapshots,command)

    return nil
}

// public 
func (self *Testcase) Crash(index int) {

    self.commands[index][CMD_ACTION] = CORPUS_FACTOR_CRASH
}

// export
func NewCorpus() *Corpus {

    corpus := new(Corpus)
    corpus.hashset = make(map[string]bool,CORPUS_THRESHOLD)
    corpus.order = make([]*Testcase,CORPUS_THRESHOLD)

    return corpus
}

// public
// if exist return nil,err
// else     return ptr,nil
func (self *Corpus) AddSet(testcase string) *Testcase,error {

    // repeat testcase
    sum := md5.Sum([]byte(testcase))
    hash := string(sum)
    _,ok := hashset[hash]
    if ok {
        return nil,errors.New("Repeat Testcase.")
    }

    // new testcase
    hashset[hash] = true

    return NewTestcase(testcase,hash),nil

}

// public
func (self *Corpus) DropSet(testPtr *Testcase) {

    delete(self.hashset[testPtr.hash])
}

// public
func (self *Corpus) UpdateWeight(testPtr *Testcase) {

    len := len(testPtr.commands) * CORPUS_FACTOR_LEN
    actions := len(self.order) * CORPUS_FACTOR_COV

    for _,cmd := range testPtr.commands {
        actions += cmd[CMD_ACTION]
    }

    // calc weight 
    testPtr.weight = actions / len;

    // insert testPtr
    pos : = sort.Search(len(self.order),func(i int) bool { return self.order[i].weight >= testPtr.weight })
    self.order = append(self.order)
    copy(self.order[(pos+1):],self.order[pos:])
    self.order[pos] = testPtr

    // threshold
    if len(self.order) > CORPUS_THRESHOLD {
        self.order = self.order[1:]
    }
    
}

// public
func (self *Corpus) Select() *Testcase,int{

    // roulette wheel selection
    sumFloat := 0.0
    for _,testPtr := range self.order {
        sumFloat += testPtr.weight
    }

    randFloat := rand.Float64() * sum
    sumFloat = 0.0

    // selected testcase
    testSelect := nil
    for _,testPtr := range self.order {
        sumFloat += testPtr.weight
        if sumFloat > randFloat {
            testSelect = testPtr
            break
        }
    }

    sumInt := 0
    for _,command := range testSelect.commands {
        sumInt += command[CMD_ACTION]
    }

    randInt := rand.Intn(sumInt)
    sumInt = 0

    // selected command
    cmdSelect := 0
    for index,command := range testSelect.commands {
        sumInt += command[CMD_ACTION]
        if sumInt > randInt {
            cmdSelect = index
            break
        }
    }

    return testSelect,cmdSelect

}



