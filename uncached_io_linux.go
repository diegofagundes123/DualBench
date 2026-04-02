//go:build linux

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func openUncachedRW(path string, readOnly bool) (*os.File, error) {
	flag := unix.O_DIRECT
	if readOnly {
		return os.OpenFile(path, os.O_RDONLY|flag, 0)
	}
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC|flag, 0o644)
}

func minDirectIOAlignment() int {
	return 4096
}
