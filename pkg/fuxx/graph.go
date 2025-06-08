package fuxx

import (
	"log"
	"math/rand"
	"strings"

	"github.com/fuxxcss/redi2fuxx/pkg/db"
)

type VertexType int

const (
	cmdVertex VertexType = iota
	keyVertex
	fieldVertex
)

type Graph struct {
	cmdV   *Vertex
	sliceV []*Vertex
}

type Vertex struct {
	vtype VertexType
	vdata string
	prev  []*Vertex
	next  []*Vertex
}

func NewGraph() *Graph {

	graph := new(Graph)
	graph.sliceV = make([]*Vertex, 0)

	return graph
}

func (self *Graph) AddVertex(isCmd VertexType, data string) *Vertex {

	vertex := new(Vertex)
	vertex.vtype = isCmd
	vertex.vdata = data
	self.sliceV = append(self.sliceV, vertex)

	return vertex
}

func (self *Graph) Contains(data string) bool {

	isContains := false
	for _, vertex := range self.sliceV {

		// contains data
		if vertex.vtype != cmdVertex && vertex.vdata == data {
			isContains = true
			break
		}
	}
	return isContains
}

func (self *Graph) Build(snapshots [3]db.Snapshot, command string) {

	// cmd vertex
	self.cmdV = self.AddVertex(cmdVertex, command)

	// deal with create
	keyMap := make(map[string]*Vertex, 0)

	for _, pair := range snapshots[0] {
		key := pair.Key
		field := pair.Field

		// create key
		if field == "" {
			keyV := self.AddVertex(keyVertex, key)
			keyMap[key] = keyV

			// edge (cmd, key)
			// cmd -> key
			self.cmdV.next = append(self.cmdV.next, keyV)
			keyV.prev = append(keyV.prev, self.cmdV)
		}
	}

	for _, pair := range snapshots[0] {
		key := pair.Key
		field := pair.Field

		// create field
		if field != "" {
			fieldV := self.AddVertex(fieldVertex, field)

			// edge (key, field)
			// cmd -> key -> field
			if keyV, ok := keyMap[key]; ok {
				keyV.next = append(keyV.next, fieldV)
				fieldV.prev = append(fieldV.prev, keyV)

				// edge (cmd, field), (key, cmd), (key, field)
				// key -> cmd -> field
				//  '-------------^
			} else {
				self.cmdV.next = append(self.cmdV.next, fieldV)
				fieldV.prev = append(fieldV.prev, self.cmdV)

				keyV := self.AddVertex(keyVertex, key)
				keyMap[key] = keyV

				keyV.next = append(keyV.next, self.cmdV)
				self.cmdV.prev = append(self.cmdV.prev, keyV)

				keyV.next = append(keyV.next, fieldV)
				fieldV.prev = append(fieldV.prev, keyV)
			}
		}
	}

	// deal with delete
	clear(keyMap)

	for _, pair := range snapshots[1] {
		key := pair.Key
		field := pair.Field

		// delete key
		if field == "" {
			keyV := self.AddVertex(keyVertex, key)
			keyMap[key] = keyV

			// edge (key, cmd)
			// key -> cmd
			self.cmdV.prev = append(self.cmdV.prev, keyV)
			keyV.next = append(keyV.next, self.cmdV)
		}
	}

	for _, pair := range snapshots[1] {
		key := pair.Key
		field := pair.Field

		// delete field
		if field != "" {
			fieldV := self.AddVertex(fieldVertex, field)

			// edge (key, field)
			// field <- key -> cmd
			if keyV, ok := keyMap[key]; ok {
				keyV.next = append(keyV.next, fieldV)
				fieldV.prev = append(fieldV.prev, keyV)

				// edge (field, cmd), (key, field)
				// key -> field -> cmd
			} else {
				self.cmdV.prev = append(self.cmdV.prev, fieldV)
				fieldV.next = append(fieldV.next, self.cmdV)

				keyV := self.AddVertex(keyVertex, key)
				keyMap[key] = keyV

				keyV.next = append(keyV.next, fieldV)
				fieldV.prev = append(fieldV.prev, keyV)
			}
		}
	}

	// deal with keep
	clear(keyMap)

	for _, pair := range snapshots[2] {
		key := pair.Key
		field := pair.Field

		// keep key
		if field == "" && strings.Contains(command, key) && !self.Contains(key) {
			keyV := self.AddVertex(keyVertex, key)
			keyMap[key] = keyV

			// edge (key, cmd)
			// key -> cmd
			self.cmdV.prev = append(self.cmdV.prev, keyV)
			keyV.next = append(keyV.next, self.cmdV)
		}
	}

	for _, pair := range snapshots[2] {
		key := pair.Key
		field := pair.Field

		// keep field
		if field != "" && strings.Contains(command, field) && !self.Contains(field) {
			fieldV := self.AddVertex(fieldVertex, field)

			// edge (key, field)
			// field <- key -> cmd
			if keyV, ok := keyMap[key]; ok {
				keyV.next = append(keyV.next, fieldV)
				fieldV.prev = append(fieldV.prev, keyV)

				// edge (field, cmd), (key, field)
				// key -> field -> cmd
			} else {
				self.cmdV.prev = append(self.cmdV.prev, fieldV)
				fieldV.next = append(fieldV.next, self.cmdV)

				keyV := self.AddVertex(keyVertex, key)
				keyMap[key] = keyV

				keyV.next = append(keyV.next, fieldV)
				fieldV.prev = append(fieldV.prev, keyV)
			}
		}
	}

}

