package fuxx

import (
    "strings"
    "slices"
    "strconv"
    "math/rand"

    "github.com/fuxxcss/redi2fuxx/pkg/db"
)

type VertexType int

const (
    cmdVertex VertexType = iota
    keyVertex 
    fieldVertex
)

type Graph struct {

    cmdV *Vertex
    sliceV []*Vertex
}

type Vertex struct {

    vtype VertexType
    vdata string
    prev []*Vertex
    next []*Vertex
}

func (self *Graph) AddVertex(isCmd VertexType,data string) *Vertex{

    vertex := new(Vertex)
    vertex.vtype = isCmd
    vertex.vdata = data
    self.sliceV = append(self.sliceV,vertex)

    return vertex
}

func (self *Graph) Contains(data string) bool {

    isContains := false
    for _,vertex := range self.sliceV {

        // contains data
        if vertex.vtype != cmdVertex && vertex.vdata == data {
            isContains = true
            break
        }
    }
    return isContains
}


func (self *Graph) Build(snapshots [3]db.Snapshot,command string) {

    // cmd vertex
    self.cmdV = self.AddVertex(cmdVertex,command)

    // deal with create
    keyMap := make(map[string]*Vertex,1)

    for _,pair := range snapshots[0] {
        key := pair.Key
        field := pair.Field

        // create key
        if field == "" {
            keyV := self.AddVertex(keyVertex,key)
            keyMap[key] = keyV
            
            // edge (cmd, key)
            // cmd -> key
            self.cmdV.next = append(self.cmdV.next,keyV)
            keyV.prev = append(keyV.prev,self.cmdV)
        }
    }

    for _,pair := range snapshots[0] {
        key := pair.Key
        field := pair.Field

        // create field
        if field != "" {
            fieldV := self.AddVertex(fieldVertex,field)

            // edge (key, field)
            // cmd -> key -> field
            if keyV,ok := keyMap[key] ; ok {
                keyV.next = append(keyV.next,fieldV)
                fieldV.prev = append(fieldV.prev,keyV)

            // edge (cmd, field), (key, cmd), (key, field)
            // key -> cmd -> field
            //  '-------------^
            }else {
                self.cmdV.next = append(self.cmdV.next,fieldV)
                fieldV.prev = append(fieldV.prev,self.cmdV)

                keyV := self.AddVertex(keyVertex,key)
                keyMap[key] = keyV

                keyV.next = append(keyV.next,self.cmdV)
                self.cmdV.prev = append(self.cmdV.prev,keyV)

                keyV.next = append(keyV.next,fieldV)
                fieldV.prev = append(fieldV.prev,keyV)
            }
        }
    }

    // deal with delete
    clear(keyMap)

    for _,pair := range snapshots[1] {
        key := pair.Key
        field := pair.Field

        // delete key
        if field == "" {
            keyV := self.AddVertex(keyVertex,key)
            keyMap[key] = keyV
            
            // edge (key, cmd)
            // key -> cmd
            self.cmdV.prev = append(self.cmdV.prev,keyV)
            keyV.next = append(keyV.next,self.cmdV)
        }
    }

    for _,pair := range snapshots[1] {
        key := pair.Key
        field := pair.Field

        // delete field
        if field != "" {
            fieldV := self.AddVertex(fieldVertex,field)

            // edge (key, field)
            // field <- key -> cmd
            if keyV,ok := keyMap[key] ; ok {
                keyV.next = append(keyV.next,fieldV)
                fieldV.prev = append(fieldV.prev,keyV)

            // edge (field, cmd), (key, field)
            // key -> field -> cmd
            }else {
                self.cmdV.prev = append(self.cmdV.prev,fieldV)
                fieldV.next = append(fieldV.next,self.cmdV)

                keyV := self.AddVertex(keyVertex,key)
                keyMap[key] = keyV

                keyV.next = append(keyV.next,fieldV)
                fieldV.prev = append(fieldV.prev,keyV)
            }
        }
    }

    // deal with keep
    clear(keyMap)

    for _,pair := range snapshots[2] {
        key := pair.Key
        field := pair.Field

        // keep key
        if field == "" && strings.Contains(command,key) && !self.Contains(key) {
            keyV := self.AddVertex(keyVertex,key)
            keyMap[key] = keyV
            
            // edge (key, cmd)
            // key -> cmd
            self.cmdV.prev = append(self.cmdV.prev,keyV)
            keyV.next = append(keyV.next,self.cmdV)
        }
    }

    for _,pair := range snapshots[2] {
        key := pair.Key
        field := pair.Field

        // keep field
        if field != "" && strings.Contains(command,field) && !self.Contains(field) {
            fieldV := self.AddVertex(fieldVertex,field)

            // edge (key, field)
            // field <- key -> cmd
            if keyV,ok := keyMap[key] ; ok {
                keyV.next = append(keyV.next,fieldV)
                fieldV.prev = append(fieldV.prev,keyV)

            // edge (field, cmd), (key, field)
            // key -> field -> cmd 
            }else {
                self.cmdV.prev = append(self.cmdV.prev,fieldV)
                fieldV.next = append(fieldV.next,self.cmdV)

                keyV := self.AddVertex(keyVertex,key)
                keyMap[key] = keyV

                keyV.next = append(keyV.next,fieldV)
                fieldV.prev = append(fieldV.prev,keyV)
            }
        }
    }

}

