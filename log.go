// Copyright 2019 The vogo Authors. All rights reserved.

package gracego

import (
	"fmt"
	"log"
)

func info(format string, args ...interface{}) {
	log.Println(serverID, "-", fmt.Sprintf(format, args...))
}
