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

func (kvs *KVStore) Init() {
	kvs.Logger.Debug("Memcache", "Init")
}

func (kvs *KVStore) Start() {
	kvs.running = true
	go func() {
		kvs.Logger.Debug("Memcache", "Started")
		for kvs.running {
			expiretime := time.Now().UTC().Unix()
			toexpire, found := kvs.expiry[expiretime]
			if found {
				kvs.Logger.Debug("KVStore", "Expiring Time %d", expiretime)
				for _, k := range toexpire {
					kvs.deleteKey(k)
				}
				delete(kvs.expiry, expiretime)
			}
			time.Sleep(900 * time.Millisecond)
		}
	}()
}

func (kvs *KVStore) Stop() {
	kvs.running = false
}

func (kvs *KVStore) Set(key string, value []byte, flags int16, expiry *time.Time) {
	keymd5 := ch.NewMD5Key(key)
	item := &Item{Key: key, Data: value, Flags: flags}
	kvs.store.Set(keymd5, item)

	if expiry != nil {
		exptime := expiry.UTC().Unix()
		kvs.Logger.Debug("KVStore", "SET [%s] %s - Expiry %d", key, value, exptime)
		kvs.expiry[exptime] = append(kvs.expiry[exptime], keymd5)
	} else {
		kvs.Logger.Debug("KVStore", "SET [%s] %s", key, value)
	}
}

func (kvs *KVStore) IsSet(key string) bool {
	_, _, isset := kvs.Get(key)
	return isset
}

func (kvs *KVStore) Get(key string) ([]byte, int16, bool) {
	keymd5 := ch.NewMD5Key(key)
	ok, valueInt := kvs.store.Get(keymd5)
	if ok {
		value := valueInt.(*Item)
		kvs.Logger.Debug("KVStore", "GET [%s] %s", value.Key, value.Data)
		return value.Data, value.Flags, true
	} else {
		kvs.Logger.Debug("KVStore", "GET [%s] NOT FOUND", key)
		return nil, 0, false
	}
}

func (kvs *KVStore) Delete(key string) bool {
	kvs.Logger.Debug("KVStore", "DELETE [%s]", key)
	return kvs.deleteKey(ch.NewMD5Key(key))
}

func (kvs *KVStore) deleteKey(key ch.Key) bool {
	kvs.store.Clear(key)
	return true
}
