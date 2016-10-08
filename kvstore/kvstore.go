package kvstore

import (
	bt "github.com/tomdionysus/binarytree"
	ch "github.com/tomdionysus/consistenthash"
	"github.com/tomdionysus/trinity/util"
	"time"
)

// In Memory Key Value Store for testing.
type KVStore struct {
	Logger  *util.Logger
	store   *bt.Tree
	expiry  map[int64][]ch.Key
	running bool
}

type Item struct {
	Key   string
	Data  []byte
	Flags int16
}

func NewKVStore(logger *util.Logger) *KVStore {
	inst := &KVStore{
		Logger:  logger,
		store:   bt.NewTree(),
		expiry:  map[int64][]ch.Key{},
		running: false,
	}
	return inst
}

func (me *KVStore) Init() {
	me.Logger.Debug("Memcache", "Init")
}

func (me *KVStore) Start() {
	me.running = true
	go func() {
		me.Logger.Debug("Memcache", "Started")
		for me.running {
			expiretime := time.Now().UTC().Unix()
			toexpire, found := me.expiry[expiretime]
			if found {
				me.Logger.Debug("KVStore", "Expiring Time %d", expiretime)
				for _, k := range toexpire {
					me.deleteKey(k)
				}
				delete(me.expiry, expiretime)
			}
			time.Sleep(900 * time.Millisecond)
		}
	}()
}

func (me *KVStore) Stop() {
	me.running = false
}

func (me *KVStore) Set(key string, value []byte, flags int16, expiry *time.Time) {
	keymd5 := ch.NewMD5Key(key)
	item := &Item{Key: key, Data: value, Flags: flags}
	me.store.Set(keymd5, item)

	if expiry != nil {
		exptime := expiry.UTC().Unix()
		me.Logger.Debug("KVStore", "SET [%s] %s - Expiry %d", key, value, exptime)
		me.expiry[exptime] = append(me.expiry[exptime], keymd5)
	} else {
		me.Logger.Debug("KVStore", "SET [%s] %s", key, value)
	}
}

func (me *KVStore) IsSet(key string) bool {
	_, _, isset := me.Get(key)
	return isset
}

func (me *KVStore) Get(key string) ([]byte, int16, bool) {
	keymd5 := ch.NewMD5Key(key)
	ok, valueInt := me.store.Get(keymd5)
	if ok {
		value := valueInt.(*Item)
		me.Logger.Debug("KVStore", "GET [%s] %s", value.Key, value.Data)
		return value.Data, value.Flags, true
	} else {
		me.Logger.Debug("KVStore", "GET [%s] NOT FOUND", key)
		return nil, 0, false
	}
}

func (me *KVStore) Delete(key string) bool {
	me.Logger.Debug("KVStore", "DELETE [%s]", key)
	return me.deleteKey(ch.NewMD5Key(key))
}

func (me *KVStore) deleteKey(key ch.Key) bool {
	me.store.Clear(key)
	return true
}
