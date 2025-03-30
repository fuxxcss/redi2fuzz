package fuxx

import (
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
    CORPUS_FACTOR_CRASH int = 100
    
    CORPUS_FACTOR_KEEP int = 0
    CORPUS_FACTOR_CREATE int = 1
    CORPUS_FACTOR_DELETE int = 1
    CORPUS_FACTOR_MIX int = 2
    CORPUS_FACTOR_FAULT int = 2
)

type Corpus struct {

    hashset map[string]bool
    order []*Testcase
}

// export
func NewTestcase(testcase string) *Testcase {

    test := new(Testcase)

    // redis split 
    slice := strings.Split(testcase,db.RediSep)
    test.commands = make([]Command,len(slice))

    for i,cmd := range slice {
        
        test.commands[i] = make(map[string]int,3)
        // text 
        test.commands[i][CMD_TEXT] = cmd
        // time 
        test.commands[i][CMD_TIME] = 0
        // action 
        test.commands[i][CMD_ACTION] = 0
    }

    return test
}

// public
func (self *Testcase) BuildGraph() error {

    db.

}

// public
func (self *Testcase) FaultWeight(index int) {

    self.commands[index][CMD_ACTION] -= CORPUS_FACTOR_BAD
}

// public
func (self *Testcase) UpdateWeight() {

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
// else     return ptr
func (self *Corpus) Exist(testcase string) *Testcase,error {

    // repeat testcase
    sum := md5.Sum([]byte(testcase))
    hash := string(sum)
    _,ok := hashset[hash]
    if ok {
        return nil,errors.New("Repeat Testcase.")
    }

    // new testcase
    hashset[testcase] = true

    return NewTestcase(testcase),nil

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

