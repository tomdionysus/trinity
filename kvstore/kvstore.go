package kvstore

import(
  "github.com/tomdionysus/trinity/util"
  "time"
)

// In Memory Key Value Store for testing.
type KVStore struct {
  Logger *util.Logger
  store map[string][]byte
  expiry map[int64][]string
  running bool
}

func NewKVStore(logger *util.Logger) *KVStore {
  inst := &KVStore{
    Logger: logger,
    store: map[string][]byte{},
    expiry: map[int64][]string{},
    running: false,
  }
  return inst
}

func (me *KVStore) Init() {
  me.Logger.Debug("Memcache","Init")
}

func (me *KVStore) Start() {
  me.running = true
  go func() {
    me.Logger.Debug("Memcache","Started")
    for me.running {
      expiretime := time.Now().UTC().Unix()
      toexpire, found := me.expiry[expiretime]
      if found {
        for _, k := range toexpire {
          me.Logger.Debug("KVStore","Expiring %s", k)
          delete(me.store,k)
        }
        delete(me.expiry, expiretime)
      }
      time.Sleep(900*time.Millisecond)
    }
  }()
}

func (me *KVStore) Stop() {
  me.running = false
}

func (me *KVStore) Set(key string, value []byte, expiry *time.Time) {
  me.store[key] = value
  if expiry!=nil {
    exptime := expiry.UTC().Unix()
    me.Logger.Debug("KVStore","SET [%s] - Expiry %d", key, exptime)
    me.expiry[exptime] = append(me.expiry[exptime], key)
  } else {
    me.Logger.Debug("KVStore","SET [%s]", key)
  }
}

func (me *KVStore) IsSet(key string) bool {
  _, isset := me.Get(key)
  return isset
}

func (me *KVStore) Get(key string) ([]byte, bool) {
  val, ok := me.store[key]
  me.Logger.Debug("KVStore","GET [%s]", key)
  return val, ok
}

func (me *KVStore) Delete(key string) bool {
  _, ok := me.store[key]
  me.Logger.Debug("KVStore","DELETE [%s]", key)
  delete(me.store,key)
  return ok
}