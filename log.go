// Copyright 2019 The vogo Authors. All rights reserved.

package gracego

import (
	"fmt"
	"log"
)

func graceLog(format string, args ...interface{}) {
	log.Println("GRAC", serverID, fmt.Sprintf(format, args...))
}
