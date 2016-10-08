package config

import (
	"fmt"
)

type NodeURLs []string

func (me *NodeURLs) Set(value string) error {
	*me = append(*me, value)
	return nil
}

func (i *NodeURLs) String() string {
	return fmt.Sprint(*i)
}
