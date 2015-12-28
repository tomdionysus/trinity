package kvstore

import(
  "github.com/tomdionysus/trinity/util"
  "time"
)

// In Memory Key Value Store for testing.
type KVStore struct {
  Logger *util.Logger
  store map[string][]byte
  flags map[string]int16
  expiry map[int64][]string
  running bool
}

func NewKVStore(logger *util.Logger) *KVStore {
  inst := &KVStore{
    Logger: logger,
    store: map[string][]byte{},
    flags: map[string]int16{},
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
        me.Logger.Debug("KVStore","Expiring Time %d", expiretime)
        for _, k := range toexpire { me.Delete(k) }
        delete(me.expiry, expiretime)
      }
      time.Sleep(900*time.Millisecond)
    }
  }()
}

func (me *KVStore) Stop() {
  me.running = false
}

func (me *KVStore) Set(key string, value []byte, flags int16, expiry *time.Time) {
  me.store[key] = value
  me.flags[key] = flags
  if expiry!=nil {
    exptime := expiry.UTC().Unix()
    me.Logger.Debug("KVStore","SET [%s] %s - Expiry %d", key, value, exptime)
    me.expiry[exptime] = append(me.expiry[exptime], key)
  } else {
    me.Logger.Debug("KVStore","SET [%s] %s", key, value)
  }
}

func (me *KVStore) IsSet(key string) bool {
  _, _, isset := me.Get(key)
  return isset
}

func (me *KVStore) Get(key string) ([]byte, int16, bool) {
  value, ok := me.store[key]
  if ok {
    flags, ok := me.flags[key]
    me.Logger.Debug("KVStore","GET [%s] %s", key, value)
    return value, flags, ok
  } else {
    me.Logger.Debug("KVStore","GET [%s] NOT FOUND", key)
    return nil, 0, false
  }
}

func (me *KVStore) Delete(key string) bool {
  _, ok := me.store[key]
  me.Logger.Debug("KVStore","DELETE [%s]", key)
  delete(me.store,key)
  delete(me.flags,key)
  return ok
}