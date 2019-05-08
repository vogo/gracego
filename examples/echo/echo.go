// Copyright 2009 The vogo Authors. All rights reserved.

package main

import (
	"fmt"
	"github.com/vogo/gracego"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var (
	server     *http.Server
	listenAddr = ":8081"
)

func main() {
	http.HandleFunc("/hello", HelloHandler)
	http.HandleFunc("/download.zip", DownloadHandler)
	http.HandleFunc("/upgrade", UpgradeHandler)
	server = &http.Server{}

	err := gracego.Start(server, "echo", listenAddr)
	if err != nil {
		fmt.Printf("failed to start server: %v\n", err)
	}
}

//HelloHandler handle hello request
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("world"))
}

//DownloadHandler download the graceup server zip
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	path, err := os.Executable()
	if err != nil {
		responseError(w, err)
		return
	}

	dir := filepath.Dir(path)
	zipFilePath := fmt.Sprintf("%s%c%s", dir, os.PathSeparator, "echo.zip")
	file, err := os.OpenFile(zipFilePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		responseError(w, err)
		return
	}

	w.Header().Add("content-type", "application/octet-stream")
	_, err = io.Copy(w, file)
	if err != nil {
		responseError(w, err)
		return
	}
}

//UpgradeHandler restart server
func UpgradeHandler(w http.ResponseWriter, r *http.Request) {
	err := gracego.Upgrade("v2", "echo", "http://127.0.0.1"+listenAddr+"/download.zip")
	if err != nil {
		responseError(w, err)
	}
}

func responseError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	_, _ = w.Write([]byte(err.Error()))
}
