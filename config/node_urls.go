package config

import (
	"fmt"
)

// NodeURLs Hold all the discovered nodes url as a list
type NodeURLs []string

// Set a new url to the list of nodes url
func (nurl *NodeURLs) Set(value string) error {
	*nurl = append(*nurl, value)
	return nil
}

func (nurl *NodeURLs) String() string {
	return fmt.Sprint(*nurl)
}
