package fuxx


type VertexType bool

const (
    cmdVertex VertexType = true
    metaVertex VertexType = false
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


func (self *Graph) Build(command string) error {

    redi := db.SingleRedi(nil)
    snapshots,err = redi.Diff()

	if err != nil {
		return err
	}

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
            keyV := self.AddVertex(metaVertex,key)
            keyMap[key] = keyV
            
            // edge (cmd, key)
            self.cmdV.Next = append(self.cmdV.Next,keyV)
        }
    }

    for _,pair := range snapshots[0] {
        key := pair.Key
        field := pair.Field

        // create field
        if field != nil {
            fieldV := self.AddVertex(metaVertex,field)

            // edge (key, field)
            if keyV := keyMap[key] {
                keyV.Next = append(keyV.Next,fieldV)

            // edge (cmd, field)
            }else {
                self.cmdV.Next = append(self.cmdV.Next,fieldV)
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
            keyV := self.AddVertex(metaVertex,key)
            keyMap[key] = keyV
            
            // edge (key, cmd)
            self.cmdV.Prev = append(self.cmdV.Prev,keyV)
        }
    }

    for _,pair := range snapshots[1] {
        key := pair.Key
        field := pair.Field

        // delete field
        if field != nil {
            fieldV := self.AddVertex(metaVertex,field)

            // edge (key, field)
            if keyV := keyMap[key] {
                keyV.Next = append(keyV.Next,fieldV)

            // edge (field, cmd)
            }else {
                self.cmdV.Prev = append(self.cmdV.Prev,fieldV)
            }
        }
    }

    // deal with keep
    clear(keyMap)

    for _,pair := range snapshots[2] {
        key := pair.Key
        field := pair.Field

        // keep key
        if field == nil && strings.Contains(command,key) {
            keyV := self.AddVertex(metaVertex,key)
            keyMap[key] = keyV
            
            // edge (key, cmd)
            self.cmdV.Prev = append(self.cmdV.Prev,keyV)
        }
    }

    for _,pair := range snapshots[2] {
        key := pair.Key
        field := pair.Field

        // keep field
        if field != nil && strings.Contains(command,field) {
            fieldV := self.AddVertex(metaVertex,field)

            // edge (key, field)
            if keyV := keyMap[key] {
                keyV.Next = append(keyV.Next,fieldV)

            // edge (field, cmd)
            }else {
                self.cmdV.Prev = append(self.cmdV.Prev,fieldV)
            }
        }
    }

    return nil

}
