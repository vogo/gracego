// Copyright 2009 The vogo Authors. All rights reserved.

package gracego

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	enableWritePid = false
	pidFileDir     string
	pidFilePath    string
)

//EnableWritePid enable to write pid file
//dir - the directory where to write pid file
func EnableWritePid(dir string) error {
	if dir == "" {
		dir = os.TempDir()
	} else {
		if _, err := os.Stat(dir); err != nil {
			return err
		}
	}

	if !strings.HasSuffix(dir, string(os.PathSeparator)) {
		dir += string(os.PathSeparator)
	}

	pidFileDir = dir

	enableWritePid = true
	return nil
}

func writePidFile() {
	if !enableWritePid {
		return
	}

	if pidFilePath == "" {

		pidFilePath = fmt.Sprintf("%s%s.pid", pidFileDir, serverName)
		log.Printf("set pid file: %s\n", pidFilePath)
	}

	pidFile, err := os.OpenFile(pidFilePath, os.O_RDWR, 0660)
	if err != nil {
		pidFile, err = os.Create(pidFilePath)
		if err != nil {
			log.Printf("failed to create pid file %s, error: %v\n", pidFilePath, err)
			return
		}
	}
	defer pidFile.Close()

	pid := fmt.Sprint(os.Getpid())
	_, err = pidFile.WriteString(pid)
	if err != nil {
		log.Printf("failed to write pid file %s, error: %v\n", pidFilePath, err)
	}
}
