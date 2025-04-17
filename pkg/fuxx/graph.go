package fuxx


type VertexType int

const (
    cmdVertex VertexType = iota
    keyVertex 
    fieldVertex
)

type Graph struct {

    cmdV *Vertex
    sliceV []Vertex
}

type Vertex struct {

    Type VertexType
    Data string
    Prev []*Vertex
    Next []*Vertex
}

func (self *Graph) AddVertex(isCmd VertexType,data string) *Vertex{

    vertex := new(Vertex)
    vertex.Type = isCmd
    vertex.Data = data
    self.sliceV = append(self.sliceV,vertex)

    return vertex
}

func (self *Graph) Contains(data string) bool {

    isContains := false
    for vertex := range self.sliceV {

        // contains data
        if vertex.Type != cmdVertex && vertex.Data == data {
            isContains = true
            break
        }
    }
    return isContains
}


func (self *Graph) Build(snapshots [3]Snapshot,command string) {

    // cmd vertex
    cmdV := self.AddVertex(cmdVertex,nil)
    self.cmdV = cmdV

    // deal with create
    keyMap := make(map[string]*Vertex,1)

    for _,pair := range snapshots[0] {
        key := pair.Key
        field := pair.Field

        // create key
        if field == nil {
            keyV := self.AddVertex(keyVertex,key)
            keyMap[key] = keyV
            
            // edge (cmd, key)
            // cmd -> key
            self.cmdV.Next = append(self.cmdV.Next,keyV)
            keyV.Prev = append(keyV.Prev,self.cmdV)
        }
    }

    for _,pair := range snapshots[0] {
        key := pair.Key
        field := pair.Field

        // create field
        if field != nil {
            fieldV := self.AddVertex(fieldVertex,field)

            // edge (key, field)
            // cmd -> key -> field
            if keyV := keyMap[key] {
                keyV.Next = append(keyV.Next,fieldV)
                fieldV.Prev = append(fieldV.Prev,keyV)

            // edge (cmd, field), (key, cmd), (key, field)
            // key -> cmd -> field
            //  '-------------^
            }else {
                self.cmdV.Next = append(self.cmdV.Next,fieldV)
                fieldV.Prev = append(fieldV.Prev,self.cmdV)

                keyV := self.AddVertex(keyVertex,key)
                keyMap[key] = keyV

                keyV.Next = append(keyV.Next,self.cmdV)
                self.cmdV.Prev = append(self.cmdV.Prev,keyV)

                keyV.Next = append(keyV.Next,fieldV)
                fieldV.Prev = append(fieldV.Prev,keyV)
            }
        }
    }

    // deal with delete
    clear(keyMap)

    for _,pair := range snapshots[1] {
        key := pair.Key
        field := pair.Field

        // delete key
        if field == nil {
            keyV := self.AddVertex(keyVertex,key)
            keyMap[key] = keyV
            
            // edge (key, cmd)
            // key -> cmd
            self.cmdV.Prev = append(self.cmdV.Prev,keyV)
            keyV.Next = append(keyV.Next,self.cmdV)
        }
    }

    for _,pair := range snapshots[1] {
        key := pair.Key
        field := pair.Field

        // delete field
        if field != nil {
            fieldV := self.AddVertex(fieldVertex,field)

            // edge (key, field)
            // field <- key -> cmd
            if keyV := keyMap[key] {
                keyV.Next = append(keyV.Next,fieldV)
                fieldV.Prev = append(fieldV.Prev,keyV)

            // edge (field, cmd), (key, field)
            // key -> field -> cmd
            }else {
                self.cmdV.Prev = append(self.cmdV.Prev,fieldV)
                fieldV.Next = append(fieldV.Next,self.cmdV)

                keyV := self.AddVertex(keyVertex,key)
                keyMap[key] = keyV

                keyV.Next = append(keyV.Next,fieldV)
                fieldV.Prev = append(fieldV.Prev,keyV)
            }
        }
    }

    // deal with keep
    clear(keyMap)

    for _,pair := range snapshots[2] {
        key := pair.Key
        field := pair.Field

        // keep key
        if field == nil && strings.Contains(command,key) && !self.Contains(key) {
            keyV := self.AddVertex(keyVertex,key)
            keyMap[key] = keyV
            
            // edge (key, cmd)
            // key -> cmd
            self.cmdV.Prev = append(self.cmdV.Prev,keyV)
            keyV.Next = append(keyV.Next,self.cmdV)
        }
    }

    for _,pair := range snapshots[2] {
        key := pair.Key
        field := pair.Field

        // keep field
        if field != nil && strings.Contains(command,field) && !self.Contains(field) {
            fieldV := self.AddVertex(fieldVertex,field)

            // edge (key, field)
            // field <- key -> cmd
            if keyV := keyMap[key] {
                keyV.Next = append(keyV.Next,fieldV)
                fieldV.Prev = append(fieldV.Prev,keyV)

            // edge (field, cmd), (key, field)
            // key -> field -> cmd 
            }else {
                self.cmdV.Prev = append(self.cmdV.Prev,fieldV)
                fieldV.Next = append(fieldV.Next,self.cmdV)

                keyV := self.AddVertex(keyVertex,key)
                keyMap[key] = keyV

                keyV.Next = append(keyV.Next,fieldV)
                fieldV.Prev = append(fieldV.Prev,keyV)
            }
        }
    }

}

