//go:build darwin || linux

package main

import (
	"os"
	"syscall"
)

func regularFileLinkCount(info os.FileInfo) (uint64, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return uint64(stat.Nlink), true
}
