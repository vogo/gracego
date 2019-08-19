// Copyright 2019 The vogo Authors. All rights reserved.
// author: wongoo

package gracego

import (
	"net"
	"os"
	"syscall"
)

func IsAddrUsedErr(err error) bool {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	callErr, ok := opErr.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	return callErr.Err == syscall.EADDRINUSE
}
