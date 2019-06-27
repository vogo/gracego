// Copyright 2019 The vogo Authors. All rights reserved.

package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/vogo/gracego"
)

var (
	server     *http.Server
	listenAddr = ":8081"
)

func main() {
	http.HandleFunc("/hello", HelloHandler)
	http.HandleFunc("/sleep5s", SleepHandler)
	http.HandleFunc("/calculate5s", CalculateHandler)
	http.HandleFunc("/download.zip", DownloadHandler)
	http.HandleFunc("/upgrade", UpgradeHandler)

	server = &http.Server{}

	err := gracego.EnableWritePid("/tmp")
	if err != nil {
		fmt.Printf("write pid error: %v", err)
	}

	err = gracego.Serve(server, "echo", listenAddr)
	if err != nil {
		fmt.Printf("server error: %v", err)
	}
}

//HelloHandler handle hello request
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	response(w, 200, "world")
}

//SleepHandler handle sleep request
func SleepHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	response(w, 200, "world")
}

//CalculateHandler handle calculation request
func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	fiveSecondCalc()
	response(w, 200, "world")
}

// the calculation will cost about 5.89s for 2.3 GHz Intel Core i5
func fiveSecondCalc() {
	for i := 0; i < math.MaxInt16; i++ {
		for j := 0; j < 1<<13; j++ {
			math.Sin(float64(i))
			math.Cos(float64(i))
		}
	}
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
	response(w, 200, "success")
}

func responseError(w http.ResponseWriter, err error) {
	response(w, 500, err.Error())
}
func response(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(msg))
}