// public
func (self *Graph) Match(graph *Graph) bool {

	/*
	log.Println("before match")
	self.Debug()
	graph.Debug()
	*/
	// select all keys
	matchKeys := make(map[*Vertex]int, 0)
	hasKeys := make(map[*Vertex]int, 0)

	// fieldVertex.prev must be one key
	for _, matchV := range self.cmdV.prev {

		// key to match
		if matchV.vtype == fieldVertex {

			matchV = matchV.prev[0]

			// alreay selected
			if _, ok := matchKeys[matchV]; ok {
				continue
			}
		}

		// field num
		nextSize := 0
		for _, matchNext := range matchV.next {

			if matchNext.vtype == fieldVertex {
				nextSize++
			}
		}

		matchKeys[matchV] = nextSize
	}

	// fieldVertex.prev must be one key
	for _, hasV := range graph.cmdV.next {

		// key has
		if hasV.vtype == fieldVertex {

			hasV = hasV.prev[0]

			// alreay selected
			if _, ok := hasKeys[hasV]; ok {
				continue
			}
		}

		// field num
		nextSize := 0
		for _, hasNext := range hasV.next {

			if hasNext.vtype == fieldVertex {
				nextSize++
			}
		}

		hasKeys[hasV] = nextSize
	}

	// match self -> graph
	for match, matchSize := range matchKeys {

		isMatched := false

		for has, hasSize := range hasKeys {

			// match succeed
			if matchSize <= hasSize {

				isMatched = true

				// need to patch
				self.cmdV.vdata = strings.Replace(self.cmdV.vdata, match.vdata, has.vdata, -1)
				match.vdata = has.vdata

				for i := 0; i < matchSize; i++ {

					self.cmdV.vdata = strings.Replace(self.cmdV.vdata, match.next[i].vdata, has.next[0].vdata, -1)
					match.next[i].vdata = has.next[0].vdata
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

func MutateStr(r *rand.Rand, str string) string {

	item := r.Intn(len(InterestingStr))
	chosen := InterestingStr[item]

	switch item {

	// empty
	case InterestEmpty:
		return chosen

	// null, terminal, hex, short str
	case InterestNULL, InterestTerminal, InterestHex, InterestLong:
		return str + chosen

	// special
	case InterestSpecial:
		special := r.Intn(len(chosen))
		return str + string(chosen[special])
	}

	return ""

}

// key, field mutate
func (self *Graph) Mutate(r *rand.Rand) {

	len := len(self.sliceV)

	for i := 0; i <= len / 2; i ++ {

		index := r.Intn(len)
		vertex := self.sliceV[index]

		// only mutate key, field, avoid token mutate
		if vertex.vtype != cmdVertex && !strings.Contains(vertex.vdata, db.RediStrSep) {

			mutatedData := MutateStr(r, vertex.vdata)
			self.cmdV.vdata = strings.Replace(self.cmdV.vdata, vertex.vdata, mutatedData, -1)
			self.sliceV[index].vdata = mutatedData
		}
	}

}

// debug
func (self *Graph) Debug() {

	log.Println("cmdV type",self.cmdV.vtype)
	log.Println("cmdV data",self.cmdV.vdata)

	log.Println("cmdV prev")
	for _, p := range self.cmdV.prev {
		log.Println("type", p.vtype, p.vdata,"size", len(p.vdata), "-> cmdV")
	}

	log.Println("cmdV next")
	for _, n := range self.cmdV.next {
		log.Println("cmdV ->","type", n.vtype, n.vdata, "size", len(n.vdata))
	}

	log.Println("all vertexs")
	for _,v := range self.sliceV {
		if v.vtype != cmdVertex {
			log.Println("vertex type",v.vtype, v.vdata, "size", len(v.vdata))
			for _, n := range v.next {
				log.Println(v.vdata," ->", n.vdata, "size", len(n.vdata))
			}
		}
	}
}
