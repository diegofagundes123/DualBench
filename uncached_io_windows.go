//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// openUncachedRW cria ou trunca arquivo com acesso direto ao disco, contornando o cache
// do sistema (FILE_FLAG_NO_BUFFERING + FILE_FLAG_WRITE_THROUGH na escrita).
func openUncachedRW(path string, readOnly bool) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var access uint32 = windows.GENERIC_READ | windows.GENERIC_WRITE
	var disposition uint32 = windows.OPEN_EXISTING
	var flags uint32 = windows.FILE_FLAG_NO_BUFFERING

	if readOnly {
		access = windows.GENERIC_READ
	} else {
		disposition = windows.CREATE_ALWAYS
		flags |= windows.FILE_FLAG_WRITE_THROUGH
	}

	h, err := windows.CreateFile(
		name,
		access,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		disposition,
		flags,
		0,
	)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(h), path), nil
}

func minDirectIOAlignment() int {
	// NTFS / exFAT em pendrives: 4096 é seguro para alinhamento com NO_BUFFERING.
	return 4096
}
