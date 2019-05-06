// Copyright 2009 The vogo Authors. All rights reserved.

package main

import (
	"fmt"
	"github.com/wongoo/gracego"
	"net/http"
	"time"
)

var (
	server *http.Server
)

func main() {
	http.HandleFunc("/hello", HelloHandler)
	server = &http.Server{}

	err := gracego.Start(server, "demo", ":8081", )
	if err != nil {
		fmt.Printf("failed to start server: %v\n", err)
	}
}

//HelloHandler handle hello request
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	_, _ = w.Write([]byte("world233333!!!!"))
}
