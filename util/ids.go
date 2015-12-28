package util

import(
  "math/rand"
  "time"
)

func GetRandomID() [16]byte {
  rand.Seed(time.Now().UTC().UnixNano())
  b := [16]byte{}
  for i:=0; i<16; i++ {
    b[i] = byte(rand.Intn(256))
  }
  return b
}