// public
func (self *Graph) Match(graph *Graph) string,bool {

    // select all keys
    matchKeys := make([]*Vertex,1)
    hasKeys := make([]*Vertex,1)

    // fieldVertex.Prev must be one key
    for _,matchV := range self.cmdV.Prev {

        if matchV.Type == fieldVertex {

            matchV = matchV.Prev[0]

            // alreay selected
            if matchKeys.Contains(matchV) {
                continue
            }
        }
        matchKeys = append(matchKeys,matchV)
    }
    
    for _,hasV := range graph.cmdV.Next {

        if hasV.Type == fieldVertex {

            hasV = hasV.Prev[0]

            // alreay selected
            if hasKeys.Contains(hasV) {
                continue
            }
        }
        hasKeys = append(hasKeys,hasV)
    }
    
    // match self -> graph
    for _,match := range matchKeys {

        // match len
        matchLen := 0
        for _,matchNext := range match.Next {
            if matchNext.Type == fieldVertex {
                ++ matchLen
            }
        }

        for _,has := range hasKeys {

            isMatched := false
            // has len
            hasLen := 0
            for _,hasNext := range has.Next {
                if hasNext.Type == fieldVertex {
                    ++ hasLen
                }
            }

            // match succeed
            if matchLen <= hasLen {

                isMatched = true

                // need to patch
                self.cmdV.Data = strings.Replace(self.cmdV.Data,match.Data,has.Data,-1)
                match.Data = has.Data

                for i := 0 ; i < matchLen ; ++ i {

                    self.cmdV.Data = strings.Replace(self.cmdV.Data,match.Next[i].Data,has.Next[1].Data,-1)
                    match.Next[i].Data,has.Next[1].Data
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

func MutateStr(str string) string {

    item := rand.Intn(len(InterestingStr))
    chosen := InterestingStr[item]

    switch item {

    // empty
    case InterestEmpty:
        return chosen

    // null, terminal, hex, short str
    case InterestNULL,InterestTerminal,InterestHex,InterestShort:
        return append(str,chosen)

    // special
    case InterestSpecial:
        special = rand.Intn(len(chosen))
        return append(str,chosen[special])
    }

}

// key, field mutate
func (self *Graph) MutateGraph() {

    len := len(self.sliceV)

    for i := 0; i <= len ; i *= 2 {

        index := rand.Intn(len)
        vertex := self.sliceV[index]
    
        // only mutate key, field, avoid token mutate
        if vertex.Type != cmdVertex && !strings.Contains(vertex.Data,db.RediStrSep){
            
            mutatedData := MutateStr(vertex.Data)
            self.cmdV.Data = strings.Replace(self.cmdV.Data,vertex.Data,mutatedData,-1)
            vertex.Data = mutatedData
        }
    }

}

// str, int mutate
func MutateToken(cmdStr string) string {

    sliceToken := strings.Split(self.cmdV.Data,db.RediTokenSep)

    for i,token := range sliceToken {

        // int mutate
        _,err := strconv.Atoi(token)

        if err == nil {

            chosen := rand.Intn(len(InterestingInt))
            sliceToken[i] = InterestingInt[chosen]
        }

        // str mutate
        sliceStr := strings.Split(token,db.RediStrSep)

        if len(sliceStr) >= 3 {

            item := rand.Intn(len(InterestingStr))

            sliceStr[1] = MutateStr(item,sliceStr[1])

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


