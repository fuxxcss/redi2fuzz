package db

import (
	"context"
	"errors"
	"log"
	"os/exec"
	"slices"
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
type Snapshot []*RediPair

// global
var (
	globalRedi *Redi
	mutexRedi  sync.Mutex
)

type Redi struct {
	Proc     *exec.Cmd
	snapshot Snapshot
	client   *redis.Client
	ctx      context.Context
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
	redi.snapshot = make(Snapshot, 1)
	redi.ctx = context.Background()

	return redi
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

	return nil
}

// public
func (self *Redi) Execute(args []string) string {

	state := utils.STATE_OK
	_, err := self.client.Do(self.ctx, args).Result()

	// execute failed
	if err != nil && err != redis.Nil {
		log.Println(err)
		// execute error
		if self.CheckAlive() {
			state = utils.STATE_BAD

			// crash
		} else {
			state = utils.STATE_ERR
		}
	}

	return state

}

// public
// [0] create, [1] delete, [2] others
func (self *Redi) Diff() ([3]Snapshot, error) {

	var ret [3]Snapshot

	new := make(Snapshot, 1)
	err := self.collect(new)

	if err != nil {
		return ret, err
	}

	old := self.snapshot

	for index, pair := range new {

		// create pair
		if !slices.Contains(old, pair) {
			ret[0] = append(ret[0], pair)
			// keep others
			new = slices.Delete(new, index, index+1)
		}
	}

	for _, pair := range old {

		// delete pair
		if !slices.Contains(new, pair) {
			ret[1] = append(ret[1], pair)
		}
	}

	ret[2] = new

	return ret, nil

}

// private
func (self *Redi) collect(snapshot Snapshot) error {

	keys, err := self.client.Keys(self.ctx, "*").Result()

	// redis query engine, type = "none"
	ft, err := self.client.Do(self.ctx, "FT._LIST").Text()
	keys = append(keys, ft)

	// Keys failed
	if err != nil {
		return errors.New("KEYS * failed.")
	}

	// keys
	for _, key := range keys {

		pair := new(RediPair)
		pair.Key = key
		pair.Field = ""
		snapshot = append(snapshot, pair)

		keyType, err := self.client.Type(self.ctx, key).Result()

		// Type failed
		if err != nil {
			return errors.New("TYPE key failed.")
		}

		// func map
		fmap := map[string]func(string, Snapshot) error{
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

func (self *Redi) collectHash(key string, snapshot Snapshot) error {

	fields, err := self.client.HKeys(self.ctx, key).Result()

	// HKEYS failed
	if err != nil {
		return errors.New("collect hash failed.")
	}

	for _, field := range fields {
		pair := new(RediPair)
		pair.Key = key
		pair.Field = field
		snapshot = append(snapshot, pair)
	}

	return nil

}

func (self *Redi) collectStream(key string, snapshot Snapshot) error {

	entries, err := self.client.XRange(self.ctx, key, "-", "+").Result()

	if err != nil {
		return errors.New("collect stream failed.")
	}

	for _, entry := range entries {
		for field := range entry.Values {
			pair := new(RediPair)
			pair.Key = key
			pair.Field = field
			snapshot = append(snapshot, pair)
		}
	}

	return nil
}
