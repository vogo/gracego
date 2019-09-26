// Copyright 2019 The vogo Authors. All rights reserved.
// author: wongoo

package gracego

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func getBorrowSockFile(pid int) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("gracego_borrow_%d.sock", pid))
}

func borrow(addr string) (*os.File, error) {
	pid, err := getPidFromAddr(addr)
	if err != nil {
		return nil, err
	}

	procInfo, ok := checkSameBinProcess(pid)
	if !ok {
		return nil, fmt.Errorf("addr %s hold by process: %s", addr, procInfo)
	}

	graceLog("borrow addr %s from process[%d]: %s", addr, pid, procInfo)

	return borrowFromPid(pid)
}

func borrowFromPid(pid int) (*os.File, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("can't find target process %d, error: %+v", pid, err)
	}

	err = proc.Signal(syscall.SIGUSR1)
	if err != nil {
		return nil, fmt.Errorf("failed to send signal to process %d, error: %+v", pid, err)
	}

	borrowSockFile := getBorrowSockFile(pid)

	ticker := time.NewTicker(time.Millisecond * 100)
	for i := 0; i < 20 && !existFile(borrowSockFile); i++ {
		<-ticker.C
	}
	ticker.Stop()

	borrowConn, err := net.Dial("unix", borrowSockFile)
	if err != nil {
		return nil, err
	}
	defer borrowConn.Close()

	go func() {
		// wait 2 seconds (timeout control)
		<-time.After(time.Second * 2)
		borrowConn.Close()
	}()

	sendFdConn := borrowConn.(*net.UnixConn)
	sockFile, err := sendFdConn.File()
	if err != nil {
		return nil, err
	}
	defer sockFile.Close()

	buf := make([]byte, syscall.CmsgSpace(4))
	if _, _, _, _, err = syscall.Recvmsg(int(sockFile.Fd()), nil, buf, 0); err != nil {
		return nil, err
	}

	var messages []syscall.SocketControlMessage
	messages, err = syscall.ParseSocketControlMessage(buf)
	if err != nil {
		return nil, err
	}

	fds, err := syscall.ParseUnixRights(&messages[0])
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fds[0]), ""), nil
}

func borrowSend() error {
	listenFile, err := listenFile()
	if err != nil {
		return err
	}

	borrowSockFile := getBorrowSockFile(os.Getpid())
	sockListener, err := net.Listen("unix", borrowSockFile)
	if err != nil {
		return err
	}

	go func() {
		// wait 2 seconds (timeout control)
		<-time.After(time.Second * 2)
		sockListener.Close()
	}()

	sockConn, err := sockListener.Accept()
	if err != nil {
		return err
	}
	defer sockConn.Close()

	conn := sockConn.(*net.UnixConn)
	sockFile, err := conn.File()
	if err != nil {
		return err
	}
	defer sockFile.Close()

	rights := syscall.UnixRights(int(listenFile.Fd()))
	err = syscall.Sendmsg(int(sockFile.Fd()), nil, rights, nil, 0)
	if err != nil {
		return err
	}

	time.Sleep(time.Second)
	return nil
}
