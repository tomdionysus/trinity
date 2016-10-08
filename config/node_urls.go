package config

import (
	"fmt"
)

type NodeURLs []string

func (nurl *NodeURLs) Set(value string) error {
	*nurl = append(*nurl, value)
	return nil
}

func (nurl *NodeURLs) String() string {
	return fmt.Sprint(*nurl)
}
