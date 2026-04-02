//go:build darwin

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

// macOS não expõe O_DIRECT como no Linux; F_NOCACHE desativa cache de página para o fd.
func openUncachedRW(path string, readOnly bool) (*os.File, error) {
	var f *os.File
	var err error
	if readOnly {
		f, err = os.OpenFile(path, os.O_RDONLY, 0)
	} else {
		f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	}
	if err != nil {
		return nil, err
	}
	raw, err := unix.FcntlInt(uintptr(f.Fd()), unix.F_NOCACHE, 1)
	if err != nil {
		f.Close()
		return nil, err
	}
	_ = raw
	return f, nil
}

func minDirectIOAlignment() int {
	return 4096
}
