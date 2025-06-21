package model

import (
	"crypto/md5"
	"log"
	"math/rand"
	"strings"
	"time"
)

/*
 * Definition
 */

// corpus len
const (
	CORPUS_MINLEN int = 15
	CORPUS_MAXLEN int = 45
)

type Corpus struct {

	hashset map[string]bool
	order   []*Line
}


/*
 * Function
 */

// public
func NewCorpus() *Corpus {

	corpus := new(Corpus)
	corpus.hashset = make(map[string]bool, 0)
	corpus.order = make([]*Line, 0)

	return corpus
}

// public
func (self *Corpus) AddFile(file string) []*Line {

	ret := make([]*Line, 0)

	// split line
	lines := strings.Split(file, LineSep)

	for _, line := range lines {
		
		// md5
		sum := md5.Sum([]byte(line))
		hash := string(sum[:])

		_, ok := self.hashset[hash]

		// repeat line
		if ok {
			continue
		}

		// new line
		self.hashset[hash] = true
		new := NewLine(line, hash)

		self.order = append(self.order, new)
		ret = append(ret, new)
	}
	
	return ret
}

func (self *Corpus) Mutate() []*Line {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// mutated len
	length := r.Intn(CORPUS_MAXLEN-CORPUS_MINLEN) + CORPUS_MINLEN

	ret := make([]*Line, 0)

	for i := 0; i < length; {

		// select one line
		line := self.Select(r)

		if line == nil {
			continue
		}

		// repair line
		isRepaired := line.Repair(ret)

		if !isRepaired {
			continue
		}

		// mutate line
		line.Mutate(r)
		ret = append(ret, line)

		// one line is ready
		i ++
	}

	return ret
}

// public
func (self *Corpus) Select(r *rand.Rand) *Line {

	// roulette wheel selection
	sum := 0

	for _, line := range self.order {
		sum += line.Weight
	}

	rand := r.Int() * sum
	sum = 0

	// select line
	for _, line := range self.order {

		sum += line.Weight

		if sum > rand {
			return line
		}
	}

	return nil
}

// debug
func (self *Corpus) Debug() {

	log.Printf("Corpus Num: %d\n", len(self.order))

	for _, line := range self.order {
		line.Debug()
	}
}