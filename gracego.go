// Copyright 2019 The vogo Authors. All rights reserved.

package gracego

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"time"
)

const (
	forkCommandArg = "-grace-forked"
)

var (
	listener      net.Listener
	server        GraceServer
	graceForkArgs []string
	serverDir     string
	serverBin     string
	serverCmdPath string
	serverName    string
	serverAddr    string
	serverForked  bool

	shutdownChan    = make(chan error, 1)
	serverID        = int(time.Now().Unix())
	shutdownTimeout = 10 * time.Second
)

//GraceServer serve net listener
type GraceServer interface {
	Serve(listener net.Listener) error
}

//GraceShutdowner support shutdown
type GraceShutdowner interface {
	Shutdown(ctx context.Context) error
}

//Shutdowner support shutdown
type Shutdowner interface {
	Shutdown() error
}

// GetServerID get server id
func GetServerID() int {
	return serverID
}

// SetShutdownTimeout set the server shutdown timeout duration
func SetShutdownTimeout(d time.Duration) {
	if d > 0 {
		shutdownTimeout = d
	}
}

//Serve serve grace server
func Serve(svr GraceServer, name, addr string) error {
	var err error
	serverCmdPath, err = os.Executable()
	if err != nil {
		return err
	}
	serverDir = filepath.Dir(serverCmdPath)

	serverAddr = addr
	serverName = name
	server = svr

	serverBin = os.Args[0]
	graceForkArgs = os.Args[1:]
	serverForked = false
	for _, arg := range graceForkArgs {
		if arg == forkCommandArg {
			serverForked = true
			break
		}
	}
	if !serverForked {
		graceForkArgs = append(graceForkArgs, forkCommandArg)
	}

	return serveServer()
}

//serveServer start grace server
func serveServer() error {
	var err error

	writePidFile()

	if serverForked {
		graceLog("listen in forked child at %s, pid %d", serverAddr, os.Getpid())

		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {
		graceLog("listen at %s, pid %d", serverAddr, os.Getpid())
		listener, err = net.Listen("tcp", serverAddr)

		// wait for address being released
		if err != nil && IsAddrUsedErr(err) {
			graceLog("wait %d seconds for error: %v", addrInUseWaitSecond, err)
			<-time.After(time.Second * time.Duration(addrInUseWaitSecond))
			listener, err = net.Listen("tcp", serverAddr)
		}
	}
	if err != nil {
		graceLog("listen failed: %v", err)
		return err
	}

	go func() {
		err = server.Serve(listener)
		if err != nil {
			graceLog("server.Serve end! %v", err)
		}
	}()

	handleSignal()
	graceLog("server end, pid %d", os.Getpid())
	return nil
}

func handleSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var sig os.Signal

	for {
		select {
		case sig = <-signalChan:
		case err := <-shutdownChan:
			if err != nil {
				graceLog("shutdown error: %v", err)
			}
			close(shutdownChan)
			return
		}

		graceLog("receive signal: %v", sig)

		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			signal.Stop(signalChan)
			shutdown()
			return
		case syscall.SIGHUP:
			restart()
			return
		}
	}
}

func shutdown() {
	if enableWritePid {
		_ = os.Remove(pidFilePath)
	}

	go func() {
		shutdownChan <- shutdownServer(server)
	}()

	select {
	case <-time.After(shutdownTimeout + time.Second):
		shutdownChan <- fmt.Errorf("shutdown timeout over %d seconds", shutdownTimeout/time.Second)
	case <-shutdownChan:
	}
}

func shutdownServer(s GraceServer) error {
	graceLog("start shutdown server %s", reflect.TypeOf(s))
	defer graceLog("finish shutdown server %s", reflect.TypeOf(s))
	switch st := s.(type) {
	case GraceShutdowner:
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return st.Shutdown(ctx)
	case Shutdowner:
		return st.Shutdown()
	default:
		return errors.New("server shutdown unsupported")
	}
}

func restart() {
	err := fork()
	if err != nil {
		graceLog("failed to restart! fork child process error: %v", err)
		return
	}
	shutdown()
}

func fork() error {
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		return fmt.Errorf("listener is not tcp listener")
	}

	listenFile, err := tcpListener.File()
	if err != nil {
		return err
	}

	graceLog("restart server %s: %s %s", serverName, serverBin, strings.Join(graceForkArgs, " "))
	cmd := exec.Command(serverBin, graceForkArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{listenFile}
	return cmd.Start()
}
