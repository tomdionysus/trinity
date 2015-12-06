package util

import(
  "math/rand"
  "time"
)

func GetRandomID(size int) []byte {
  rand.Seed(time.Now().UTC().UnixNano())
  b := []byte{}
  for i:=0; i<size; i++ {
    b = append(b, byte(rand.Intn(256)))
  }
  return b
}