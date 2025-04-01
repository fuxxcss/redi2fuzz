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
    graph *Graph
    weight float
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
        // time 
        testPtr.commands[i][CMD_TIME] = 0
        // action 
        testPtr.commands[i][CMD_ACTION] = 0
    }

    testPtr.hash = hash

    return testPtr
}

// public
func (self *Testcase) BuildGraph(index int) error {

    redi := db.SingleRedi(nil)
    create_delete,err = redi.Diff()

	if err != nil {
		return err
	}

    ... todo

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



func (self *Corpus) Select(){}


func match(meta1,meta2 VertexP) (bool,map[string]string) {
    // replace metadata name
    replace := make(map[string]string,0)
    // different metadata type，match failed
    if meta1.V_meta.M_type != meta2.V_meta.M_type { return false,nil }
    // range its child meta
    for edge1 := meta1.V_out ; edge1 != nil ; edge1 = edge1.E_from_next {
        // keep edge meta->meta
        if edge1.E_type != ETYPE_m_to_m { continue }
        matched := false
        // every child in meta1 can be matched in meta2
        for edge2 := meta2.V_out ; edge2 != nil ; edge2 = edge2.E_from_next {
            if edge2.E_type != ETYPE_m_to_m { continue }
            if matched,rep := match(edge1.E_to,edge2.E_to) ; matched {
                // collect
                for key,value := range rep { 
                    replace[key] = value 
                } 
                break
            }
        }
        if matched == false { return false,nil }
    }
    // same metadata type , collect
    replace[meta1.V_meta.M_name] = meta2.V_meta.M_name
    return true,replace
}

func repair(stmt_used,stmt_touse VertexP,line string) bool {
    // replace metadata name
    replace := make(map[string]string,0)
    // parent metadata
    meta_touse := make([]VertexP,0)
    meta_used := make([]VertexP,0)
    // range edge to stmt_touse，must be etype_m_to_s
    for edge := stmt_touse.V_in ; edge != nil ; edge = edge.E_to_next {
        // parent metadata
        metadata := edge.E_from
        if metadata.V_meta.M_parent != "" { continue }
        meta_touse = append(meta_touse,metadata)
    }
    // range edge from stmt_used，must be etype_s_to_m
    for edge := stmt_used.V_out ; edge != nil ; edge = edge.E_from_next {
        // parent metadata
        metadata := edge.E_to
        if metadata.V_meta.M_parent != "" { continue }
        meta_used = append(meta_used,metadata)
    }
    // range match()
    for _,touse := range meta_touse {
        matched := false
        for _,used := range meta_used {
            if matched,rep := match(touse,used) ; matched {
                // collect
                for key,value := range rep { 
                    replace[key] = value 
                }
                break
            }
        }
        if matched == false { return false}
    }
    // replace key->value
    for key,value := range replace { 
        strings.Replace(line,key,value,-1) 
    }
    return true
}

// RPC
func (self *Mutator) Mutate(arg int,reply *int) error {
    rand.Seed(time.Now().Unix())
    corpus_num := self.Corpus_num
    average_len := self.Average_len
    var chosen string
    var offset,index int
    var graph *Graph
    //stmt tmp，used for repair()
    stmts := make([]VertexP,0)
    // generate testcase ,len = average_len
    for len := average_len ; len > 0 ; len -- {
        // cyclic corpus
        if average_len > corpus_num {
            offset = len % corpus_num 
            if offset == 0 { offset = corpus_num }
            offset --

        // choose from first
        }else {
            offset = 0
        }
        // chosen lines text
        chosen = self.Corpus[offset]
        //fmt.Println(offset,chosen)
        // lines's Graph
        graph = self.Corpus_graph[offset]
        //random choose one line from lines
    rand_choose:
        index = rand.Intn(self.Corpus_len[offset])
        line := self.one_line(chosen,index)
        line_stmt := graph.Stmts[index]
        // repair
        if line_stmt.V_in_num == 0 {
            // line has no parent
            goto repaired
        }else {
            // range find repairable
            ret := false
            for _,stmt := range stmts {
                ret = repair(stmt,line_stmt,line) 
                if ret == true { goto repaired }
            }
            // not repairable , choose again
            if ret == false { goto rand_choose }
        }
        // repairable
    repaired:
        self.Testcase += line
        stmts = append(stmts,line_stmt)
    }
    *reply = len(self.Testcase)
    return nil
}

