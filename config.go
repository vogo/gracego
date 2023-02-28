// Copyright 2019 The vogo Authors. All rights reserved.
// author: wongoo

package gracego

var addrInUseWaitSecond = 5

func SetAddrInUseWaitSecond(seconds int) {
	addrInUseWaitSecond = seconds
}
