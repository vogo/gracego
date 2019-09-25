// Copyright 2019 The vogo Authors. All rights reserved.
// author: wongoo

package gracego

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func existFile(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	return false
}

func execCmd(cmdline string) ([]byte, error) {
	cmd := exec.Command("/bin/sh", "-c", cmdline)
	result, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return bytes.ReplaceAll(result, []byte{'\n'}, nil), nil
}

func getPidFromAddr(addr string) (int, error) {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return 0, errors.New("can't get port from address")
	}
	port := addr[idx+1:]

	result, err := execCmd(fmt.Sprintf("lsof -i:%s |grep LISTEN | tail -1 |awk '{print $2}'", port))
	if err != nil {
		return 0, fmt.Errorf("failed to get pid from port %s, error: %+v", port, err)
	}
	pid, err := strconv.Atoi(string(result))
	if err != nil {
		return 0, fmt.Errorf("failed to get pid from port %s, result: %s, error: %v", port, result, err)
	}
	return pid, nil
}

func checkSameBinProcess(pid int) (string, bool) {
	cmdline, err := getCommandline(pid)
	if err != nil {
		graceLog("can't get command line for pid %d, error: %v", pid, err)
		return "", false
	}
	if strings.Contains(cmdline, serverBin) {
		return cmdline, true
	}
	return cmdline, false
}

func getCommandline(pid int) (string, error) {
	linuxCmdlineFile := fmt.Sprintf("/proc/%d/cmdline", pid)
	if existFile(linuxCmdlineFile) {
		data, err := ioutil.ReadFile(linuxCmdlineFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	result, err := execCmd(fmt.Sprintf("ps -o 'command' -p %d |tail -1", pid))
	if err != nil {
		return "", err
	}
	return string(result), nil
}
