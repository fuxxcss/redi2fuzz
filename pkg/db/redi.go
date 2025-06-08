package db

import (
	"context"
	"errors"
	"log"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"github.com/fuxxcss/redi2fuxx/pkg/utils"
	"github.com/redis/go-redis/v9"
)

// Redi strings
const (
	RediSep      string = "\n"
	RediTokenSep string = " "
	RediStrSep   string = "\""
	RediPort     string = "--port"
)

// Redi Pair
// different key can have same field name
type RediPair struct {
	Key   string
	Field string
}

// snapshot meta data
type Snapshot []RediPair

// global
var (
	globalRedi *Redi
	mutexRedi  sync.Mutex
)

type Redi struct {
	// proc
	path string
	args []string
	Proc *exec.Cmd
	// snapshot
	snapshot Snapshot
	// env
	client *redis.Client
	ctx    context.Context
}

// export
func SingleRedi(port string) *Redi {

	if globalRedi == nil {
		mutexRedi.Lock()
		defer mutexRedi.Unlock()
		if globalRedi == nil {
			globalRedi = NewRedi(port)
		}
	}
	return globalRedi
}

// export
func NewRedi(port string) *Redi {

	redi := new(Redi)

	// redi connect
	redi.client = redis.NewClient(&redis.Options{
		Addr:     "localhost:" + port,
		Password: "",
		DB:       0,
	})

	redi.Proc = nil
	redi.snapshot = make(Snapshot, 0)
	redi.ctx = context.Background()

	return redi
}

// public
func (self *Redi) Restart() {

	self.Proc = exec.Command(self.path, self.args...)
	err := self.Proc.Start()

	// db failed
	if err != nil {
		log.Fatalln("err: redi failed.")
	}

	// waiting redi startup
	for {
		alive := self.CheckAlive()
		if alive {
			break
		}
	}

	// db succeed
	log.Printf("[*] DB %v StartUp.\n", self.path)
}

// public
func (self *Redi) CheckAlive() bool {

	// redi state
	_, err := self.client.Ping(self.ctx).Result()

	// redi is not alive
	if err != nil {
		return false
	}

	return true

}

// public
func (self *Redi) CleanUp() error {

	_, err := self.client.FlushAll(self.ctx).Result()

	// flushall failed
	if err != nil {
		return err
	}

	self.snapshot = make(Snapshot, 0)

	return nil
}

// public
func (self *Redi) SplitLine(str string) []string {

	return strings.Split(str, RediSep)
}

// public
func (self *Redi) SplitToken(str string) []string {

	sliceToken := strings.Split(str, RediTokenSep)
	tokens := make([]string, 0)

	for _, token := range sliceToken {
		if token != "" {
			tokens = append(tokens, token)
		}
	}

	return tokens

}

// public
func (self *Redi) Execute(tokens []string) (string,error) {

	// marshal string
	args := []interface{}{}
	for _, token := range tokens {
		args = append(args, token)
	}

	state := utils.STATE_OK
	_, err := self.client.Do(self.ctx, args...).Result()

	// execute failed
	if err != nil && err != redis.Nil {

		// execute error
		if self.CheckAlive() {
			state = utils.STATE_BAD

			// crash
		} else {
			state = utils.STATE_ERR
		}
	}

	return state, err

}

// public
// [0] create, [1] delete, [2] others
func (self *Redi) Diff() ([3]Snapshot, error) {

	var ret [3]Snapshot

	// update new
	new := make(Snapshot, 0)
	err := self.collect(&new)

	if err != nil {
		return ret, err
	}

	old := self.snapshot.Copy()
	/*
	log.Println("old snapshot")
	old.Debug()
	*/
	// update old
	self.snapshot = new.Copy()
	/*
	log.Println("new snapshot")
	new.Debug()
	*/
	// loop create, keep
	cnt := 0
	for index, pair := range self.snapshot {

		// create pair
		if !old.Contains(pair) {
			ret[0] = append(ret[0], pair)

			// keep others, delete cnt ++
			new = slices.Delete(new, index-cnt, index-cnt+1)
			cnt++
		}
	}

	for _, pair := range old {

		// delete pair
		if !self.snapshot.Contains(pair) {
			ret[1] = append(ret[1], pair)
		}
	}

	ret[2] = new

	return ret, nil

}

// private
func (self *Redi) collect(snapshot *Snapshot) error {

	keys, _ := self.client.Keys(self.ctx, "*").Result()

	// redis stack query engine, type = "none"
	ft, err := self.client.Do(self.ctx, "FT._LIST").Text()

	if err == nil {
		keys = append(keys, ft)
	}

	// keys
	for _, key := range keys {

		var pair RediPair
		pair.Key = key
		pair.Field = ""
		*snapshot = append(*snapshot, pair)

		keyType, err := self.client.Type(self.ctx, key).Result()

		// Type failed
		if err != nil {
			return errors.New("TYPE key failed.")
		}

		// func map
		fmap := map[string]func(string, *Snapshot) error{
			"hash": self.collectHash,
			// "geo" : collect geo,
			"stream": self.collectStream,
			// "none" : collect ft,
			// "TSDB-TYPE" : collect ts,
		}

		f, ok := fmap[keyType]
		if ok {
			err := f(key, snapshot)

			// failed
			if err != nil {
				return err
			}
		}

	}
	return nil

}

func (self *Redi) collectHash(key string, snapshot *Snapshot) error {

	fields, err := self.client.HKeys(self.ctx, key).Result()

	// HKEYS failed
	if err != nil {
		return errors.New("collect hash failed.")
	}

	for _, field := range fields {
		var pair RediPair
		pair.Key = key
		pair.Field = field
		*snapshot = append(*snapshot, pair)
	}

	return nil

}

func (self *Redi) collectStream(key string, snapshot *Snapshot) error {

	entries, err := self.client.XRange(self.ctx, key, "-", "+").Result()

	if err != nil {
		return errors.New("collect stream failed.")
	}

	for _, entry := range entries {
		for field := range entry.Values {
			var pair RediPair
			pair.Key = key
			pair.Field = field
			*snapshot = append(*snapshot, pair)
		}
	}

	return nil
}

// public
func (self *Snapshot) Contains(pair RediPair) bool {

	isContains := false
	for _, p := range *self {

		// contains key, field
		if p.Key == pair.Key && p.Field == pair.Field {
			isContains = true
			break
		}
	}
	return isContains

}

// public
func (self *Snapshot) Copy() Snapshot {

	new := make(Snapshot, 0)
	for _, pair := range *self {
		new = append(new, pair)
	}
	return new

}

// debug
func (self *Snapshot) Debug() {

	for _, pair := range *self {
		k := pair.Key
		f := pair.Field
		log.Println("key", k, "size", len(k), " -> ", "field", f, "size", len(f))
	}
}

// debug
func DiffDebug(snapshots [3]Snapshot) {

	if len(snapshots[0]) > 0 {
		log.Println("create snapshot")
		snapshots[0].Debug()
	}

	if len(snapshots[1]) > 0 {
		log.Println("delete snapshot")
		snapshots[1].Debug()
	}

	if len(snapshots[2]) > 0 {
		log.Println("keep snapshot")
		snapshots[2].Debug()
	}
}


