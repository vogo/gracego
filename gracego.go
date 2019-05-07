// Copyright 2009 The vogo Authors. All rights reserved.

package gracego

import (
	"context"
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
	pidFilePath   string
	serverDir     string
	serverBin     string
	serverBinPath string
	serverName    string
	serverAddr    string
	serverForked  bool
)

//GraceServer serve net listener
type GraceServer interface {
	Serve(listener net.Listener) error
	Shutdown(ctx context.Context) error
}

//StartServer start grace server
func Start(svr GraceServer, name, addr string) error {
	serverAddr = addr
	serverName = name
	server = svr

	serverBinPath, err := os.Executable()
	if err != nil {
		return err
	}
	log.Println(serverBinPath)
	serverDir = filepath.Dir(serverBinPath)

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

	return startServer()
}

//startServer start grace server
func startServer() error {
	var err error

	pidFilePath = fmt.Sprintf("%s%s.pid", os.TempDir(), serverName)
	log.Printf("pid file: %s\n", pidFilePath)
	updatePidFile()

	if serverForked {
		log.Println("listening in forked child ...")

		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {
		log.Println("listening ...")
		listener, err = net.Listen("tcp", serverAddr)
	}
	if err != nil {
		log.Printf("listening failed: %v\n", err)
		return err
	}

	go func() {
		err = server.Serve(listener)
		if err != nil {
			log.Printf("server.Serve end! %v\n", err)
			os.Exit(1)
		}
	}()

	handleSignal()
	return nil
}

func handleSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		sig := <-ch
		log.Printf("signal: %v\n", sig)

		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			log.Println("start shutdown server")
			signal.Stop(ch)
			_ = shutdown()
			return
		case syscall.SIGHUP:
			_ = restart()
			return
		}
	}
}

func shutdown() error {
	ctx, _ := context.WithTimeout(context.Background(), forkTimeout)
	_ = os.Remove(pidFilePath)
	return server.Shutdown(ctx)
}

func restart() error {
	ctx, _ := context.WithTimeout(context.Background(), forkTimeout)
	err := fork()
	if err != nil {
		log.Printf("graceful restart failed: %v\n", err)
		return err
	}
	return server.Shutdown(ctx)
}

func fork() error {
	tl, ok := listener.(*net.TCPListener)
	if !ok {
		return fmt.Errorf("listener is not tcp listener")
	}

	listenFile, err := tl.File()
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

func updatePidFile() {
	pid := fmt.Sprint(os.Getpid())
	log.Printf("pid: %s\n", pid)

	pidFile, err := os.OpenFile(pidFilePath, os.O_RDWR, 0660)
	if err != nil {
		pidFile, err = os.Create(pidFilePath)
		if err != nil {
			panic(err)
		}
	}
	defer pidFile.Close()

	_, err = pidFile.WriteString(pid)
	if err != nil {
		log.Printf("write pid error: %v\n", err)
	}
}
