// Copyright 2009 The vogo Authors. All rights reserved.

package gracego

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//Upgrade graceup server
func Upgrade(version, path, upgradeUrl string) error {
	if err := upgradeServerBin(version, path, upgradeUrl); err != nil {
		return err
	}
	return restart()
}

//upgradeServerBin graceup server bin file
func upgradeServerBin(version, path, upgradeUrl string) error {
	if server == nil || serverBin == "" {
		return errors.New("server not started")
	}

	versionDir := fmt.Sprintf("%s%c%s", serverDir, os.PathSeparator, version)
	err := os.Mkdir(versionDir, 0770)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		return err
	}

	u, err := url.Parse(upgradeUrl)
	if err != nil {
		return err
	}
	uri := u.RequestURI()
	index := strings.LastIndex(uri, "/")
	if index < 0 {
		return fmt.Errorf("invalid download url: %s", upgradeUrl)
	}

	fileName := uri[index+1:]
	if ! acceptFileSuffix(fileName) {
		return fmt.Errorf("invalid suffix for download url: %s", upgradeUrl)
	}

	upgradeBin := fmt.Sprintf("%s%c%s", versionDir, os.PathSeparator, path)
	_, err = os.Open(upgradeBin)
	if err == nil {
		return link(upgradeBin, serverBin)
	}

	downloadPath := fmt.Sprintf("%s%c%s", versionDir, os.PathSeparator, fileName)
	err = downloadFile(downloadPath, upgradeUrl)
	if err != nil {
		return err
	}

	err = unzip(downloadPath, versionDir)
	if err != nil {
		return err
	}

	return link(upgradeBin, serverBinPath)
}

func link(src string, dest string) error {
	_ = os.Remove(dest)
	return os.Link(src, dest)
}

func acceptFileSuffix(f string) bool {
	return strings.HasSuffix(f, ".jar") || strings.HasSuffix(f, ".zip")
}

// downloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filepath string, url string) error {
	_ = os.Remove(filepath)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		log.Printf("can't create file: %v", err)
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
