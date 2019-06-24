// Copyright 2019 The vogo Authors. All rights reserved.

package gracego

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	forkCommandArg = "-grace-forked"
	forkTimeout    = 20 * time.Second
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

	shutdownChan = make(chan error, 1)
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
		log.Printf("listen in forked child, pid %d", os.Getpid())

		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {
		log.Printf("listen at %s, pid %d", serverAddr, os.Getpid())
		listener, err = net.Listen("tcp", serverAddr)
	}
	if err != nil {
		log.Printf("listen failed: %v\n", err)
		return err
	}

	go func() {
		err = server.Serve(listener)
		if err != nil {
			log.Printf("server.Serve end! %v\n", err)
		}
	}()

	handleSignal()
	log.Printf("serve end, pid %d", os.Getpid())
	return nil
}

func handleSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var sig os.Signal

	for {
		select {
		case sig = <-signalChan:
			break
		case err := <-shutdownChan:
			if err != nil {
				log.Printf("shutdown error: %v", err)
			}
			return
		}

		log.Printf("receive signal: %v\n", sig)

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
	log.Println("start shutdown server")
	ctx, cancel := context.WithTimeout(context.Background(), forkTimeout)
	defer cancel()

	if enableWritePid {
		_ = os.Remove(pidFilePath)
	}

	shutdownChan <- shutdownServer(server, ctx)
}

func shutdownServer(s GraceServer, ctx context.Context) error {
	if st, ok := s.(GraceShutdowner); ok {
		return st.Shutdown(ctx)
	}
	if st, ok := s.(Shutdowner); ok {
		return st.Shutdown()
	}
	return errors.New("server shutdown unsupported")
}

func restart() {
	err := fork()
	if err != nil {
		log.Printf("failed to restart! fork child process error: %v\n", err)
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

	log.Printf("restart server %s: %s %s\n", serverName, serverBin, strings.Join(graceForkArgs, " "))
	cmd := exec.Command(serverBin, graceForkArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{listenFile}
	return cmd.Start()
}
