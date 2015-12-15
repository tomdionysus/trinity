package kvstore

// In Memory Key Value Store for testing.
type KVStore struct {
  store map[string][]byte
}

func NewKVStore() *KVStore {
  inst := &KVStore{
    store: map[string][]byte{},
  }
  return inst
}

func (me *KVStore) Set(key string, value []byte) {
  me.store[key] = value
}

func (me *KVStore) IsSet(key string) bool {
  _, isset := me.Get(key)
  return isset
}

func (me *KVStore) Get(key string) ([]byte, bool) {
  val, ok := me.store[key]
  return val, ok
}

func (me *KVStore) Delete(key string) bool {
  _, ok := me.store[key]
  delete(me.store,key)
  return ok
}