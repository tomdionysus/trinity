package util

import (
	"strconv"
)

// MemcachedSetArgsHelper helps handleing argument for the set, add and replace memcached commands
func MemcachedSetArgsHelper(args []string) (int, int, int, error) {
	expirytime, err := strconv.Atoi(args[3])
	if err != nil {
		return 0, 0, 0, err
	}

	flags, err := strconv.Atoi(args[2])
	flags = flags & 0xFFFF
	if err != nil {
		return 0, 0, 0, err
	}

	bytes, err := strconv.Atoi(args[4])
	if err != nil {
		return 0, 0, 0, err
	}

	return expirytime, flags, bytes, nil
}
