package consistenthash

import(
  "bytes"
  bt "github.com/tomdionysus/binarytree"
)

type Key [16]byte

// Return true if this key is less than the supplied ByteSliceKey.
func (me Key) LessThan(other bt.Comparable) bool {
  return bytes.Compare(me[:], getSlice(other)) < 0
} 

// Return true if this key is equal to the supplied ByteSliceKey.
func (me Key) EqualTo(other bt.Comparable) bool {
  return bytes.Compare(me[:], getSlice(other)) == 0
} 

// Return true if this key is greater than the supplied ByteSliceKey.
func (me Key) GreaterThan(other bt.Comparable) bool {
  return bytes.Compare(me[:], getSlice(other)) > 0
} 

func (me Key) ValueOf() interface{} {
  return [16]byte(me)
}

func getSlice(t bt.Comparable) []byte {
  x := t.ValueOf().([16]byte)
  return x[:]
}