// public
func (self *Graph) Match(graph *Graph) bool {

    // select all keys
    matchKeys := make([]*Vertex,1)
    hasKeys := make([]*Vertex,1)

    // fieldVertex.prev must be one key
    for _,matchV := range self.cmdV.prev {

        if matchV.vtype == fieldVertex {

            matchV = matchV.prev[0]

            // alreay selected
            if slices.Contains(matchKeys,matchV) {
                continue
            }
        }
        matchKeys = append(matchKeys,matchV)
    }
    
    for _,hasV := range graph.cmdV.next {

        if hasV.vtype == fieldVertex {

            hasV = hasV.prev[0]

            // alreay selected
            if slices.Contains(hasKeys,hasV) {
                continue
            }
        }
        hasKeys = append(hasKeys,hasV)
    }
    
    // match self -> graph
    for _,match := range matchKeys {

        // match len
        matchLen := 0
        for _,matchNext := range match.next {
            if matchNext.vtype == fieldVertex {
                matchLen ++
            }
        }

        isMatched := false

        for _,has := range hasKeys {

            // has len
            hasLen := 0
            for _,hasNext := range has.next {
                if hasNext.vtype == fieldVertex {
                    hasLen ++
                }
            }

            // match succeed
            if matchLen <= hasLen {

                isMatched = true

                // need to patch
                self.cmdV.vdata = strings.Replace(self.cmdV.vdata,match.vdata,has.vdata,-1)
                match.vdata = has.vdata

                for i := 0 ; i < matchLen ; i ++ {

                    self.cmdV.vdata = strings.Replace(self.cmdV.vdata,match.next[i].vdata,has.next[1].vdata,-1)
                    match.next[i].vdata = has.next[1].vdata
                }

                break
            }
            
        }

        // match failed
        if !isMatched {

            return false
        }
    }
    
    return true
}

func MutateStr(r *rand.Rand,str string) string {

    item := r.Intn(len(InterestingStr))
    chosen := InterestingStr[item]

    switch item {

    // empty
    case InterestEmpty:
        return chosen

    // null, terminal, hex, short str
    case InterestNULL,InterestTerminal,InterestHex,InterestShort:
        return str + chosen

    // special
    case InterestSpecial:
        special := r.Intn(len(chosen))
        return str + string(chosen[special])
    }

    return ""
    
}

// key, field mutate
func (self *Graph) MutateGraph(r *rand.Rand) {

    len := len(self.sliceV)

    for i := 0; i <= len ; i *= 2 {

        index := r.Intn(len)
        vertex := self.sliceV[index]
    
        // only mutate key, field, avoid token mutate
        if vertex.vtype != cmdVertex && !strings.Contains(vertex.vdata,db.RediStrSep){
            
            mutatedData := MutateStr(r,vertex.vdata)
            self.cmdV.vdata = strings.Replace(self.cmdV.vdata,vertex.vdata,mutatedData,-1)
            vertex.vdata = mutatedData
        }
    }

}

// str, int mutate
func MutateToken(r *rand.Rand,cmdStr string) string {

    sliceToken := strings.Split(cmdStr,db.RediTokenSep)

    for i,token := range sliceToken {

        // int mutate
        _,err := strconv.Atoi(token)

        if err == nil {

            chosen := r.Intn(len(InterestingInt))
            sliceToken[i] = InterestingInt[chosen]
        }

        // str mutate
        sliceStr := strings.Split(token,db.RediStrSep)

        if len(sliceStr) >= 3 {

            sliceStr[1] = MutateStr(r,sliceStr[1])

            mutatedStr := ""

            // assemble
            for _,str := range sliceStr {
                
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

    return mutatedToken

}


